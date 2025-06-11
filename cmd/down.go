package cmd

import (
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/procedure"
	"github.com/vulcanshen-tpi/task-compose/utils"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ShutdownProcess struct {
	process *os.Process
	name    string
}

func (p *ShutdownProcess) kill(wg *sync.WaitGroup) {
	if err := p.process.Kill(); err != nil {
		if spinner, ok := procedure.TaskSpinner.GetSpinner(p.name); ok {
			spinner.ErrorWithMessagef("Error killing process: %s", err.Error())
		}
	} else {
		if spinner, ok := procedure.TaskSpinner.GetSpinner(p.name); ok {
			spinner.CompleteWithMessagef("Shutdown Completed PID: %d", p.process.Pid)
		}
	}
	wg.Done()
}

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Kill previous tasks processes",
	Long:  "Kill previous tasks processes with command: task-compose down",
	PreRun: func(cmd *cobra.Command, args []string) {
		app.DetachMode = true
		procedure.InitializeSpinnerAgent()
		procedure.StartSpinnerAgent()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		procedure.StopSpinnerAgent()
	},
	Run: func(cmd *cobra.Command, args []string) {

		if dir, err := os.Getwd(); err == nil {
			var pidFilePath = filepath.Join(dir, procedure.PidFile)
			pidFile, err := os.ReadFile(pidFilePath)
			if err != nil && !os.IsNotExist(err) {
				utils.SharedAppLogger.Error(err)
				return
			}
			var pids procedure.TaskProcessLog

			if err = yaml.Unmarshal(pidFile, &pids); err != nil {
				utils.SharedAppLogger.Error(err)
				return
			}

			var shutdownProcess []ShutdownProcess

			var waitGroup = &sync.WaitGroup{}
			waitGroup.Add(len(pids.Tasks))

			for _, task := range pids.Tasks {
				procedure.TaskSpinner.RegisterSpinner(task.Name, task.Name+"|", "Shutting down")

				process, err := os.FindProcess(task.Pid)

				if err != nil {
					if spinner, ok := procedure.TaskSpinner.GetSpinner(task.Name); ok {
						spinner.ErrorWithMessagef("Error finding process: %s", err.Error())
					}
					continue
				} else {
					shutdownProcess = append(shutdownProcess, ShutdownProcess{process: process, name: task.Name})
				}
			}

			for _, process := range shutdownProcess {
				go process.kill(waitGroup)
			}

			waitGroup.Wait()
		}
	},
}
