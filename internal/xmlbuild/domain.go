package xmlbuild

import (
	"fmt"

	"github.com/ankele/pvm/internal/model"
	"libvirt.org/go/libvirtxml"
)

func BuildDomain(spec model.DomainSpec) (string, error) {
	domain := baseDomain(spec.Name, spec.Description, spec.MemoryMiB, spec.VCPU, spec.Machine, spec.Arch, spec.Autostart)
	devices := domain.Devices

	for _, disk := range spec.Disks {
		devices.Disks = append(devices.Disks, diskToXML(disk))
	}
	for idx, nic := range spec.Networks {
		devices.Interfaces = append(devices.Interfaces, nicToXML(nic, idx))
	}
	if len(spec.Graphics) == 0 {
		spec.Graphics = []model.GraphicsSpec{{Type: "vnc", Listen: "127.0.0.1", AutoPort: true}}
	}
	for _, graphics := range spec.Graphics {
		devices.Graphics = append(devices.Graphics, graphicsToXML(graphics))
	}
	devices.Videos = []libvirtxml.DomainVideo{{
		Model: libvirtxml.DomainVideoModel{Type: "virtio"},
	}}
	domain.Devices = devices
	return domain.Marshal()
}

func BuildInstallDomain(spec model.VMInstallSpec, rootDisk model.DomainDiskSpec) (string, error) {
	domain := baseDomain(spec.Name, spec.Description, spec.MemoryMiB, spec.VCPU, spec.Machine, spec.Arch, spec.Autostart)
	devices := domain.Devices
	devices.Disks = append(devices.Disks, diskToXML(rootDisk))
	devices.Disks = append(devices.Disks, libvirtxml.DomainDisk{
		Device: "cdrom",
		Driver: &libvirtxml.DomainDiskDriver{
			Name: "qemu",
			Type: "raw",
		},
		Source: &libvirtxml.DomainDiskSource{
			File: &libvirtxml.DomainDiskSourceFile{File: spec.ISOPath},
		},
		Target: &libvirtxml.DomainDiskTarget{
			Dev: "sda",
			Bus: "sata",
		},
		ReadOnly: &libvirtxml.DomainDiskReadOnly{},
	})
	for idx, nic := range spec.Networks {
		devices.Interfaces = append(devices.Interfaces, nicToXML(nic, idx))
	}
	if len(spec.Graphics) == 0 {
		spec.Graphics = []model.GraphicsSpec{{Type: "vnc", Listen: "127.0.0.1", AutoPort: true}}
	}
	for _, graphics := range spec.Graphics {
		devices.Graphics = append(devices.Graphics, graphicsToXML(graphics))
	}
	devices.Videos = []libvirtxml.DomainVideo{{
		Model: libvirtxml.DomainVideoModel{Type: "virtio"},
	}}
	domain.Devices = devices
	domain.OS.BootDevices = []libvirtxml.DomainBootDevice{
		{Dev: "cdrom"},
		{Dev: "hd"},
	}
	return domain.Marshal()
}

func baseDomain(name, description string, memoryMiB uint64, vcpu uint32, machine, arch string, autostart bool) libvirtxml.Domain {
	memValue := uint(memoryMiB)
	vcpuValue := uint(vcpu)
	domain := libvirtxml.Domain{
		Type:        "kvm",
		Name:        name,
		Description: description,
		Memory: &libvirtxml.DomainMemory{
			Unit:  "MiB",
			Value: memValue,
		},
		CurrentMemory: &libvirtxml.DomainCurrentMemory{
			Unit:  "MiB",
			Value: memValue,
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: vcpuValue,
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "restart",
		Devices: &libvirtxml.DomainDeviceList{
			Consoles: []libvirtxml.DomainConsole{{
				Target: &libvirtxml.DomainConsoleTarget{Type: "serial", Port: new(uint)},
				Source: &libvirtxml.DomainChardevSource{
					Pty: &libvirtxml.DomainChardevSourcePty{},
				},
			}},
			Serials: []libvirtxml.DomainSerial{{
				Target: &libvirtxml.DomainSerialTarget{Port: new(uint)},
				Source: &libvirtxml.DomainChardevSource{
					Pty: &libvirtxml.DomainChardevSourcePty{},
				},
			}},
			MemBalloon: &libvirtxml.DomainMemBalloon{Model: "virtio"},
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Type:    "hvm",
				Machine: machine,
				Arch:    arch,
			},
			BootDevices: []libvirtxml.DomainBootDevice{{Dev: "hd"}},
			BootMenu: &libvirtxml.DomainBootMenu{
				Enable: "yes",
			},
		},
	}
	if domain.OS.Type.Machine == "" {
		domain.OS.Type.Machine = "q35"
	}
	if domain.OS.Type.Arch == "" {
		domain.OS.Type.Arch = "x86_64"
	}
	if autostart {
		// libvirt autostart is not encoded in domain XML; backend sets it after define/create.
	}
	return domain
}

func diskToXML(spec model.DomainDiskSpec) libvirtxml.DomainDisk {
	disk := libvirtxml.DomainDisk{
		Device: spec.Device,
		Driver: &libvirtxml.DomainDiskDriver{
			Name: "qemu",
			Type: spec.Format,
		},
		Target: &libvirtxml.DomainDiskTarget{
			Dev: spec.Target,
			Bus: spec.Bus,
		},
	}
	if disk.Device == "" {
		disk.Device = "disk"
	}
	if disk.Driver.Type == "" {
		disk.Driver.Type = "raw"
	}
	switch {
	case spec.Pool != "" && spec.Volume != "":
		disk.Source = &libvirtxml.DomainDiskSource{
			Volume: &libvirtxml.DomainDiskSourceVolume{
				Pool:   spec.Pool,
				Volume: spec.Volume,
			},
		}
	case spec.Source != "":
		disk.Source = &libvirtxml.DomainDiskSource{
			File: &libvirtxml.DomainDiskSourceFile{File: spec.Source},
		}
	default:
		disk.Source = &libvirtxml.DomainDiskSource{}
	}
	if spec.ReadOnly {
		disk.ReadOnly = &libvirtxml.DomainDiskReadOnly{}
	}
	return disk
}

func nicToXML(spec model.VMNICSpec, idx int) libvirtxml.DomainInterface {
	nic := libvirtxml.DomainInterface{
		Model: &libvirtxml.DomainInterfaceModel{Type: defaultString(spec.Model, "virtio")},
		Alias: &libvirtxml.DomainAlias{Name: defaultString(spec.Alias, fmt.Sprintf("net%d", idx))},
	}
	if spec.MAC != "" {
		nic.MAC = &libvirtxml.DomainInterfaceMAC{Address: spec.MAC}
	}
	switch {
	case spec.Network != "":
		nic.Source = &libvirtxml.DomainInterfaceSource{
			Network: &libvirtxml.DomainInterfaceSourceNetwork{Network: spec.Network},
		}
	case spec.Bridge != "":
		nic.Source = &libvirtxml.DomainInterfaceSource{
			Bridge: &libvirtxml.DomainInterfaceSourceBridge{Bridge: spec.Bridge},
		}
	default:
		nic.Source = &libvirtxml.DomainInterfaceSource{
			Network: &libvirtxml.DomainInterfaceSourceNetwork{Network: "default"},
		}
	}
	return nic
}

func graphicsToXML(spec model.GraphicsSpec) libvirtxml.DomainGraphic {
	switch spec.Type {
	case "spice":
		graphics := libvirtxml.DomainGraphic{
			Spice: &libvirtxml.DomainGraphicSpice{
				Listen:   defaultString(spec.Listen, "127.0.0.1"),
				TLSPort:  int(spec.TLSPort),
				AutoPort: yesNo(spec.AutoPort || spec.Port == 0),
			},
		}
		if spec.Port != 0 {
			graphics.Spice.Port = int(spec.Port)
		}
		return graphics
	default:
		graphics := libvirtxml.DomainGraphic{
			VNC: &libvirtxml.DomainGraphicVNC{
				Listen:   defaultString(spec.Listen, "127.0.0.1"),
				AutoPort: yesNo(spec.AutoPort || spec.Port == 0),
			},
		}
		if spec.Port != 0 {
			graphics.VNC.Port = int(spec.Port)
		}
		return graphics
	}
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
