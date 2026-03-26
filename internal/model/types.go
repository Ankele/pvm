package model

import "time"

type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

type VMState string

const (
	VMStateRunning   VMState = "running"
	VMStatePaused    VMState = "paused"
	VMStateShutoff   VMState = "shutoff"
	VMStateCrashed   VMState = "crashed"
	VMStateBlocked   VMState = "blocked"
	VMStatePMSuspend VMState = "pmsuspended"
	VMStateUnknown   VMState = "unknown"
)

type LaunchMode string

const (
	LaunchModeClone  LaunchMode = "clone"
	LaunchModeDirect LaunchMode = "direct"
)

type SnapshotKind string

const (
	SnapshotKindDomain SnapshotKind = "domain"
	SnapshotKindVolume SnapshotKind = "volume"
)

type ActionResult struct {
	Message string `json:"message"`
}

type SystemView struct {
	Backend         string `json:"backend"`
	URI             string `json:"uri"`
	Platform        string `json:"platform"`
	SupportsLibvirt bool   `json:"supports_libvirt"`
	SupportsZFS     bool   `json:"supports_zfs"`
}

type GraphicsSpec struct {
	Type     string `json:"type"`
	Listen   string `json:"listen"`
	AutoPort bool   `json:"auto_port"`
	Port     int32  `json:"port"`
	TLSPort  int32  `json:"tls_port"`
}

type VMNICSpec struct {
	VM      string `json:"vm,omitempty"`
	Alias   string `json:"alias,omitempty"`
	Network string `json:"network,omitempty"`
	Bridge  string `json:"bridge,omitempty"`
	MAC     string `json:"mac,omitempty"`
	Model   string `json:"model,omitempty"`
}

type VMNICUpdateSpec struct {
	VMNICSpec
}

type DomainDiskSpec struct {
	Device   string `json:"device"`
	Bus      string `json:"bus"`
	Target   string `json:"target"`
	Source   string `json:"source"`
	Format   string `json:"format"`
	Pool     string `json:"pool,omitempty"`
	Volume   string `json:"volume,omitempty"`
	ReadOnly bool   `json:"read_only,omitempty"`
}

type DomainSpec struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	MemoryMiB   uint64           `json:"memory_mib"`
	VCPU        uint32           `json:"vcpu"`
	Machine     string           `json:"machine,omitempty"`
	Arch        string           `json:"arch,omitempty"`
	Autostart   bool             `json:"autostart,omitempty"`
	Disks       []DomainDiskSpec `json:"disks,omitempty"`
	Networks    []VMNICSpec      `json:"networks,omitempty"`
	Graphics    []GraphicsSpec   `json:"graphics,omitempty"`
}

type VMInstallSpec struct {
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	MemoryMiB     uint64         `json:"memory_mib"`
	VCPU          uint32         `json:"vcpu"`
	Machine       string         `json:"machine,omitempty"`
	Arch          string         `json:"arch,omitempty"`
	Pool          string         `json:"pool"`
	DiskName      string         `json:"disk_name"`
	DiskSizeGiB   uint64         `json:"disk_size_gib"`
	DiskFormat    string         `json:"disk_format,omitempty"`
	ISOPath       string         `json:"iso_path"`
	Networks      []VMNICSpec    `json:"networks,omitempty"`
	Graphics      []GraphicsSpec `json:"graphics,omitempty"`
	Autostart     bool           `json:"autostart,omitempty"`
	SharedImage   bool           `json:"shared_image,omitempty"`
	TargetBus     string         `json:"target_bus,omitempty"`
	ConsoleDriver string         `json:"console_driver,omitempty"`
}

type VMLaunchSpec struct {
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	Mode             LaunchMode     `json:"mode"`
	MemoryMiB        uint64         `json:"memory_mib"`
	VCPU             uint32         `json:"vcpu"`
	Machine          string         `json:"machine,omitempty"`
	Arch             string         `json:"arch,omitempty"`
	ImagePath        string         `json:"image_path,omitempty"`
	SourceVolume     string         `json:"source_volume,omitempty"`
	TargetPool       string         `json:"target_pool,omitempty"`
	TargetVolumeName string         `json:"target_volume_name,omitempty"`
	TargetFormat     string         `json:"target_format,omitempty"`
	Networks         []VMNICSpec    `json:"networks,omitempty"`
	Graphics         []GraphicsSpec `json:"graphics,omitempty"`
	Autostart        bool           `json:"autostart,omitempty"`
}

type PoolDefineSpec struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	SourceName string `json:"source_name,omitempty"`
	TargetPath string `json:"target_path,omitempty"`
	Autostart  bool   `json:"autostart,omitempty"`
}

type VolumeCreateSpec struct {
	Pool            string `json:"pool"`
	Name            string `json:"name"`
	CapacityBytes   uint64 `json:"capacity_bytes"`
	AllocationBytes uint64 `json:"allocation_bytes,omitempty"`
	Format          string `json:"format,omitempty"`
	Type            string `json:"type,omitempty"`
}

type VolumeResizeSpec struct {
	Pool          string `json:"pool"`
	Name          string `json:"name"`
	CapacityBytes uint64 `json:"capacity_bytes"`
}

type NetworkDefineSpec struct {
	Name        string `json:"name"`
	Bridge      string `json:"bridge,omitempty"`
	ForwardMode string `json:"forward_mode,omitempty"`
	Domain      string `json:"domain,omitempty"`
	IPv4CIDR    string `json:"ipv4_cidr,omitempty"`
	DHCPStart   string `json:"dhcp_start,omitempty"`
	DHCPEnd     string `json:"dhcp_end,omitempty"`
	Autostart   bool   `json:"autostart,omitempty"`
}

type InterfaceProtocolSpec struct {
	Family  string `json:"family"`
	Address string `json:"address,omitempty"`
	Prefix  uint32 `json:"prefix,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	DHCP    bool   `json:"dhcp,omitempty"`
}

type InterfaceDefineSpec struct {
	Name          string                  `json:"name"`
	Type          string                  `json:"type"`
	MAC           string                  `json:"mac,omitempty"`
	StartMode     string                  `json:"start_mode,omitempty"`
	MTU           uint32                  `json:"mtu,omitempty"`
	BridgeMembers []string                `json:"bridge_members,omitempty"`
	Protocols     []InterfaceProtocolSpec `json:"protocols,omitempty"`
}

type VMSnapshotCreateSpec struct {
	VM            string `json:"vm"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	IncludeMemory bool   `json:"include_memory,omitempty"`
	Quiesce       bool   `json:"quiesce,omitempty"`
}

type VolumeSnapshotCreateSpec struct {
	Pool   string `json:"pool"`
	Volume string `json:"volume"`
	Name   string `json:"name"`
}

type VMDiskView struct {
	Device   string `json:"device"`
	Bus      string `json:"bus"`
	Target   string `json:"target"`
	Source   string `json:"source"`
	Format   string `json:"format,omitempty"`
	Pool     string `json:"pool,omitempty"`
	Volume   string `json:"volume,omitempty"`
	ReadOnly bool   `json:"read_only,omitempty"`
}

type VMNICView struct {
	Alias   string `json:"alias,omitempty"`
	MAC     string `json:"mac,omitempty"`
	Model   string `json:"model,omitempty"`
	Network string `json:"network,omitempty"`
	Bridge  string `json:"bridge,omitempty"`
}

type VMGraphicsView struct {
	Type      string `json:"type"`
	Listen    string `json:"listen,omitempty"`
	Port      int32  `json:"port,omitempty"`
	TLSPort   int32  `json:"tls_port,omitempty"`
	Websocket int32  `json:"websocket,omitempty"`
}

type VMView struct {
	Name             string           `json:"name"`
	UUID             string           `json:"uuid,omitempty"`
	State            VMState          `json:"state"`
	Persistent       bool             `json:"persistent"`
	Autostart        bool             `json:"autostart"`
	SharedSource     bool             `json:"shared_source,omitempty"`
	SourceMode       string           `json:"source_mode,omitempty"`
	MemoryMiB        uint64           `json:"memory_mib,omitempty"`
	CurrentMemoryMiB uint64           `json:"current_memory_mib,omitempty"`
	VCPU             uint32           `json:"vcpu,omitempty"`
	CPUTimeNanos     uint64           `json:"cpu_time_nanos,omitempty"`
	Disks            []VMDiskView     `json:"disks,omitempty"`
	NICs             []VMNICView      `json:"nics,omitempty"`
	Graphics         []VMGraphicsView `json:"graphics,omitempty"`
}

type GraphicsInfo struct {
	VM      string           `json:"vm"`
	Entries []VMGraphicsView `json:"entries"`
}

type StoragePoolView struct {
	Name            string `json:"name"`
	UUID            string `json:"uuid,omitempty"`
	Type            string `json:"type"`
	Active          bool   `json:"active"`
	Autostart       bool   `json:"autostart"`
	CapacityBytes   uint64 `json:"capacity_bytes,omitempty"`
	AllocationBytes uint64 `json:"allocation_bytes,omitempty"`
	AvailableBytes  uint64 `json:"available_bytes,omitempty"`
	TargetPath      string `json:"target_path,omitempty"`
	SourceName      string `json:"source_name,omitempty"`
}

type VolumeView struct {
	Pool            string `json:"pool"`
	Name            string `json:"name"`
	Key             string `json:"key,omitempty"`
	Path            string `json:"path,omitempty"`
	Type            string `json:"type,omitempty"`
	Format          string `json:"format,omitempty"`
	CapacityBytes   uint64 `json:"capacity_bytes,omitempty"`
	AllocationBytes uint64 `json:"allocation_bytes,omitempty"`
	BackingPath     string `json:"backing_path,omitempty"`
}

type NetworkView struct {
	Name        string `json:"name"`
	UUID        string `json:"uuid,omitempty"`
	Bridge      string `json:"bridge,omitempty"`
	ForwardMode string `json:"forward_mode,omitempty"`
	Domain      string `json:"domain,omitempty"`
	IPv4CIDR    string `json:"ipv4_cidr,omitempty"`
	DHCPStart   string `json:"dhcp_start,omitempty"`
	DHCPEnd     string `json:"dhcp_end,omitempty"`
	Active      bool   `json:"active"`
	Autostart   bool   `json:"autostart"`
}

type InterfaceView struct {
	Name          string   `json:"name"`
	MAC           string   `json:"mac,omitempty"`
	Type          string   `json:"type"`
	StartMode     string   `json:"start_mode,omitempty"`
	Active        bool     `json:"active"`
	MTU           uint32   `json:"mtu,omitempty"`
	BridgeMembers []string `json:"bridge_members,omitempty"`
	Addresses     []string `json:"addresses,omitempty"`
}

type SnapshotView struct {
	Kind        SnapshotKind `json:"kind"`
	Name        string       `json:"name"`
	Parent      string       `json:"parent,omitempty"`
	Domain      string       `json:"domain,omitempty"`
	Pool        string       `json:"pool,omitempty"`
	Volume      string       `json:"volume,omitempty"`
	Description string       `json:"description,omitempty"`
	CreatedAt   time.Time    `json:"created_at,omitempty"`
	State       string       `json:"state,omitempty"`
}
