package xmlbuild

import (
	"github.com/ankele/pvm/internal/model"
	"libvirt.org/go/libvirtxml"
)

func BuildPool(spec model.PoolDefineSpec) (string, error) {
	pool := libvirtxml.StoragePool{
		Type: spec.Type,
		Name: spec.Name,
		Source: &libvirtxml.StoragePoolSource{
			Name: spec.SourceName,
		},
	}
	if spec.TargetPath != "" {
		pool.Target = &libvirtxml.StoragePoolTarget{Path: spec.TargetPath}
	}
	return pool.Marshal()
}

func BuildVolume(spec model.VolumeCreateSpec) (string, error) {
	vol := libvirtxml.StorageVolume{
		Name: spec.Name,
		Type: spec.Type,
		Capacity: &libvirtxml.StorageVolumeSize{
			Unit:  "bytes",
			Value: spec.CapacityBytes,
		},
		Allocation: &libvirtxml.StorageVolumeSize{
			Unit:  "bytes",
			Value: spec.AllocationBytes,
		},
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{Type: defaultString(spec.Format, "raw")},
		},
	}
	return vol.Marshal()
}
