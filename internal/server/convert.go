package server

import (
	"time"

	pvmv1 "github.com/ankele/pvm/api/gen/pvm/v1"
	"github.com/ankele/pvm/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoSystemInfo(v *model.SystemView) *pvmv1.SystemInfo {
	if v == nil {
		return nil
	}
	return &pvmv1.SystemInfo{
		Backend:         v.Backend,
		Uri:             v.URI,
		Platform:        v.Platform,
		SupportsLibvirt: v.SupportsLibvirt,
		SupportsZfs:     v.SupportsZFS,
	}
}

func toProtoVM(v *model.VMView) *pvmv1.VM {
	if v == nil {
		return nil
	}
	out := &pvmv1.VM{
		Name:             v.Name,
		Uuid:             v.UUID,
		State:            string(v.State),
		Persistent:       v.Persistent,
		Autostart:        v.Autostart,
		SharedSource:     v.SharedSource,
		SourceMode:       v.SourceMode,
		MemoryMib:        v.MemoryMiB,
		CurrentMemoryMib: v.CurrentMemoryMiB,
		Vcpu:             v.VCPU,
		CpuTimeNanos:     v.CPUTimeNanos,
	}
	for _, disk := range v.Disks {
		out.Disks = append(out.Disks, &pvmv1.VMDisk{
			Device:   disk.Device,
			Bus:      disk.Bus,
			Target:   disk.Target,
			Source:   disk.Source,
			Format:   disk.Format,
			Pool:     disk.Pool,
			Volume:   disk.Volume,
			ReadOnly: disk.ReadOnly,
		})
	}
	for _, nic := range v.NICs {
		out.Nics = append(out.Nics, &pvmv1.VMNIC{
			Alias:   nic.Alias,
			Mac:     nic.MAC,
			Model:   nic.Model,
			Network: nic.Network,
			Bridge:  nic.Bridge,
		})
	}
	for _, g := range v.Graphics {
		out.Graphics = append(out.Graphics, &pvmv1.VMGraphics{
			Type:      g.Type,
			Listen:    g.Listen,
			Port:      g.Port,
			TlsPort:   g.TLSPort,
			Websocket: g.Websocket,
		})
	}
	return out
}

func toProtoGraphicsInfo(v *model.GraphicsInfo) *pvmv1.GraphicsInfo {
	if v == nil {
		return nil
	}
	out := &pvmv1.GraphicsInfo{Vm: v.VM}
	for _, g := range v.Entries {
		out.Entries = append(out.Entries, &pvmv1.VMGraphics{
			Type:      g.Type,
			Listen:    g.Listen,
			Port:      g.Port,
			TlsPort:   g.TLSPort,
			Websocket: g.Websocket,
		})
	}
	return out
}

func toProtoPool(v *model.StoragePoolView) *pvmv1.StoragePool {
	if v == nil {
		return nil
	}
	return &pvmv1.StoragePool{
		Name:            v.Name,
		Uuid:            v.UUID,
		Type:            v.Type,
		Active:          v.Active,
		Autostart:       v.Autostart,
		CapacityBytes:   v.CapacityBytes,
		AllocationBytes: v.AllocationBytes,
		AvailableBytes:  v.AvailableBytes,
		TargetPath:      v.TargetPath,
		SourceName:      v.SourceName,
	}
}

func toProtoVolume(v *model.VolumeView) *pvmv1.Volume {
	if v == nil {
		return nil
	}
	return &pvmv1.Volume{
		Pool:            v.Pool,
		Name:            v.Name,
		Key:             v.Key,
		Path:            v.Path,
		Type:            v.Type,
		Format:          v.Format,
		CapacityBytes:   v.CapacityBytes,
		AllocationBytes: v.AllocationBytes,
		BackingPath:     v.BackingPath,
	}
}

func toProtoNetwork(v *model.NetworkView) *pvmv1.Network {
	if v == nil {
		return nil
	}
	return &pvmv1.Network{
		Name:        v.Name,
		Uuid:        v.UUID,
		Bridge:      v.Bridge,
		ForwardMode: v.ForwardMode,
		Domain:      v.Domain,
		Ipv4Cidr:    v.IPv4CIDR,
		DhcpStart:   v.DHCPStart,
		DhcpEnd:     v.DHCPEnd,
		Active:      v.Active,
		Autostart:   v.Autostart,
	}
}

func toProtoInterface(v *model.InterfaceView) *pvmv1.HostInterface {
	if v == nil {
		return nil
	}
	return &pvmv1.HostInterface{
		Name:          v.Name,
		Mac:           v.MAC,
		Type:          v.Type,
		StartMode:     v.StartMode,
		Active:        v.Active,
		Mtu:           v.MTU,
		BridgeMembers: append([]string(nil), v.BridgeMembers...),
		Addresses:     append([]string(nil), v.Addresses...),
	}
}

func toProtoSnapshot(v *model.SnapshotView) *pvmv1.Snapshot {
	if v == nil {
		return nil
	}
	return &pvmv1.Snapshot{
		Kind:        string(v.Kind),
		Name:        v.Name,
		Parent:      v.Parent,
		Domain:      v.Domain,
		Pool:        v.Pool,
		Volume:      v.Volume,
		Description: v.Description,
		CreatedAt:   toProtoTime(v.CreatedAt),
		State:       v.State,
	}
}

func toProtoTime(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func domainSpecFromProto(in *pvmv1.DomainSpec) model.DomainSpec {
	if in == nil {
		return model.DomainSpec{}
	}
	out := model.DomainSpec{
		Name:        in.GetName(),
		Description: in.GetDescription(),
		MemoryMiB:   in.GetMemoryMib(),
		VCPU:        in.GetVcpu(),
		Machine:     in.GetMachine(),
		Arch:        in.GetArch(),
		Autostart:   in.GetAutostart(),
	}
	for _, disk := range in.GetDisks() {
		out.Disks = append(out.Disks, model.DomainDiskSpec{
			Device:   disk.GetDevice(),
			Bus:      disk.GetBus(),
			Target:   disk.GetTarget(),
			Source:   disk.GetSource(),
			Format:   disk.GetFormat(),
			Pool:     disk.GetPool(),
			Volume:   disk.GetVolume(),
			ReadOnly: disk.GetReadOnly(),
		})
	}
	for _, nic := range in.GetNetworks() {
		out.Networks = append(out.Networks, nicSpecFromProto(nic))
	}
	for _, g := range in.GetGraphics() {
		out.Graphics = append(out.Graphics, graphicsSpecFromProto(g))
	}
	return out
}

func installSpecFromProto(in *pvmv1.InstallVMSpec) model.VMInstallSpec {
	if in == nil {
		return model.VMInstallSpec{}
	}
	out := model.VMInstallSpec{
		Name:        in.GetName(),
		Description: in.GetDescription(),
		MemoryMiB:   in.GetMemoryMib(),
		VCPU:        in.GetVcpu(),
		Machine:     in.GetMachine(),
		Arch:        in.GetArch(),
		Pool:        in.GetPool(),
		DiskName:    in.GetDiskName(),
		DiskSizeGiB: in.GetDiskSizeGib(),
		DiskFormat:  in.GetDiskFormat(),
		ISOPath:     in.GetIsoPath(),
		Autostart:   in.GetAutostart(),
	}
	for _, nic := range in.GetNetworks() {
		out.Networks = append(out.Networks, nicSpecFromProto(nic))
	}
	for _, g := range in.GetGraphics() {
		out.Graphics = append(out.Graphics, graphicsSpecFromProto(g))
	}
	return out
}

func launchSpecFromProto(in *pvmv1.LaunchVMSpec) model.VMLaunchSpec {
	if in == nil {
		return model.VMLaunchSpec{}
	}
	out := model.VMLaunchSpec{
		Name:             in.GetName(),
		Description:      in.GetDescription(),
		Mode:             model.LaunchMode(in.GetMode()),
		MemoryMiB:        in.GetMemoryMib(),
		VCPU:             in.GetVcpu(),
		Machine:          in.GetMachine(),
		Arch:             in.GetArch(),
		ImagePath:        in.GetImagePath(),
		SourceVolume:     in.GetSourceVolume(),
		TargetPool:       in.GetTargetPool(),
		TargetVolumeName: in.GetTargetVolumeName(),
		TargetFormat:     in.GetTargetFormat(),
		Autostart:        in.GetAutostart(),
	}
	for _, nic := range in.GetNetworks() {
		out.Networks = append(out.Networks, nicSpecFromProto(nic))
	}
	for _, g := range in.GetGraphics() {
		out.Graphics = append(out.Graphics, graphicsSpecFromProto(g))
	}
	return out
}

func nicSpecFromProto(in *pvmv1.NICSpec) model.VMNICSpec {
	if in == nil {
		return model.VMNICSpec{}
	}
	return model.VMNICSpec{
		Alias:   in.GetAlias(),
		Network: in.GetNetwork(),
		Bridge:  in.GetBridge(),
		MAC:     in.GetMac(),
		Model:   in.GetModel(),
	}
}

func graphicsSpecFromProto(in *pvmv1.GraphicsSpec) model.GraphicsSpec {
	if in == nil {
		return model.GraphicsSpec{}
	}
	return model.GraphicsSpec{
		Type:     in.GetType(),
		Listen:   in.GetListen(),
		AutoPort: in.GetAutoPort(),
		Port:     in.GetPort(),
		TLSPort:  in.GetTlsPort(),
	}
}

func poolSpecFromProto(in *pvmv1.PoolSpec) model.PoolDefineSpec {
	if in == nil {
		return model.PoolDefineSpec{}
	}
	return model.PoolDefineSpec{
		Name:       in.GetName(),
		Type:       in.GetType(),
		SourceName: in.GetSourceName(),
		TargetPath: in.GetTargetPath(),
		Autostart:  in.GetAutostart(),
	}
}

func volumeCreateSpecFromProto(in *pvmv1.VolumeCreateSpec) model.VolumeCreateSpec {
	if in == nil {
		return model.VolumeCreateSpec{}
	}
	return model.VolumeCreateSpec{
		Pool:            in.GetPool(),
		Name:            in.GetName(),
		CapacityBytes:   in.GetCapacityBytes(),
		AllocationBytes: in.GetAllocationBytes(),
		Format:          in.GetFormat(),
		Type:            in.GetType(),
	}
}

func volumeResizeSpecFromProto(in *pvmv1.VolumeResizeSpec) model.VolumeResizeSpec {
	if in == nil {
		return model.VolumeResizeSpec{}
	}
	return model.VolumeResizeSpec{
		Pool:          in.GetPool(),
		Name:          in.GetName(),
		CapacityBytes: in.GetCapacityBytes(),
	}
}

func networkSpecFromProto(in *pvmv1.NetworkSpec) model.NetworkDefineSpec {
	if in == nil {
		return model.NetworkDefineSpec{}
	}
	return model.NetworkDefineSpec{
		Name:        in.GetName(),
		Bridge:      in.GetBridge(),
		ForwardMode: in.GetForwardMode(),
		Domain:      in.GetDomain(),
		IPv4CIDR:    in.GetIpv4Cidr(),
		DHCPStart:   in.GetDhcpStart(),
		DHCPEnd:     in.GetDhcpEnd(),
		Autostart:   in.GetAutostart(),
	}
}

func interfaceSpecFromProto(in *pvmv1.InterfaceSpec) model.InterfaceDefineSpec {
	if in == nil {
		return model.InterfaceDefineSpec{}
	}
	out := model.InterfaceDefineSpec{
		Name:          in.GetName(),
		Type:          in.GetType(),
		MAC:           in.GetMac(),
		StartMode:     in.GetStartMode(),
		MTU:           in.GetMtu(),
		BridgeMembers: append([]string(nil), in.GetBridgeMembers()...),
	}
	for _, p := range in.GetProtocols() {
		out.Protocols = append(out.Protocols, model.InterfaceProtocolSpec{
			Family:  p.GetFamily(),
			Address: p.GetAddress(),
			Prefix:  p.GetPrefix(),
			Gateway: p.GetGateway(),
			DHCP:    p.GetDhcp(),
		})
	}
	return out
}

func vmSnapshotSpecFromProto(in *pvmv1.VMSnapshotSpec) model.VMSnapshotCreateSpec {
	if in == nil {
		return model.VMSnapshotCreateSpec{}
	}
	return model.VMSnapshotCreateSpec{
		VM:            in.GetVm(),
		Name:          in.GetName(),
		Description:   in.GetDescription(),
		IncludeMemory: in.GetIncludeMemory(),
		Quiesce:       in.GetQuiesce(),
	}
}

func volumeSnapshotSpecFromProto(in *pvmv1.VolumeSnapshotSpec) model.VolumeSnapshotCreateSpec {
	if in == nil {
		return model.VolumeSnapshotCreateSpec{}
	}
	return model.VolumeSnapshotCreateSpec{
		Pool:   in.GetPool(),
		Volume: in.GetVolume(),
		Name:   in.GetName(),
	}
}
