package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/config"
	"github.com/vulcanshen-tpi/task-compose/utils"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate minimal task-compose.yaml file",
	Long:  "Generate minimal task-compose.yaml in the current directory with a basic 'echo' task.",
	Run: func(cmd *cobra.Command, args []string) {

		if dir, err := os.Getwd(); err == nil {
			message := fmt.Sprintf("dir: %s", dir)
			utils.SharedAppLogger.Info(message)
		}

		//minimalConfig := config.LauncherConfig{
		//	Tasks: []config.TaskConfig{
		//		{
		//			Name:       "echo",
		//			Executable: "echo",
		//			Args:       []string{"hello", "world"},
		//		},
		//	},
		//}
		minimalConfig := map[string]any{
			"tasks": []map[string]any{ // map 的值也可以是另一個 map
				{
					"name":       "echo",
					"executable": "echo",
					"args":       []string{"hello", "world"},
				},
			},
		}

		currentDir, err := os.Getwd()
		if err != nil {
			utils.SharedAppLogger.Fatal(err)
		}

		var outputFile = config.GetDefaultFileNameWithExtension()

		if app.OutputFileName != "" {
			outputFile = app.OutputFileName
		}

		outputFile = filepath.Join(currentDir, outputFile)

		if _, err := os.Stat(outputFile); err == nil {
			utils.SharedAppLogger.Warn(fmt.Sprintf("'task-compose.yaml' already exists in '%s'. Aborting to prevent overwrite.\n", currentDir))
			utils.SharedAppLogger.Warn("If you wish to overwrite, please delete the existing file first.")
			return
		} else if !os.IsNotExist(err) {
			// Handle other potential errors from os.Stat
			utils.SharedAppLogger.Fatal(err)
		}

		yamlContent, err := yaml.Marshal(minimalConfig)
		if err != nil {
			utils.SharedAppLogger.Fatal(err)
		}

		err = os.WriteFile(outputFile, yamlContent, 0644)
		if err != nil {
			utils.SharedAppLogger.Fatal(err)
		}
		utils.SharedAppLogger.Info("Successfully generated task-compose.yaml", "path", outputFile)
	},
}

func init() {
	InitCmd.PersistentFlags().StringVarP(&app.OutputFileName, "output", "o", "", "Specify the generated configuration file name")
}
