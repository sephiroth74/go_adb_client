package types

import (
	"fmt"
	"os"
	"time"
)

type LogcatOptions struct {
	// -e Only prints lines where the log message matches <expr>, where <expr> is a regular expression.
	Expr string
	// -d	Dumps the log to the screen and exits.
	Dump bool
	// -f <filename>	Writes log message output to <filename>. The default is stdout.
	Filename string
	// redirect the output to this file instead of the default output
	File *os.File
	// -s	Equivalent to the filter expression '*:S', which sets priority for all tags to silent and is used to precede a list of filter expressions that add content.
	Tags []LogcatTag
	// -v <format>	Sets the output format for log messages. The default is the threadtime format
	Format string
	// -t '<time>'	Prints the most recent lines since the specified time. This option includes -d functionality. See the -P option for information about quoting parameters with embedded spaces.
	Since *time.Time
	// --pid=<pid> ...
	Pids []string

	Timeout time.Duration
}

func NewLogcatOptions() LogcatOptions {
	return LogcatOptions{
		Expr:     "",
		Dump:     false,
		Filename: "",
		File:     nil,
		Tags:     nil,
		Format:   "",
		Since:    nil,
		Pids:     nil,
		Timeout:  0,
	}
}

type LogcatLevel string

const (
	LogcatVerbose LogcatLevel = "V"
	LogcatDebug   LogcatLevel = "D"
	LogcatInfo    LogcatLevel = "I"
	LogcatWarn    LogcatLevel = "W"
	LogcatError   LogcatLevel = "E"
)

type LogcatTag struct {
	Name  string
	Level LogcatLevel
}

func (l LogcatTag) String() string {
	return fmt.Sprintf("%s:%s", l.Name, l.Level)
}
