package config

import (
	"fmt"
)

type TaskConfigCheck int

const (
	Unvisited TaskConfigCheck = iota // 0: 未訪問
	Visiting                         // 1: 正在訪問 (當前路徑中)
	Visited                          // 2: 已訪問 (已完成遍歷)
)

func (lc *LauncherConfig) Validate() error {
	configs := lc.Tasks
	tasks := make(map[string]TaskConfig)
	taskChecks := make(map[string]TaskConfigCheck)
	for _, config := range configs {
		if _, exists := tasks[config.Name]; exists {
			// 如果 config.Name 已經存在於 map 中，則表示有重複名稱
			return fmt.Errorf("duplicate task name found: %s", config.Name)
		}
		tasks[config.Name] = config
		taskChecks[config.Name] = Unvisited
	}

	// check for missing dependencies
	for name, task := range tasks {
		if len(task.DependsOn) > 0 {
			for _, dependency := range task.DependsOn {
				if _, ok := tasks[dependency]; !ok {
					return fmt.Errorf("task %s missing dependency %s", name, dependency)
				}
			}
		}
	}

	// check for circular dependencies
	for name := range tasks {
		if taskChecks[name] == Unvisited {
			// 對每個未訪問的任務啟動 DFS
			if err := checkCycleDFS(name, tasks, taskChecks); err != nil {
				return err // 發現環狀依賴，立即返回錯誤
			}
		}
	}

	for _, task := range configs {
		for _, dependencyName := range task.DependsOn {
			depConfig, ok := tasks[dependencyName]
			if !ok {
				// 這個錯誤應該在 ValidateConfig 中被捕獲，這裡是二次防禦
				return fmt.Errorf("internal error: dependency %s for task %s not found", dependencyName, task.Name)
			}

			// 檢查依賴任務是否有設定 healthcheck
			// healthcheck.HTTP.URL 和 healthcheck.Command.Scripts 都是 string，如果沒有設定，Go 的零值是 "" 或 nil slice
			// 判斷方式：如果 HTTP.URL 非空，或者 Command.Scripts 非空且長度大於 0
			hasHTTPHealthcheck := depConfig.Healthcheck.HTTP != nil && depConfig.Healthcheck.HTTP.URL != ""
			hasCommandHealthcheck := depConfig.Healthcheck.Command != nil && depConfig.Healthcheck.Command.Scripts != nil && len(depConfig.Healthcheck.Command.Scripts) > 0

			if !hasHTTPHealthcheck && !hasCommandHealthcheck {
				return fmt.Errorf("task %s depends on %s, but %s has no healthcheck configured. All depended-on tasks must have a healthcheck",
					task.Name, dependencyName, dependencyName)
			}
		}
	}

	AppTasksConfig = tasks
	return nil
}

func checkCycleDFS(taskName string, tasks map[string]TaskConfig, taskStates map[string]TaskConfigCheck) error {

	taskStates[taskName] = Visiting

	for _, dependencyName := range tasks[taskName].DependsOn {
		state := taskStates[dependencyName]
		if state == Visiting {
			return fmt.Errorf("circular dependency detected: %s -> %s", taskName, dependencyName)
		}
		if state == Unvisited {
			if err := checkCycleDFS(dependencyName, tasks, taskStates); err != nil {
				return err
			}
		}
	}

	taskStates[taskName] = Visited
	return nil
}
