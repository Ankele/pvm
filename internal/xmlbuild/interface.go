package xmlbuild

import (
	"github.com/ankele/pvm/internal/model"
	"libvirt.org/go/libvirtxml"
)

func BuildInterface(spec model.InterfaceDefineSpec) (string, error) {
	iface := libvirtxml.Interface{
		Name: spec.Name,
		Start: &libvirtxml.InterfaceStart{
			Mode: defaultString(spec.StartMode, "onboot"),
		},
	}
	if spec.MAC != "" {
		iface.MAC = &libvirtxml.InterfaceMAC{Address: spec.MAC}
	}
	if spec.MTU > 0 {
		iface.MTU = &libvirtxml.InterfaceMTU{Size: uint(spec.MTU)}
	}
	for _, protocol := range spec.Protocols {
		p := libvirtxml.InterfaceProtocol{Family: protocol.Family}
		if protocol.DHCP {
			p.DHCP = &libvirtxml.InterfaceDHCP{}
		}
		if protocol.Address != "" {
			p.IPs = append(p.IPs, libvirtxml.InterfaceIP{
				Address: protocol.Address,
				Prefix:  uint(protocol.Prefix),
			})
		}
		if protocol.Gateway != "" {
			p.Route = append(p.Route, libvirtxml.InterfaceRoute{Gateway: protocol.Gateway})
		}
		iface.Protocol = append(iface.Protocol, p)
	}
	switch spec.Type {
	case "bond":
		bond := &libvirtxml.InterfaceBond{}
		for _, member := range spec.BridgeMembers {
			bond.Interfaces = append(bond.Interfaces, libvirtxml.Interface{Name: member})
		}
		iface.Bond = bond
	case "vlan":
		if len(spec.BridgeMembers) > 0 {
			tag := uint(0)
			iface.VLAN = &libvirtxml.InterfaceVLAN{
				Tag:       &tag,
				Interface: &libvirtxml.Interface{Name: spec.BridgeMembers[0]},
			}
		}
	case "bridge":
		bridge := &libvirtxml.InterfaceBridge{}
		for _, member := range spec.BridgeMembers {
			bridge.Interfaces = append(bridge.Interfaces, libvirtxml.Interface{Name: member})
		}
		iface.Bridge = bridge
	}
	return iface.Marshal()
}
