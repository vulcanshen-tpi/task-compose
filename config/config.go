package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/utils"
	"os"
	"path/filepath"
)

type HttpCheckExpectJson struct {
	Value    string `mapstructure:"value"`
	Jsonpath string `mapstructure:"jsonpath"`
}

type HttpCheckExpectPlain struct {
	Contains string `mapstructure:"contains"`
}

type HttpCheckExpect struct {
	Json  *HttpCheckExpectJson  `mapstructure:"json"`
	Plain *HttpCheckExpectPlain `mapstructure:"plain"`
}

// HTTPCheck 定義了 HTTP 健康檢查的配置
type HTTPCheck struct {
	URL    string           `mapstructure:"url"`
	Expect *HttpCheckExpect `mapstructure:"expect"`
}

type CommandCheck struct {
	Scripts []string `mapstructure:"scripts"`
}

type CheckFrequency struct {
	Interval string `mapstructure:"interval"`
	Timeout  string `mapstructure:"timeout"`
	Tries    int    `mapstructure:"tries"`
	Delay    string `mapstructure:"delay"`
}

// HealthCheckConfig 定義了應用程式的健康檢查配置
type HealthCheckConfig struct {
	HTTP      *HTTPCheck      `mapstructure:"http"`
	Command   *CommandCheck   `mapstructure:"command"`
	Frequency *CheckFrequency `mapstructure:"frequency"`
}

// TaskConfig 定義了單個應用程式的配置
type TaskConfig struct {
	Name        string            `mapstructure:"name"`
	BaseDir     string            `mapstructure:"base_dir"`
	Envs        []string          `mapstructure:"envs"`
	Executable  string            `mapstructure:"executable"`
	Args        []string          `mapstructure:"args"`
	Healthcheck HealthCheckConfig `mapstructure:"healthcheck"`
	DependsOn   []string          `mapstructure:"depends_on"`
}

// LauncherConfig 定義了整個 task-compose.yaml 的根配置
type LauncherConfig struct {
	Tasks []TaskConfig `mapstructure:"tasks"`
}

const defaultFileName = "task-compose"

var AppConfig LauncherConfig

var AppTasksConfig map[string]TaskConfig

func InitConfig() {
	//logger := log.New(os.Stdout, "", 0)
	viper.SetEnvPrefix("CMD_COMPOSE")
	viper.AutomaticEnv()

	if app.TasksComposeFile == "" {
		dir, err := os.Getwd()
		if err == nil {
			app.TasksComposeFile = filepath.Join(dir, fmt.Sprintf("%s.yaml", defaultFileName))
		}
	}

	if absFilePath, err := filepath.Abs(app.TasksComposeFile); err == nil {
		app.TasksComposeFile = absFilePath
	}

	viper.SetConfigFile(app.TasksComposeFile)

	if err := viper.ReadInConfig(); err == nil {
		utils.SharedAppLogger.Info(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
		//logger.Printf("%s|%s", AppLogPrefix, fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	} else {
		utils.SharedAppLogger.Fatal(err)
		//logger.Fatalf("%s|%s", AppLogPrefix, utils.Convertor.ToErrorColor(err.Error()))
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		utils.SharedAppLogger.Fatal(err)
	}
}
