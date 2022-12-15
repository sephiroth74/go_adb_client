package constants

import (
	"fmt"
	"time"
)

var SYSTEM_BIN = "/system/bin"

var BUSYBOX = fmt.Sprintf("%s/busybox", SYSTEM_BIN)
var WHICH = fmt.Sprintf("%s/which", SYSTEM_BIN)
var SLEEP = fmt.Sprintf("%s/sleep", SYSTEM_BIN)
var CMD = fmt.Sprintf("%s/cmd", SYSTEM_BIN)

// default timeout for adb shell calls
var DEFAULT_TIMEOUT = time.Duration(5) * time.Second
