package log

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Error(message string) {
	fmt.Printf("[%v] %v\n", color.RedString("FAILED"), message)
	os.Exit(-1)
}

func Info(message string) {
	fmt.Printf("[%v] %v\n", color.GreenString("SUCCESS"), message)
}
