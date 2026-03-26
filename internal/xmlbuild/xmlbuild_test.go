package xmlbuild

import (
	"strings"
	"testing"

	"github.com/ankele/pvm/internal/model"
)

func TestBuildDomainIncludesVolumeDiskNICAndDefaultGraphics(t *testing.T) {
	t.Parallel()

	xml, err := BuildDomain(model.DomainSpec{
		Name:      "demo",
		MemoryMiB: 2048,
		VCPU:      2,
		Disks: []model.DomainDiskSpec{{
			Device: "disk",
			Bus:    "virtio",
			Target: "vda",
			Pool:   "images",
			Volume: "demo-root",
			Format: "qcow2",
		}},
		Networks: []model.VMNICSpec{{
			Alias:   "net0",
			Network: "default",
			Model:   "virtio",
		}},
	})
	if err != nil {
		t.Fatalf("BuildDomain() error = %v", err)
	}

	for _, want := range []string{
		"<name>demo</name>",
		"<memory unit=\"MiB\">2048</memory>",
		"<vcpu>2</vcpu>",
		"<source pool=\"images\" volume=\"demo-root\"></source>",
		"<target dev=\"vda\" bus=\"virtio\"></target>",
		"<alias name=\"net0\"></alias>",
		"<source network=\"default\"></source>",
		"<graphics type=\"vnc\" autoport=\"yes\" listen=\"127.0.0.1\"></graphics>",
	} {
		if !strings.Contains(xml, want) {
			t.Fatalf("BuildDomain() XML missing %q\n%s", want, xml)
		}
	}
}

func TestBuildInstallDomainAddsCDROMAndBootOrder(t *testing.T) {
	t.Parallel()

	xml, err := BuildInstallDomain(model.VMInstallSpec{
		Name:        "installer",
		MemoryMiB:   4096,
		VCPU:        4,
		ISOPath:     "/var/lib/libvirt/boot/os.iso",
		Networks:    []model.VMNICSpec{{Network: "default"}},
		Graphics:    []model.GraphicsSpec{{Type: "spice", Listen: "0.0.0.0", AutoPort: true}},
		DiskFormat:  "raw",
		Description: "boot from iso",
	}, model.DomainDiskSpec{
		Device: "disk",
		Bus:    "virtio",
		Target: "vda",
		Pool:   "images",
		Volume: "installer-root",
		Format: "raw",
	})
	if err != nil {
		t.Fatalf("BuildInstallDomain() error = %v", err)
	}

	for _, want := range []string{
		"<boot dev=\"cdrom\"></boot>",
		"<boot dev=\"hd\"></boot>",
		"<disk type=\"file\" device=\"cdrom\">",
		"<source file=\"/var/lib/libvirt/boot/os.iso\"></source>",
		"<graphics type=\"spice\" autoport=\"yes\" listen=\"0.0.0.0\"></graphics>",
	} {
		if !strings.Contains(xml, want) {
			t.Fatalf("BuildInstallDomain() XML missing %q\n%s", want, xml)
		}
	}
}

func TestBuildStorageNetworkAndInterfaceXML(t *testing.T) {
	t.Parallel()

	poolXML, err := BuildPool(model.PoolDefineSpec{
		Name:       "tank",
		Type:       "zfs",
		SourceName: "tank/vm",
		TargetPath: "/tank/vm",
	})
	if err != nil {
		t.Fatalf("BuildPool() error = %v", err)
	}
	if !strings.Contains(poolXML, "<pool type=\"zfs\">") || !strings.Contains(poolXML, "<name>tank</name>") {
		t.Fatalf("BuildPool() XML unexpected\n%s", poolXML)
	}

	volumeXML, err := BuildVolume(model.VolumeCreateSpec{
		Name:          "demo-root",
		CapacityBytes: 10 << 30,
		Format:        "qcow2",
	})
	if err != nil {
		t.Fatalf("BuildVolume() error = %v", err)
	}
	if !strings.Contains(volumeXML, "<format type=\"qcow2\"></format>") {
		t.Fatalf("BuildVolume() XML unexpected\n%s", volumeXML)
	}

	networkXML, err := BuildNetwork(model.NetworkDefineSpec{
		Name:        "isolated",
		Bridge:      "virbr10",
		ForwardMode: "nat",
		IPv4CIDR:    "192.168.100.1/24",
		DHCPStart:   "192.168.100.10",
		DHCPEnd:     "192.168.100.200",
	})
	if err != nil {
		t.Fatalf("BuildNetwork() error = %v", err)
	}
	for _, want := range []string{
		"<bridge name=\"virbr10\" stp=\"on\" delay=\"0\"></bridge>",
		"<forward mode=\"nat\"></forward>",
		"<ip address=\"192.168.100.1\" prefix=\"24\">",
		"<range start=\"192.168.100.10\" end=\"192.168.100.200\"></range>",
	} {
		if !strings.Contains(networkXML, want) {
			t.Fatalf("BuildNetwork() XML missing %q\n%s", want, networkXML)
		}
	}

	ifaceXML, err := BuildInterface(model.InterfaceDefineSpec{
		Name:          "br-test0",
		Type:          "bridge",
		MAC:           "52:54:00:12:34:56",
		StartMode:     "onboot",
		MTU:           9000,
		BridgeMembers: []string{"eth0", "eth1"},
		Protocols: []model.InterfaceProtocolSpec{{
			Family:  "ipv4",
			Address: "10.0.0.10",
			Prefix:  24,
			Gateway: "10.0.0.1",
		}},
	})
	if err != nil {
		t.Fatalf("BuildInterface() error = %v", err)
	}
	for _, want := range []string{
		"<mac address=\"52:54:00:12:34:56\"></mac>",
		"<mtu size=\"9000\"></mtu>",
		"<start mode=\"onboot\"></start>",
		"<interface type=\"ethernet\" name=\"eth0\"></interface>",
		"<interface type=\"ethernet\" name=\"eth1\"></interface>",
		"<ip address=\"10.0.0.10\" prefix=\"24\"></ip>",
		"<route gateway=\"10.0.0.1\"></route>",
	} {
		if !strings.Contains(ifaceXML, want) {
			t.Fatalf("BuildInterface() XML missing %q\n%s", want, ifaceXML)
		}
	}
}
