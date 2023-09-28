package logging

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

var (
	Log = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "|",
		Level:           log.DebugLevel,
		CallerFormatter: log.FilenameCallerFormatter,
	})
)

// Log.ErrorLevelStyle: lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "203",Dark:  "204"}).Foreground(lipgloss.Color("0")),
