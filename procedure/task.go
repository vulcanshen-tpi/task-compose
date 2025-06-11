package procedure

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/config"
	"github.com/vulcanshen-tpi/task-compose/utils"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/oliveagle/jsonpath"
	"gopkg.in/yaml.v3"
)

type Task struct {
	Name        string
	BaseDir     string
	Envs        []string
	Executable  string
	Args        []string
	DependsOn   []*Task
	Healthcheck config.HealthCheckConfig
	process     *exec.Cmd
	Healthy     bool
	logger      *utils.AppLogger
	Terminated  bool
}

type TaskProcess struct {
	Name string `yaml:"name"`
	Pid  int    `yaml:"pid"`
}

type TaskProcessLog struct {
	Tasks []*TaskProcess `yaml:"tasks"`
}

const (
	PidFile                   = ".taskpid.yaml"
	healthCheckDefaultTimeout = 10 * time.Second
	healthCheckInterval       = 5 * time.Second
	healthCheckTries          = 3
	healthCheckStartDelay     = 1 * time.Second
)

var TaskProcesses = TaskProcessLog{}

func CreateTask(config config.TaskConfig) (*Task, error) {
	var task = Task{
		Name:        config.Name,
		BaseDir:     config.BaseDir,
		Envs:        config.Envs,
		Executable:  config.Executable,
		Args:        config.Args,
		Healthcheck: config.Healthcheck,
	}
	return &task, nil
}

func (t *Task) AppendDependencies(dependency *Task) {
	t.DependsOn = append(t.DependsOn, dependency)
}

func (t *Task) Start(wg *sync.WaitGroup) {
	TaskSpinner.RegisterSpinner(t.Name, t.Name+"|", "Waiting")
	t.logger = utils.NewAppLogger(t.Name, utils.Color.GetRandomColorCode())
	for {
		check, terminated := t.checkDependencies()
		if check {
			break
		}
		if terminated {
			t.Terminated = true
			t.terminate()
			return
		}
		time.Sleep(1 * time.Second)
	}

	if spinner, ok := TaskSpinner.GetSpinner(t.Name); ok {
		spinner.UpdateMessage("Launching")
	}

	t.runCommand()
	t.logTaskProcess()

	var interval = healthCheckInterval
	var tries = healthCheckTries
	var startDelay = healthCheckStartDelay

	if freq := t.Healthcheck.Frequency; freq != nil {
		if freq.Interval != "" {
			interval, _ = time.ParseDuration(freq.Interval)
		}
		if freq.Tries > 0 {
			tries = freq.Tries
		}

		if freq.Delay != "" {
			startDelay, _ = time.ParseDuration(freq.Delay)
		}
	}

	time.Sleep(startDelay)

	var ticker = time.NewTicker(interval)
	failures := 0
	defer ticker.Stop()

	for range ticker.C {
		if check := t.doHealthCheck(); check {
			t.Healthy = true
			break
		}
		failures++
		t.logger.Warn(fmt.Sprintf("Health check %d/%d try", failures, tries))
		if failures >= tries {
			t.Healthy = false
			break
		}
	}

	if t.Healthy {
		if t.isHealthCheckConfigured() {
			t.logger.Success(fmt.Sprintf("Healthy check %d/%d success", failures+1, tries))
		}
		if spinner, ok := TaskSpinner.GetSpinner(t.Name); ok {
			spinner.CompleteWithMessage("Done")
		}
	} else {
		if spinner, ok := TaskSpinner.GetSpinner(t.Name); ok {
			spinner.ErrorWithMessage("Healthcheck failed")
		}
		t.terminate()
		t.Terminated = true
	}

	if app.DetachMode {
		wg.Done()
	}

}

func (t *Task) checkDependencies() (bool, bool) {
	var check = true
	for _, dependency := range t.DependsOn {
		if dependency.Terminated {
			check = false
			return check, true
		}
		check = dependency.Healthy && check
	}
	return check, false
}

func (t *Task) isHealthCheckConfigured() bool {
	return !(t.Healthcheck.HTTP == nil && t.Healthcheck.Command == nil)
}

func (t *Task) doHealthCheck() bool {

	if !t.isHealthCheckConfigured() {
		return true
	}

	var timeout = healthCheckDefaultTimeout

	if t.Healthcheck.Frequency != nil && t.Healthcheck.Frequency.Timeout != "" {
		timeout, _ = time.ParseDuration(t.Healthcheck.Frequency.Timeout)
	}

	if t.Healthcheck.HTTP != nil {
		client := &http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(t.Healthcheck.HTTP.URL)
		if err != nil {
			return false
		}

		if resp.StatusCode < 200 && resp.StatusCode >= 300 {
			return false
		}

		if t.Healthcheck.HTTP.Expect != nil {
			if t.Healthcheck.HTTP.Expect.Json != nil {
				if t.Healthcheck.HTTP.Expect.Json.Jsonpath == "" {
					t.logger.Debug(fmt.Sprintf("jsonpath not set"))
					return false
				}

				var contentType = resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.logger.Debug(fmt.Sprintf("Health check contenttype: %s", contentType))
					return false
				}

				bodyBytes, err := io.ReadAll(resp.Body)
				_ = resp.Body.Close()

				if err != nil {
					t.logger.Debug(fmt.Sprintf("Health check body error: %v", err))
					return false
				}

				var jsonResp interface{}
				err = json.Unmarshal(bodyBytes, &jsonResp)
				if err != nil {
					return false
				}

				checkValue, err := jsonpath.JsonPathLookup(jsonResp, t.Healthcheck.HTTP.Expect.Json.Jsonpath)

				if err != nil {
					return false
				}

				if t.Healthcheck.HTTP.Expect.Json.Value == "" {
					return true
				} else {
					return checkValue == t.Healthcheck.HTTP.Expect.Json.Value
				}
			}

			if t.Healthcheck.HTTP.Expect.Plain != nil {

				if t.Healthcheck.HTTP.Expect.Plain.Contains == "" {
					return false
				}

				bodyBytes, err := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if err != nil {
					return false
				}
				bodyString := string(bodyBytes)

				return strings.Contains(bodyString, t.Healthcheck.HTTP.Expect.Plain.Contains)

			}
		}

	}

	if t.Healthcheck.Command != nil {
		if len(t.Healthcheck.Command.Scripts) > 0 {
			var cmd = t.Healthcheck.Command.Scripts[0]
			var args = t.Healthcheck.Command.Scripts[1:]
			var process = exec.Command(cmd, args...)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := process.Start(); err != nil {
				return false
			}
			done := make(chan error, 1)
			go func() {
				done <- process.Wait()
			}()
			select {
			case <-ctx.Done():
				if process.Process != nil {
					err := syscall.Kill(-process.Process.Pid, syscall.SIGKILL)
					if err != nil {
						return false
					}
				}
				return false
			case err := <-done:
				if err != nil {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true

}

func (t *Task) logTaskProcess() {
	var processLog = TaskProcess{
		Name: t.Name,
		Pid:  t.process.Process.Pid,
	}
	TaskProcesses.Tasks = append(TaskProcesses.Tasks, &processLog)
	data, err := yaml.Marshal(&TaskProcesses)
	if err != nil {
		log.Println("Error marshalling task processes:", t.Name, err)
		return
	}

	if dir, err := os.Getwd(); err == nil {
		var pidFile = filepath.Join(dir, PidFile)
		if err = os.WriteFile(pidFile, data, 0644); err != nil {
			log.Println("Error creating pidfile:", pidFile, err)
		}
	}
}

func (t *Task) runCommand() {
	t.process = exec.Command(t.Executable, t.Args...)
	//log.Println(utils.Convertor.ToJson(t))
	if t.BaseDir != "" {
		t.process.Dir = t.BaseDir
	}
	t.process.Env = t.Envs
	//t.process.Stderr = os.Stderr
	//t.process.Stdout = os.Stdout

	if !app.DetachMode {
		// front ground detach mode

		stdoutPipe, err := t.process.StdoutPipe()
		if err != nil {
			t.logger.Error(err)
		}
		stderrPipe, err := t.process.StderrPipe()
		if err != nil {
			t.logger.Error(err)
		}
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Text()
				t.logger.Info(line)
			}
			if err := stdoutPipe.Close(); err != nil {
				t.logger.Error(err)
				//logger.Printf("%s|%v\n", prefix, utils.Convertor.ToErrorColor(err.Error()))
			}
		}()
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				line := scanner.Text()
				t.logger.Error(fmt.Errorf(line))
			}
			if err := stdoutPipe.Close(); err != nil && err != io.EOF {
				t.logger.Error(err)
			}
		}()
	}

	if err := t.process.Start(); err != nil {
		if spinner, ok := TaskSpinner.GetSpinner(t.Name); ok {
			spinner.ErrorWithMessagef("Error starting command: %s", err)
		}
	}
}

func (t *Task) terminate() {
	if t.process != nil && t.process.Process != nil {
		if err := t.process.Process.Kill(); err != nil {
			if spinner, ok := TaskSpinner.GetSpinner(t.Name); ok {
				spinner.ErrorWithMessagef("Error killing process: %s", err.Error())
			}
		}
	}
}
