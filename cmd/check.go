package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/config"
	"github.com/vulcanshen-tpi/task-compose/utils"
	"os"

	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Confirm the correctness of the YAML content",
	Long:  "Confirm the correctness of the YAML content: task-compose check",
	Run: func(cmd *cobra.Command, args []string) {

		if dir, err := os.Getwd(); err == nil {
			message := fmt.Sprintf("dir: %s", dir)
			utils.SharedAppLogger.Info(message)
		}

		if err := CheckConfig(); err != nil {
			utils.SharedAppLogger.Fatal(err)
		}

		if len(config.AppConfig.Tasks) > 0 {

			if app.ShowDetail {
				utils.SharedAppLogger.Info(fmt.Sprintf("Found %d tasks in config.\n", len(config.AppConfig.Tasks)))
				jsonData, err := json.MarshalIndent(config.AppConfig.Tasks, "", "  ")
				if err != nil {
					utils.SharedAppLogger.Fatal(err)
				}
				utils.SharedAppLogger.Info(string(jsonData))
			}

			utils.SharedAppLogger.Success("configuration check success")

		} else {
			utils.SharedAppLogger.Warn("No applications defined in the configuration.")
		}

	},
}

func init() {
	CheckCmd.PersistentFlags().BoolVar(&app.ShowDetail, "detail", false, "Show configuration details")
}

func CheckConfig() error {

	config.InitConfig()

	if err := config.AppConfig.Validate(); err != nil {
		return fmt.Errorf("Error validating config: %v\n", err)
	}

	//if len(config.AppConfig.Tasks) > 0 {
	//	_, err := config.AppConfig.GetLayeredStartupOrder()
	//	if err != nil {
	//		return fmt.Errorf("Error getting layered startup order: %v\n", err)
	//	}
	//}

	return nil
}
