package logs

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
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

func (logger *Logger) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(logger.out, a...)
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
	logger.prettyPrint(patch, "-ldiff")
}

// Inspired by https://github.com/pksunkara/pygments
func (logger *Logger) prettyPrint(code string, lexer string) {
	if _, err := exec.LookPath("pygmentize"); err != nil {
		logger.Println(code)
		return
	}

	cmd := exec.Command("pygmentize", "-fterminal256", lexer, "-O style=monokai")
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
