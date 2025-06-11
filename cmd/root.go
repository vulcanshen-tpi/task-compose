package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/utils"
	"os"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   "task-compose",
	Short: "A versatile application launcher with flexible configuration options",
	Long: `task-compose is a versatile command-line tool designed to simplify the management of complex, multi-process applications or task sequences.
	
It allows you to define a series of commands (tasks) and their operational parameters within a single, human-readable YAML configuration file.
This includes specifying the executable, arguments, working directories, environment variables, and crucial inter-task dependencies.

A core feature of task-compose is its robust health check mechanism. Each task can be configured with HTTP-based health checks (including JSON response validation via JSONPath) or command-line script-based health checks.
Tasks with dependencies will only start once their prerequisites are deemed healthy, ensuring a stable and reliable startup order for your services.

Whether you're spinning up a local development environment with multiple microservices, orchestrating integration tests, or automating complex workflows,
task-compose provides a declarative and efficient way to manage your system's components.
	`,
}

func Execute() {
	rootCmd.AddCommand(CheckCmd)
	rootCmd.AddCommand(UpCmd)
	rootCmd.AddCommand(DownCmd)
	rootCmd.AddCommand(VersionCmd)

	//utils.SharedAppLogger.Info(fmt.Sprintf("args:%v, execution mode: %v", os.Args, app.ExecutionMode))
	if len(os.Args) == 1 && app.ExecutionMode == "GUI" {
		rootCmd.SetArgs([]string{UpCmd.Use})
	}

	if err := rootCmd.Execute(); err != nil {
		utils.SharedAppLogger.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(SwitchLaunchDir)
	rootCmd.PersistentFlags().StringVarP(&app.TasksComposeFile, "configfile", "f", "", "Specify the path to the configuration file, default is 'task-compose.yaml'")
	rootCmd.PersistentFlags().BoolVar(&app.DebugMode, "debug", false, "Launch in tasks in debug mode")
	if app.ExecutionMode == "" {
		app.ExecutionMode = "GUI"
	}
}

func SwitchLaunchDir() {

	if app.ExecutionMode == "GUI" {
		// gui protocol executable
		path, err := os.Executable()
		if err != nil {
			utils.SharedAppLogger.Fatal(err)
		}
		if err = os.Chdir(filepath.Dir(path)); err != nil {
			utils.SharedAppLogger.Fatal(err)
		}
	} else {
		// cli mode
		dir, err := os.Getwd()
		if err != nil {
			utils.SharedAppLogger.Fatal(err)
		}
		if err = os.Chdir(dir); err != nil {
			utils.SharedAppLogger.Fatal(err)
		}
	}

	if dir, err := os.Getwd(); err == nil {
		message := fmt.Sprintf("launch dir %s", dir)
		utils.SharedAppLogger.Info(message)
	}
}
