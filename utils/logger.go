package utils

import (
	"fmt"
	"github.com/vulcanshen-tpi/task-compose/app"
	"log"
	"os"
	"strings"
	"time"
)

type AppLogger struct {
	prefix  string
	color   int
	console *log.Logger
	file    *log.Logger
}

var SharedAppLogger = AppLogger{
	prefix:  "task-compose",
	color:   178,
	console: log.New(os.Stdout, "", 0),
}

func makeDir() {
	err := os.MkdirAll("logs", 0755) // 0755 是目錄的權限
	if err != nil {
		SharedAppLogger.console.Fatal(err)
		return
	}
}

func NewAppLogger(prefix string, color int) *AppLogger {
	consoleLogger := log.New(os.Stdout, "", 0)
	makeDir()
	today := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("logs/%s-%s.log", prefix, today)

	var appLogger = &AppLogger{
		prefix:  prefix,
		color:   color,
		console: consoleLogger,
	}

	if file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fileLogger := log.New(file, "", 0)
		appLogger.file = fileLogger
	}
	return appLogger
}

func (apl *AppLogger) getPrefix() string {
	return Convertor.Colored(fmt.Sprintf("%s|", apl.prefix), apl.color)
}

func (apl *AppLogger) Info(message ...string) {
	var msg = strings.Join(message, " ")
	if !app.DetachMode {
		apl.console.Printf("%s%s", apl.getPrefix(), msg)
	}
	if apl.file != nil {
		apl.file.Printf("%s", msg)
	}
}

func (apl *AppLogger) Success(message ...string) {
	var msg = strings.Join(message, " ")
	if !app.DetachMode {
		apl.console.Printf("%s%s", apl.getPrefix(), Convertor.ToSuccessColor(msg))
	}
	if apl.file != nil {
		apl.file.Printf("%s", msg)
	}
}

func (apl *AppLogger) Log(message ...string) {
	var msg = strings.Join(message, " ")
	if !app.DetachMode {
		apl.console.Printf("%s%s", apl.getPrefix(), Convertor.ToLogColor(msg))
	}
}

func (apl *AppLogger) Warn(message ...string) {
	var msg = strings.Join(message, " ")
	if !app.DetachMode {
		apl.console.Printf("%s%s", apl.getPrefix(), Convertor.ToWarningColor(msg))
	}
	if apl.file != nil {
		apl.file.Printf("%s", msg)
	}
}

func (apl *AppLogger) Error(err error) {
	if !app.DetachMode {
		apl.console.Printf("%s%s", apl.getPrefix(), Convertor.ToErrorColor(err.Error()))
	}

	if apl.file != nil {
		apl.file.Printf("%s", err.Error())
	}
}

func (apl *AppLogger) Debug(message ...string) {
	if app.DebugMode {
		var msg = strings.Join(message, " ")
		apl.console.Printf("%s%s", apl.getPrefix(), Convertor.ToDebugColor(msg))
		if apl.file != nil {
			apl.file.Printf("%s", msg)
		}
	}
}

func (apl *AppLogger) Fatal(err error) {
	if apl.file != nil {
		apl.file.Printf("%s", err.Error())
	}
	apl.console.Fatalf("%s%s", apl.getPrefix(), Convertor.ToErrorColor(err.Error()))
}
