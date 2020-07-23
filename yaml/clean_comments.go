package yaml

import (
	"strings"

	"github.com/thoas/go-funk"
)

func CleanComments(input string) string {
	lines := strings.Split(input, "\n")
	cleanLines := funk.Map(lines, func(line string) string {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			return ""
		}
		return line
	}).([]string)
	return strings.Join(cleanLines, "\n")
}
