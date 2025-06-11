package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/utils"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version number and build details of task-compose",
	Run: func(cmd *cobra.Command, args []string) {
		var info = make(map[string]string)
		info["version"] = app.Version
		info["build_date"] = app.BuildDate
		info["commit_hash"] = app.CommitHash
		info["execution_mode"] = app.ExecutionMode
		if jsonData, err := json.MarshalIndent(info, "", "  "); err == nil {
			utils.SharedAppLogger.Info(string(jsonData))
		} else {
			utils.SharedAppLogger.Fatal(err)
		}
	},
}
