package io

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Logger struct {
	out          io.Writer
	enableColors bool
	debug        bool
}

func NewLogger(out io.Writer, enableColors bool, debug bool) *Logger {
	return &Logger{
		out:          out,
		enableColors: enableColors,
		debug:        debug,
	}
}

func NewTestLogger() (*Logger, *bytes.Buffer) {
	out := new(bytes.Buffer)
	return NewLogger(out, false, false), out
}

func (logger *Logger) Debug(a ...interface{}) (n int, err error) {
	if logger.debug {
		logger.Write([]byte("DEBUG: "))
		return logger.Println(a...)
	}
	return 0, nil
}

func (logger *Logger) Debugf(format string, a ...interface{}) (n int, err error) {
	if logger.debug {
		logger.Write([]byte("DEBUG: "))
		return logger.Printf(format, a...)
	}
	return 0, nil
}

func (logger *Logger) Write(p []byte) (n int, err error) {
	return logger.out.Write(p)
}

func (logger *Logger) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(logger.out, a...)
}

func (logger *Logger) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(logger.out, format, a...)
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

// PrettyPrintDiff - prints a diff using syntax highlighting if bat is available
func (logger *Logger) PrettyPrintDiff(patch string) {
	logger.prettyPrint(patch, "diff")
}

func (logger *Logger) prettyPrint(code string, language string) {
	if _, err := exec.LookPath("bat"); err != nil {
		logger.Println(code)
		return
	}

	color := "never"
	if logger.enableColors {
		color = "always"
	}

	cmd := exec.Command("bat", "--language", language, "--color", color, "--style", "plain")
	cmd.Stdin = strings.NewReader(code)

	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Println("error running cmd")
		logger.Println(stderr.String())
		logger.Println(err)
		logger.Println(code)
		return
	}

	logger.Println(out.String())
}
