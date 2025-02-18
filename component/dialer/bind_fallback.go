package dialer

import (
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/Dreamacro/clash/component/iface"
)

func lookupLocalAddr(ifaceName string, network string, destination netip.Addr, port int) (net.Addr, error) {
	ifaceObj, err := iface.ResolveInterface(ifaceName)
	if err != nil {
		return nil, err
	}

	var addr *netip.Prefix
	switch network {
	case "udp4", "tcp4":
		addr, err = ifaceObj.PickIPv4Addr(destination)
	case "tcp6", "udp6":
		addr, err = ifaceObj.PickIPv6Addr(destination)
	default:
		if destination.IsValid() {
			if destination.Is4() {
				addr, err = ifaceObj.PickIPv4Addr(destination)
			} else {
				addr, err = ifaceObj.PickIPv6Addr(destination)
			}
		} else {
			addr, err = ifaceObj.PickIPv4Addr(destination)
		}
	}
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(network, "tcp") {
		return &net.TCPAddr{
			IP:   addr.Addr().AsSlice(),
			Port: port,
		}, nil
	} else if strings.HasPrefix(network, "udp") {
		return &net.UDPAddr{
			IP:   addr.Addr().AsSlice(),
			Port: port,
		}, nil
	}

	return nil, iface.ErrAddrNotFound
}

func fallbackBindIfaceToListenConfig(ifaceName string, _ *net.ListenConfig, network, address string) (string, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		port = "0"
	}

	local, _ := strconv.ParseUint(port, 10, 16)

	addr, err := lookupLocalAddr(ifaceName, network, netip.Addr{}, int(local))
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}
