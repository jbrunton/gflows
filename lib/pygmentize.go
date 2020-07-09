package lib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// PrettyPrintDiff - highlights the given patch if pygmentize is available
func PrettyPrintDiff(patch string) {
	prettyPrint(patch, "-ldiff")
}

// Inspired by https://github.com/pksunkara/pygments
func prettyPrint(code string, lexer string) {
	if _, err := exec.LookPath("pygmentize"); err != nil {
		fmt.Println(code)
		return
	}

	cmd := exec.Command("pygmentize", "-fterminal256", lexer, "-O style=monokai")
	cmd.Stdin = strings.NewReader(code)

	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("error running cmd")
		fmt.Println(stderr.String())
		fmt.Println(err)
		fmt.Println(code)
		return
	}

	fmt.Println(out.String())
}
