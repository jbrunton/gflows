package styles

import (
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora"
)

var flagsRegex, argNameRegex *regexp.Regexp

// Styles - wraps Aurora instance with some convenience functions
type Styles struct {
	au aurora.Aurora
}

// NewStyles - construct a new Styles instance
func NewStyles(enableColors bool) *Styles {
	au := aurora.NewAurora(enableColors)
	return &Styles{au}
}

// Bold - apply bold style
func (styles *Styles) Bold(s string) aurora.Value {
	return styles.au.Bold(s)
}

// StyleError - highlight errors in red
func (styles *Styles) StyleError(s string) string {
	return styles.au.Red(s).Bold().String()
}

// StyleEnumOption - styles an enum option
func (styles *Styles) StyleEnumOption(s string) string {
	return styles.au.BgGreen(s).Black().Bold().String()
}

// StyleEnumOptions - enumerate valid enum options
func (styles *Styles) StyleEnumOptions(opts []string) string {
	var styledOptions []string
	for _, opt := range opts {
		styledOptions = append(styledOptions, styles.StyleEnumOption(opt))
	}
	return strings.Join(styledOptions, ", ")
}

// StyleHeading - style headings
func (styles *Styles) StyleHeading(s string) aurora.Value {
	return styles.au.Bold(s)
}

// StyleCommand - style commands
func (styles *Styles) StyleCommand(s string) aurora.Value {
	return styles.au.Green(s).Bold()
}

// StyleOK - style ok logs
func (styles *Styles) StyleOK(s string) aurora.Value {
	return styles.au.Green(s).Bold()
}

// StyleWarning - style warning logs
func (styles *Styles) StyleWarning(s string) aurora.Value {
	return styles.au.Yellow(s).Bold()
}

// StyleOptions - style command options and flags
func (styles *Styles) StyleOptions(s string) aurora.Value {
	return styles.au.Yellow(s).Bold()
}

// StyleFlags - style flag usage examples
func (styles *Styles) StyleFlags(s string) string {
	var styledUsages []string
	for _, flagUsage := range strings.Split(s, "\n") {
		styledUsage := flagsRegex.ReplaceAllStringFunc(flagUsage, func(flag string) string {
			return styles.StyleOptions(flag).String()
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
