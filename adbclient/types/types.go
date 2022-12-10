package types

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"it.sephiroth/adbclient/util"
	"github.com/alecthomas/repr"
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
