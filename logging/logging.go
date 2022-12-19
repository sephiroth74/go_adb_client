package logging

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

var (
	Log = zerolog.New(zerolog.ConsoleWriter{
		Out:           os.Stderr,
		NoColor:       false,
		TimeFormat:    "15:04:05.999Z07:00",
		FormatMessage: func(i interface{}) string { return fmt.Sprintf("→ %s", i) },
		FormatLevel: func(i interface{}) string {
			var l string
			if ll, ok := i.(string); ok {
				switch ll {
				case zerolog.LevelTraceValue:
					l = colorize("[TRACE]", 35, false)
				case zerolog.LevelDebugValue:
					l = colorize("[DEBUG]", 34, false)
				case zerolog.LevelInfoValue:
					l = colorize("[INFO] ", 37, false)
				case zerolog.LevelWarnValue:
					l = colorize("[WARN] ", 33, false)
				case zerolog.LevelErrorValue:
					l = colorize(colorize("[ERROR]", 31, false), 1, false)
				case zerolog.LevelFatalValue:
					l = colorize(colorize("[FATAL]", 31, false), 1, false)
				case zerolog.LevelPanicValue:
					l = colorize(colorize("[PANIC]", 31, false), 1, false)
				default:
					l = colorize("[???]", 1, false)
				}
			} else {
				if i == nil {
					l = colorize("[???]", 1, false)
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
