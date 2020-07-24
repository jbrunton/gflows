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
}

func NewLogger(out io.Writer, enableColors bool) *Logger {
	return &Logger{
		out:          out,
		enableColors: enableColors,
	}
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

// PrettyPrintDiff - highlights the given patch if pygmentize is available
func (logger *Logger) PrettyPrintDiff(patch string) {
	logger.prettyPrint(patch, "diff")
}

// Inspired by https://github.com/pksunkara/pygments
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
