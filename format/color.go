package format

import "fmt"

const (
	SGR_CODE_RED   = "\\033[31m"
	SGR_CODE_GREEN = "\\033[32m"
)

var SGRCodes = map[string]int{
	"bold":  1,
	"red":   31,
	"green": 32,
	"cyan":  36,
}

func Color(style, string string) string {
	code := SGRCodes[style]

	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, string)
}
