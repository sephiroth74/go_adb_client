package scanner

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	adbclient "github.com/sephiroth74/go_adb_client"
	"github.com/sephiroth74/go_adb_client/types"
)

type Scanner struct {
	Results chan *types.TcpDevice
}

func NewScanner() *Scanner {
	return &Scanner{
		Results: make(chan *types.TcpDevice),
	}
}

func (s *Scanner) Scan() {
	go func() {
		wg := new(sync.WaitGroup)
		baseHost := "192.168.1.%d:5555"
		// Adding routines to workgroup and running then
		for i := 1; i <= 255; i++ {
			host := fmt.Sprintf(baseHost, i)
			wg.Add(1)
			go worker(i, host, s.Results, wg)
		}
		wg.Wait()
		close(s.Results)
	}()
}

func worker(index int, host string, ch chan *types.TcpDevice, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()
	conn, err := net.DialTimeout("tcp", host, time.Duration(1)*time.Second)
	if err != nil {
		ch <- nil
		return
	}

	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	var remoteAddr = conn.RemoteAddr().String()
	deviceName, _ := getDeviceName(remoteAddr)

	device, err := types.NewTcpDevice(deviceName, &remoteAddr)
	if err != nil {
		ch <- nil
		return
	}
	ch <- device
}

func getDeviceName(deviceAddr string) (string, error) {
	slice := strings.Split(deviceAddr, ":")

	if len(slice) == 1 {
		slice = append(slice, "5555")
	}

	if len(slice) != 2 {
		return "", fmt.Errorf("invalid address %s", deviceAddr)
	}

	port, err := strconv.Atoi(slice[1])
	if err != nil {
		return "", err
	}

	ip := net.ParseIP(slice[0])
	client := adbclient.NewClient(types.ClientAddr{
		IP:   ip,
		Port: port,
	}, nil, true)

	device := adbclient.NewDevice(client)
	_, err = device.Client.Connect(1 * time.Second)

	if err != nil {
		return "", err
	}

	name := device.Name()
	_, _ = device.Client.Disconnect()

	if name != nil {
		return *name, nil
	} else {
		return "", fmt.Errorf("name not found")
	}
}
