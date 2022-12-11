package types

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/alecthomas/repr"
	"it.sephiroth/adbclient/util"
)

// ClientAddr

type Serial interface {
	ClientAddr | Device | MdnsDevice
	Serial() string
	String() string
}

type ClientAddr struct {
	IP   net.IP
	Port int
}

func (c ClientAddr) ToString() string {
	return fmt.Sprintf("%s:%d", c.IP, c.Port)
}

func (c ClientAddr) String() string {
	return fmt.Sprintf("ClientAddr{IP:%s, Port=%d}", c.IP, c.Port)
}

func (c ClientAddr) Serial() string {
	return fmt.Sprintf("%s:%d", c.IP, c.Port)
}

// Device

type Device struct {
	Addr      ClientAddr
	Product   []byte
	Model     []byte
	Device    []byte
	Transport []byte
}

func (m Device) Serial() string {
	return m.Addr.Serial()
}

func (m Device) String() string {
	return repr.String(m)
}

func NewClientAddress(addr *string) (*ClientAddr, error) {
	ip_port := strings.Split(*addr, ":")
	ip, err := util.Map(strings.Split(ip_port[0], "."), func(s string) (byte, error) {
		v, e := strconv.Atoi(s)
		if e != nil {
			return 0, e
		}
		return byte(v), nil
	})

	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(ip_port[1])

	if err != nil {
		return nil, err
	}

	address := new(ClientAddr)
	address.IP = net.IPv4(ip[0], ip[1], ip[2], ip[3])
	address.Port = port
	return address, nil
}

func NewDevice(addr *string) (*Device, error) {
	client_addr, err := NewClientAddress(addr)

	if err != nil {
		return nil, err
	}

	device := new(Device)
	device.Addr = *client_addr
	return device, nil
}

type MdnsDevice struct {
	Name           string
	ConnectionType string
	Address        ClientAddr
}

func (m MdnsDevice) Serial() string {
	return fmt.Sprintf("%s.%s", m.Name, m.ConnectionType)
}

func (m MdnsDevice) String() string {
	return repr.String(m)
}

func NewMdnsDevice(name string, ctype string, addr *string) (*MdnsDevice, error) {
	client_addr, err := NewClientAddress(addr)
	if err != nil {
		return nil, err
	}

	mdns := new(MdnsDevice)
	mdns.Address = *client_addr
	mdns.Name = name
	mdns.ConnectionType = ctype
	return mdns, nil
}

type Intent struct {
	Action    string
	Data      string
	MimeType  string
	Category  string
	Component string
	Flags     int32
	Extra     Extras
}

type Extras struct {
	Es                      map[string]string
	Ez                      map[string]bool
	Ei                      map[string]int
	El                      map[string]int64
	Ef                      map[string]float32
	Eu                      map[string]string
	Ecn                     map[string]string
	Eia                     map[string][]int
	Ela                     map[string][]int64
	Efa                     map[string][]float32
	GrantReadUriPermission  bool
	GrantWriteUriPermission bool
	ExcludeStoppedPackages  bool
	IncludeStoppedPackages  bool
}

func (i Intent) String() string {
	// sb := strings.Builder{}
	sb := []string{}
	if i.Action != "" {
		sb = append(sb, fmt.Sprintf("-a %s", i.Action))
	}

	if i.Data != "" {
		sb = append(sb, fmt.Sprintf("-d %s", i.Data))
	}

	if i.MimeType != "" {
		sb = append(sb, fmt.Sprintf("-t %s", i.MimeType))
	}

	if i.Category != "" {
		sb = append(sb, fmt.Sprintf("-c %s", i.Category))
	}

	if i.Component != "" {
		sb = append(sb, fmt.Sprintf("-n %s", i.Component))
	}

	if !reflect.DeepEqual(Extras{}, i.Extra) {
		sb = append(sb, i.Extra.String())
	}

	return strings.Join(sb, " ")
}

func (e Extras) String() string {
	s := []string{}
	if e.Es != nil && len(e.Es) > 0 {
		for k, v := range e.Es {
			s = append(s, fmt.Sprintf("--es %s %s", k, v))
		}
	}

	if e.Ez != nil && len(e.Ez) > 0 {
		for k, v := range e.Ez {
			s = append(s, fmt.Sprintf("--ez %s %t", k, v))
		}
	}

	if e.Ei != nil && len(e.Ei) > 0 {
		for k, v := range e.Ei {
			s = append(s, fmt.Sprintf("--ei %s %d", k, v))
		}
	}

	if e.El != nil && len(e.El) > 0 {
		for k, v := range e.El {
			s = append(s, fmt.Sprintf("--el %s %d", k, v))
		}
	}

	if e.Ef != nil && len(e.Ef) > 0 {
		for k, v := range e.Ef {
			s = append(s, fmt.Sprintf("--ef %s %f", k, v))
		}
	}

	if e.Eu != nil && len(e.Eu) > 0 {
		for k, v := range e.Eu {
			s = append(s, fmt.Sprintf("--eu %s %s", k, v))
		}
	}

	if e.Ecn != nil && len(e.Ecn) > 0 {
		for k, v := range e.Ecn {
			s = append(s, fmt.Sprintf("--ecn %s %s", k, v))
		}
	}

	if e.Eia != nil && len(e.Eia) > 0 {
		for k, v := range e.Eia {
			s = append(s, fmt.Sprintf("--eia %s %s", k, strings.Trim(strings.Replace(fmt.Sprint(v), " ", ",", -1), "[]")))
		}
	}

	if e.Ela != nil && len(e.Ela) > 0 {
		for k, v := range e.Ela {
			s = append(s, fmt.Sprintf("--ela %s %s", k, strings.Trim(strings.Replace(fmt.Sprint(v), " ", ",", -1), "[]")))
		}
	}

	if e.Efa != nil && len(e.Efa) > 0 {
		for k, v := range e.Efa {
			s = append(s, fmt.Sprintf("--ela %s %s", k, strings.Trim(strings.Replace(fmt.Sprint(v), " ", ",", -1), "[]")))
		}
	}

	if e.GrantReadUriPermission {
		s = append(s, "--grant-read-uri-permission")
	}

	if e.GrantWriteUriPermission {
		s = append(s, "--grant-write-uri-permission")
	}

	if e.ExcludeStoppedPackages {
		s = append(s, "--exclude-stopped-packages")
	}

	if e.IncludeStoppedPackages {
		s = append(s, "--include-stopped-packages")
	}

	return strings.Join(s, " ")
}

type IntentBuilder struct {
	Intent *Intent
}

func NewIntent() *Intent {
	return &Intent{
		Flags: 0,
		Extra: Extras{
			Es:  make(map[string]string),
			Ez:  make(map[string]bool),
			Ei:  make(map[string]int),
			El:  make(map[string]int64),
			Ef:  make(map[string]float32),
			Eu:  make(map[string]string),
			Ecn: make(map[string]string),
			Eia: make(map[string][]int),
			Ela: make(map[string][]int64),
			Efa: make(map[string][]float32),
		},
	}
}
