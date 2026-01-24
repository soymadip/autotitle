package logger

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var (
	// Icons
	IconSuccess = "✔"
	IconInfo    = "ℹ"
	IconWarn    = "⚠"
	IconError   = "✘"
	IconFatal   = "☠"

	// Colors
	ColorSuccess = color.New(color.FgGreen).SprintFunc()
	ColorInfo    = color.New(color.FgCyan).SprintFunc()
	ColorWarn    = color.New(color.FgYellow).SprintFunc()
	ColorError   = color.New(color.FgRed).SprintFunc()
	ColorFatal   = color.New(color.FgRed, color.Bold).SprintFunc()
	ColorGray    = color.New(color.FgHiBlack).SprintFunc()

	// State
	isQuiet   bool
	isNoColor bool
)

// Init configures the logger
func Init(quiet, noColor bool) {
	isQuiet = quiet
	isNoColor = noColor || os.Getenv("NO_COLOR") != ""

	if !isNoColor && !isatty.IsTerminal(os.Stdout.Fd()) {
		isNoColor = true
	}

	if isNoColor {
		color.NoColor = true
	}
}

// Success prints a success message (green checkmark)
func Success(format string, a ...interface{}) {
	if isQuiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s\n", ColorSuccess(IconSuccess), msg)
}

// Info prints an info message (cyan i)
func Info(format string, a ...interface{}) {
	if isQuiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s\n", ColorInfo(IconInfo), msg)
}

// Warn prints a warning message (yellow warning sign)
func Warn(format string, a ...interface{}) {
	if isQuiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s %s\n", ColorWarn(IconWarn), msg)
}

// Error prints an error message to stderr (red x)
func Error(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s %s\n", ColorError(IconError), msg)
}

// Fatal prints a fatal error message and exits (red skull)
func Fatal(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s %s\n", ColorFatal(IconFatal), msg)
	os.Exit(1)
}

// Step prints a step description (gray text)
func Step(format string, a ...interface{}) {
	if isQuiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("  %s\n", ColorGray(msg))
}

// IsTerminal checks if stdout is a terminal
func IsTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
