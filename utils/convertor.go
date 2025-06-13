package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

type convertor struct{}

var Convertor = convertor{}

func (convertor *convertor) ToJson(any any) string {
	jsonData, err := json.Marshal(any)
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	return string(jsonData)
}

func (convertor *convertor) ToLogColor(message string) string {
	return convertor.Colored(message, LogColor)
}

func (convertor *convertor) ToErrorColor(message string) string {
	return convertor.Colored(message, ErrorColor)
}

func (convertor *convertor) ToWarningColor(message string) string {
	return convertor.Colored(message, WarningColor)
}

func (convertor *convertor) ToSuccessColor(message string) string {
	return convertor.Colored(message, SuccessColor)
}

func (convertor *convertor) ToDebugColor(message string) string {
	return convertor.Colored(message, DebugColor)
}

func (convertor *convertor) Colored(message string, colorCode int) string {
	color := fmt.Sprintf("\033[38;5;%dm", colorCode)
	return fmt.Sprintf("%s%s%s", color, message, "\033[0m")
}
