package scanner

import (
	"fmt"
	"github.com/pterm/pterm"
	"net"
	"sync"
	"time"
)

type Scanner struct {
	Results chan *string
}

func NewScanner() *Scanner {
	return &Scanner{
		Results: make(chan *string),
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

func worker(index int, host string, ch chan *string, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()
	pterm.Debug.Printf("[%d] Trying to connect to %s", index, host)
	conn, err := net.DialTimeout("tcp", host, time.Duration(1)*time.Second)
	if err != nil {
		ch <- nil
		return
	}

	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	var remoteAddr = conn.RemoteAddr().String()
	ch <- &remoteAddr
}
