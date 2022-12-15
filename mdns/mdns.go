package mdns

import (
	"regexp"
	"strings"

	"it.sephiroth/adbclient/connection"
	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
)

type Mdns struct {
	Conn *connection.Connection
}

func (m Mdns) Check() (transport.Result, error) {
	return transport.Invoke(&m.Conn.ADBPath, 0, "mdns", "check")
}

func (m Mdns) Services() ([]types.MdnsDevice, error) {
	// adb-JA37001FF3	_adb._tcp.	192.168.1.105:5555
	result, err := transport.Invoke(&m.Conn.ADBPath, 0, "mdns", "services")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(result.Output(), "\n")
	devices := []types.MdnsDevice{}

	if len(lines) > 1 {
		r := regexp.MustCompile(`([^\s\t]+)\t([^\s\t]+)\t([^\n]+)`)

		for i := 1; i < len(lines); i++ {
			m := r.FindStringSubmatch(lines[i])
			if len(m) > 3 {
				device, err := types.NewMdnsDevice(m[1], m[2], &m[3])
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
