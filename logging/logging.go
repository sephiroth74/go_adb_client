package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var (
	Log = zerolog.New(zerolog.ConsoleWriter{
		Out:           os.Stderr,
		NoColor:       false,
		TimeFormat:    "15:04:05",
		FormatMessage: func(i interface{}) string { return fmt.Sprintf("» %s", i) },
		FormatLevel: func(i interface{}) string {
			var l string
			if ll, ok := i.(string); ok {
				switch ll {
				case zerolog.LevelTraceValue:
					l = colorize("T", 35, false)
				case zerolog.LevelDebugValue:
					l = colorize("D", 34, false)
				case zerolog.LevelInfoValue:
					l = colorize("I", 37, false)
				case zerolog.LevelWarnValue:
					l = colorize("W", 33, false)
				case zerolog.LevelErrorValue:
					l = colorize("E", 31, false)
				case zerolog.LevelFatalValue:
					l = colorize("F", 31, false)
				case zerolog.LevelPanicValue:
					l = colorize("P", 31, false)
				default:
					l = colorize("?", 1, false)
				}
			} else {
				if i == nil {
					l = colorize("?", 1, false)
				} else {
					l = strings.ToUpper(fmt.Sprintf("%s", i))[0:3]
				}
			}
			return l
		},
	}).With().Timestamp().Logger()
)

func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
