package main

import (
	"log"
	"net"
	"testing"

	"it.sephiroth/adbclient"
	"pkg.re/essentialkaos/ek.v12/env"
)

var path   = env.Which("adb")
var client = adbclient.NewClient(net.IPv4(192, 168, 1, 121), 5555)

func TestConnect(t *testing.T) {
	log.Println("TestConnect")
	client.Connect()
}

// func main() {
// 	path := env.Which("adb")
// 	client := adbclient.NewClient(net.IPv4(192, 168, 1, 121), 5555)

// 	fmt.Println("path: ", path)
// 	fmt.Println("client: ", client)

// 	result := client.Disconnect()

// 	log.Println("exit code: ", result.ExitCode)

// 	if result.ExitCode != 0 {
// 		log.Println("stderr: ", string(result.Stderr))
// 	} else {
// 		log.Println("stdout: ", string(result.Stdout))
// 	}

// 	client.Connect()
// }