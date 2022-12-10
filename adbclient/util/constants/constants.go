package constants

import "fmt"

var SYSTEM_BIN = "/system/bin"

var BUSYBOX = fmt.Sprintf("%s/busybox", SYSTEM_BIN)
var WHICH 	= fmt.Sprintf("%s/which", SYSTEM_BIN)
var SLEEP 	= fmt.Sprintf("%s/sleep", SYSTEM_BIN)
