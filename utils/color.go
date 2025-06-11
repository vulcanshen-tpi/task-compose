package utils

import "math/rand"

type color struct{}

var Color = color{}

const ErrorColor = 9
const AppColor = 178
const WarningColor = 220
const DebugColor = 39
const SuccessColor = 82

func (c *color) GetRandomColorCode() int {
	var colorCode = rand.Intn(256)
	for colorCode == ErrorColor ||
		colorCode == AppColor ||
		colorCode == WarningColor ||
		colorCode == DebugColor {
		colorCode = rand.Intn(256)
	}
	return colorCode
}
