package mdns

import (
	"regexp"
	"strings"

	"github.com/sephiroth74/go_adb_client/connection"
	"github.com/sephiroth74/go_adb_client/process"
	"github.com/sephiroth74/go_adb_client/types"
)

type Mdns struct {
	Conn *connection.Connection
}

func (m Mdns) Check() (process.OutputResult, error) {
	return process.SimpleOutput(m.Conn.NewAdbCommand().WithArgs("mdns", "check"), m.Conn.Verbose)
}

func (m Mdns) Services() ([]types.MdnsDevice, error) {
	// adb-JA37001FF3	_adb._tcp.	192.168.1.105:5555
	cmd := m.Conn.NewAdbCommand().WithCommand("mdns").WithArgs("services")
	result, err := process.SimpleOutput(cmd, m.Conn.Verbose)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(result.Output(), "\n")
	var devices []types.MdnsDevice

	if len(lines) > 1 {
		r := regexp.MustCompile(`([^\s\t]+)\t([^\s\t]+)\t([^\n]+)`)

		for i := 1; i < len(lines); i++ {
			m := r.FindStringSubmatch(lines[i])
			if len(m) > 3 {
				device, err := types.NewMdnsDevice(&m[1], m[2], &m[3])
				if err == nil {
					devices = append(devices, *device)
				}
			}
		}
	}

	return devices, err
}

func NewMdns(conn *connection.Connection) *Mdns {
	mdns := new(Mdns)
	mdns.Conn = conn
	return mdns
}
