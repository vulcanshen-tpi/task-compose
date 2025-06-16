/*
Copyright © 2025 Vulcan Shen vulcan.shen@tpisoftware.com
*/
package main

import (
	"fmt"
	"github.com/vulcanshen-tpi/task-compose/app"
	"github.com/vulcanshen-tpi/task-compose/cmd"
	"github.com/vulcanshen-tpi/task-compose/utils"
	"os"
	"path/filepath"
)

func main() {

	if len(os.Args) == 1 && app.Portable == "true" {
		// gui mode (portable)
		path, err := os.Executable()
		if err != nil {
			utils.SharedAppLogger.Fatal(err)
		}
		if err = os.Chdir(filepath.Dir(path)); err != nil {
			utils.SharedAppLogger.Fatal(err)
		}

		cmd.RootCmd.AddCommand(cmd.UpCmd)
		cmd.RootCmd.SetArgs([]string{cmd.UpCmd.Use})

		if err := cmd.RootCmd.Execute(); err != nil {
			utils.SharedAppLogger.Fatal(fmt.Errorf("error executing UpCmd in GUI mode: %w", err))
			os.Exit(1)
		}

		os.Exit(0)
		return
	}

	// cli mode
	dir, err := os.Getwd()
	if err != nil {
		utils.SharedAppLogger.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		utils.SharedAppLogger.Fatal(err)
	}
	cmd.Execute()
}
