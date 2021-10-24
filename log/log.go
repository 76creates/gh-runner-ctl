package log

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
)

// TODO: child logger with additional string? zerolog maybe?

// Debug prints the message in a debug format
func Debug(content string) {
	if viper.GetBool("debug") {
		log("debug", os.Stdout, content)
	}
}

// DebugF prints a formatted message in a debug format
func DebugF(format string, a ...interface{}) {
	Debug(fmt.Sprintf(format, a...))
}

// Warning prints the message in a warning format
func Warning(content string) {
	log("warning", os.Stdout, content)
}

// WarningF prints a formatted message in a warning format
func WarningF(format string, a ...interface{}) {
	Warning(fmt.Sprintf(format, a...))
}

// Error prints the message in a error format
func Error(content string) {
	log("error", os.Stderr, content)
}

// ErrorF prints a formatted message in a error format
func ErrorF(format string, a ...interface{}) {
	Error(fmt.Sprintf(format, a...))
}

func log(lvl string, out io.Writer, content string) {
	var msg string
	if viper.GetBool("github.actions") {
		msg = fmt.Sprintf("::%s::%s", lvl, content)
	} else {
		var color string
		switch lvl {
		case "debug": color = "\033[38;5;44m"
		case "warning": color = "\033[38;5;184m"
		case "error": color = "\033[38;5;160m"
		}
		msg = fmt.Sprintf("%s[%s] %s\033[0m\n", color, strings.ToUpper(lvl), content)
	}

	// TODO: handle error somehow
	fmt.Fprint(out, msg)
}
