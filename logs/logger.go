package logs

import (
	"fmt"
	"io"
	"strings"
)

type Logger struct {
	out io.Writer
}

func NewLogger(out io.Writer) *Logger {
	return &Logger{
		out: out,
	}
}

func (logger *Logger) Printfln(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(logger.out, format+"\n", a...)
}

func (logger *Logger) PrintStatusErrors(errors []string, firstLineOnly bool) {
	for _, err := range errors {
		message := err
		if firstLineOnly {
			message = strings.Split(message, "\n")[0]
		}
		logger.Printfln("  ► %s", message)
	}
}
