package styles

import (
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora"
)

var flagsRegex, argNameRegex *regexp.Regexp

// StyleError - highlight errors in red
func StyleError(s string) string {
	return aurora.Red(s).Bold().String()
}

// StyleEnumOption - styles an enum option
func StyleEnumOption(s string) string {
	return aurora.BgGreen(s).Black().Bold().String()
}

// StyleEnumOptions - enumerate valid enum options
func StyleEnumOptions(opts []string) string {
	var styledOptions []string
	for _, opt := range opts {
		styledOptions = append(styledOptions, StyleEnumOption(opt))
	}
	return strings.Join(styledOptions, ", ")
}

// StyleHeading - style headings
func StyleHeading(s string) aurora.Value {
	return aurora.Bold(s)
}

// StyleCommand - style commands
func StyleCommand(s string) aurora.Value {
	return aurora.Green(s).Bold()
}

// StyleOK - style ok logs
func StyleOK(s string) aurora.Value {
	return aurora.Green(s).Bold()
}

// StyleCommandUsage - style command usage examples
func StyleCommandUsage(s string) string {
	styledCommand := StyleCommand(s).String()
	styledCommand = strings.ReplaceAll(styledCommand, "[flags]", StyleOptions("[flags]").String())
	styledCommand = argNameRegex.ReplaceAllStringFunc(styledCommand, func(argName string) string {
		return StyleOptions(argName).String()
	})
	return styledCommand
}

// StyleOptions - style command options and flags
func StyleOptions(s string) aurora.Value {
	return aurora.Yellow(s).Bold()
}

// StyleFlags - style flag usage examples
func StyleFlags(s string) string {
	var styledUsages []string
	for _, flagUsage := range strings.Split(s, "\n") {
		styledUsage := flagsRegex.ReplaceAllStringFunc(flagUsage, func(flag string) string {
			return StyleOptions(flag).String()
		})
		styledUsages = append(styledUsages, styledUsage)
	}
	return strings.Join(styledUsages, "\n")
}

func init() {
	// matches either of:
	//   -h, --help
	//       --help
	flagsRegex = regexp.MustCompile(`^\s+-\S,\s+--\S+|^\s+--\S+`)

	// matches: <my-arg>
	argNameRegex = regexp.MustCompile(`<\S+>`)
}
