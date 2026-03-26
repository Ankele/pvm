package xmlbuild

import (
	"net"

	"github.com/ankele/pvm/internal/model"
	"libvirt.org/go/libvirtxml"
)

func BuildNetwork(spec model.NetworkDefineSpec) (string, error) {
	netDef := libvirtxml.Network{
		Name: spec.Name,
		Bridge: &libvirtxml.NetworkBridge{
			Name:  spec.Bridge,
			STP:   "on",
			Delay: "0",
		},
		Domain: &libvirtxml.NetworkDomain{Name: spec.Domain},
	}
	if spec.ForwardMode != "" {
		netDef.Forward = &libvirtxml.NetworkForward{Mode: spec.ForwardMode}
	}
	if spec.IPv4CIDR != "" {
		ip, ipNet, err := net.ParseCIDR(spec.IPv4CIDR)
		if err != nil {
			return "", err
		}
		ones, _ := ipNet.Mask.Size()
		ipDef := libvirtxml.NetworkIP{
			Address: ip.String(),
			Prefix:  uint(ones),
		}
		if spec.DHCPStart != "" || spec.DHCPEnd != "" {
			ipDef.DHCP = &libvirtxml.NetworkDHCP{
				Ranges: []libvirtxml.NetworkDHCPRange{{
					Start: spec.DHCPStart,
					End:   spec.DHCPEnd,
				}},
			}
		}
		netDef.IPs = []libvirtxml.NetworkIP{ipDef}
	}
	return netDef.Marshal()
}
