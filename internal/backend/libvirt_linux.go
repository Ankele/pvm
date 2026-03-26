//go:build linux && cgo

package backend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ankele/pvm/internal/model"
	"github.com/ankele/pvm/internal/xmlbuild"
	"github.com/ankele/pvm/internal/zfs"
	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

type libvirtBackend struct {
	opts Options
	conn *libvirt.Connect
	zfs  *zfs.Client
}

func New(_ context.Context, opts Options) (Backend, error) {
	conn, err := openConnection(opts)
	if err != nil {
		return nil, err
	}
	return &libvirtBackend{
		opts: opts,
		conn: conn,
		zfs:  zfs.New("zfs"),
	}, nil
}

func (b *libvirtBackend) Close() error {
	if b.conn == nil {
		return nil
	}
	return b.conn.Close()
}

func (b *libvirtBackend) Info(ctx context.Context) (*model.SystemView, error) {
	return &model.SystemView{
		Backend:         "libvirt",
		URI:             b.opts.URI,
		Platform:        runtime.GOOS + "/" + runtime.GOARCH,
		SupportsLibvirt: true,
		SupportsZFS:     b.zfs.Available(ctx),
	}, nil
}

func (b *libvirtBackend) ListVMs(ctx context.Context) ([]model.VMView, error) {
	domains, err := b.conn.ListAllDomains(0)
	if err != nil {
		return nil, wrapLibvirt(err, "list domains")
	}
	items := make([]model.VMView, 0, len(domains))
	for i := range domains {
		dom := &domains[i]
		view, viewErr := b.domainView(ctx, dom)
		_ = dom.Free()
		if viewErr != nil {
			return nil, viewErr
		}
		items = append(items, *view)
	}
	return items, nil
}

func (b *libvirtBackend) GetVM(ctx context.Context, name string) (*model.VMView, error) {
	dom, err := b.conn.LookupDomainByName(name)
	if err != nil {
		return nil, wrapLookup(err, "domain", name)
	}
	defer dom.Free()
	return b.domainView(ctx, dom)
}

func (b *libvirtBackend) DefineVM(_ context.Context, spec model.DomainSpec) (*model.VMView, error) {
	xml, err := xmlbuild.BuildDomain(spec)
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build domain xml")
	}
	dom, err := b.conn.DomainDefineXML(xml)
	if err != nil {
		return nil, wrapLibvirt(err, "define domain")
	}
	defer dom.Free()
	if spec.Autostart {
		if err := dom.SetAutostart(true); err != nil {
			return nil, wrapLibvirt(err, "set domain autostart")
		}
	}
	return b.domainView(context.Background(), dom)
}

func (b *libvirtBackend) InstallVM(ctx context.Context, spec model.VMInstallSpec) (*model.VMView, error) {
	rootVol, err := b.ensureVolume(ctx, model.VolumeCreateSpec{
		Pool:          spec.Pool,
		Name:          spec.DiskName,
		CapacityBytes: gibToBytes(spec.DiskSizeGiB),
		Format:        defaultString(spec.DiskFormat, "raw"),
	})
	if err != nil {
		return nil, err
	}
	rootDisk := model.DomainDiskSpec{
		Device: "disk",
		Bus:    defaultString(spec.TargetBus, "virtio"),
		Target: "vda",
		Format: defaultString(spec.DiskFormat, "raw"),
		Pool:   rootVol.Pool,
		Volume: rootVol.Name,
	}
	xml, err := xmlbuild.BuildInstallDomain(spec, rootDisk)
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build install domain xml")
	}
	dom, err := b.conn.DomainDefineXML(xml)
	if err != nil {
		return nil, wrapLibvirt(err, "define install domain")
	}
	defer dom.Free()
	if spec.Autostart {
		if err := dom.SetAutostart(true); err != nil {
			return nil, wrapLibvirt(err, "set domain autostart")
		}
	}
	if err := dom.Create(); err != nil {
		return nil, wrapLibvirt(err, "start install domain")
	}
	return b.domainView(ctx, dom)
}

func (b *libvirtBackend) LaunchVM(ctx context.Context, spec model.VMLaunchSpec) (*model.VMView, error) {
	rootDisk, err := b.prepareLaunchDisk(ctx, spec)
	if err != nil {
		return nil, err
	}
	domainSpec := model.DomainSpec{
		Name:        spec.Name,
		Description: spec.Description,
		MemoryMiB:   spec.MemoryMiB,
		VCPU:        spec.VCPU,
		Machine:     spec.Machine,
		Arch:        spec.Arch,
		Autostart:   spec.Autostart,
		Disks:       []model.DomainDiskSpec{rootDisk},
		Networks:    spec.Networks,
		Graphics:    spec.Graphics,
	}
	xml, err := xmlbuild.BuildDomain(domainSpec)
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build launch domain xml")
	}
	dom, err := b.conn.DomainDefineXML(xml)
	if err != nil {
		return nil, wrapLibvirt(err, "define launch domain")
	}
	defer dom.Free()
	if spec.Autostart {
		if err := dom.SetAutostart(true); err != nil {
			return nil, wrapLibvirt(err, "set domain autostart")
		}
	}
	if err := dom.Create(); err != nil {
		return nil, wrapLibvirt(err, "start launched domain")
	}
	view, err := b.domainView(ctx, dom)
	if err != nil {
		return nil, err
	}
	view.SourceMode = string(spec.Mode)
	view.SharedSource = spec.Mode == model.LaunchModeDirect
	return view, nil
}

func (b *libvirtBackend) StartVM(ctx context.Context, name string) (*model.VMView, error) {
	dom, err := b.conn.LookupDomainByName(name)
	if err != nil {
		return nil, wrapLookup(err, "domain", name)
	}
	defer dom.Free()
	if err := dom.Create(); err != nil {
		return nil, wrapLibvirt(err, "start domain")
	}
	return b.domainView(ctx, dom)
}

func (b *libvirtBackend) ShutdownVM(ctx context.Context, name string) (*model.VMView, error) {
	return b.withDomainAction(ctx, name, func(dom *libvirt.Domain) error { return dom.Shutdown() })
}

func (b *libvirtBackend) DestroyVM(ctx context.Context, name string) (*model.VMView, error) {
	return b.withDomainAction(ctx, name, func(dom *libvirt.Domain) error { return dom.Destroy() })
}

func (b *libvirtBackend) RebootVM(ctx context.Context, name string) (*model.VMView, error) {
	return b.withDomainAction(ctx, name, func(dom *libvirt.Domain) error { return dom.Reboot(0) })
}

func (b *libvirtBackend) PauseVM(ctx context.Context, name string) (*model.VMView, error) {
	return b.withDomainAction(ctx, name, func(dom *libvirt.Domain) error { return dom.Suspend() })
}

func (b *libvirtBackend) ResumeVM(ctx context.Context, name string) (*model.VMView, error) {
	return b.withDomainAction(ctx, name, func(dom *libvirt.Domain) error { return dom.Resume() })
}

func (b *libvirtBackend) UndefineVM(_ context.Context, name string) (*model.ActionResult, error) {
	dom, err := b.conn.LookupDomainByName(name)
	if err != nil {
		return nil, wrapLookup(err, "domain", name)
	}
	defer dom.Free()
	if err := dom.UndefineFlags(libvirt.DOMAIN_UNDEFINE_SNAPSHOTS_METADATA | libvirt.DOMAIN_UNDEFINE_MANAGED_SAVE | libvirt.DOMAIN_UNDEFINE_NVRAM); err != nil {
		if err := dom.Undefine(); err != nil {
			return nil, wrapLibvirt(err, "undefine domain")
		}
	}
	return &model.ActionResult{Message: "domain undefined"}, nil
}

func (b *libvirtBackend) AddVMNIC(ctx context.Context, spec model.VMNICSpec) (*model.VMView, error) {
	dom, err := b.conn.LookupDomainByName(spec.VM)
	if err != nil {
		return nil, wrapLookup(err, "domain", spec.VM)
	}
	defer dom.Free()
	xml, err := interfaceDeviceXML(spec, 0)
	if err != nil {
		return nil, err
	}
	if err := dom.AttachDeviceFlags(xml, b.domainModifyFlags(dom)); err != nil {
		return nil, wrapLibvirt(err, "attach vm nic")
	}
	return b.domainView(ctx, dom)
}

func (b *libvirtBackend) UpdateVMNIC(ctx context.Context, spec model.VMNICUpdateSpec) (*model.VMView, error) {
	dom, err := b.conn.LookupDomainByName(spec.VM)
	if err != nil {
		return nil, wrapLookup(err, "domain", spec.VM)
	}
	defer dom.Free()
	xml, err := interfaceDeviceXML(spec.VMNICSpec, 0)
	if err != nil {
		return nil, err
	}
	if err := dom.UpdateDeviceFlags(xml, b.domainModifyFlags(dom)); err != nil {
		return nil, wrapLibvirt(err, "update vm nic")
	}
	return b.domainView(ctx, dom)
}

func (b *libvirtBackend) RemoveVMNIC(ctx context.Context, vmName, alias string) (*model.VMView, error) {
	dom, err := b.conn.LookupDomainByName(vmName)
	if err != nil {
		return nil, wrapLookup(err, "domain", vmName)
	}
	defer dom.Free()
	if err := dom.DetachDeviceAlias(alias, b.domainModifyFlags(dom)); err != nil {
		return nil, wrapLibvirt(err, "detach vm nic")
	}
	return b.domainView(ctx, dom)
}

func (b *libvirtBackend) GetVMGraphics(ctx context.Context, name string) (*model.GraphicsInfo, error) {
	vm, err := b.GetVM(ctx, name)
	if err != nil {
		return nil, err
	}
	return &model.GraphicsInfo{VM: vm.Name, Entries: vm.Graphics}, nil
}

func (b *libvirtBackend) CreateDomainSnapshot(_ context.Context, spec model.VMSnapshotCreateSpec) (*model.SnapshotView, error) {
	dom, err := b.conn.LookupDomainByName(spec.VM)
	if err != nil {
		return nil, wrapLookup(err, "domain", spec.VM)
	}
	defer dom.Free()
	snapshotXML, err := buildDomainSnapshotXML(spec)
	if err != nil {
		return nil, err
	}
	var flags libvirt.DomainSnapshotCreateFlags
	if spec.Quiesce {
		flags |= libvirt.DOMAIN_SNAPSHOT_CREATE_QUIESCE
	}
	if !spec.IncludeMemory {
		flags |= libvirt.DOMAIN_SNAPSHOT_CREATE_DISK_ONLY
	}
	snapshot, err := dom.CreateSnapshotXML(snapshotXML, flags)
	if err != nil {
		return nil, wrapLibvirt(err, "create domain snapshot")
	}
	defer snapshot.Free()
	return b.domainSnapshotView(&snapshot, spec.VM)
}

func (b *libvirtBackend) ListDomainSnapshots(_ context.Context, vmName string) ([]model.SnapshotView, error) {
	dom, err := b.conn.LookupDomainByName(vmName)
	if err != nil {
		return nil, wrapLookup(err, "domain", vmName)
	}
	defer dom.Free()
	snapshots, err := dom.ListAllSnapshots(0)
	if err != nil {
		return nil, wrapLibvirt(err, "list domain snapshots")
	}
	items := make([]model.SnapshotView, 0, len(snapshots))
	for i := range snapshots {
		snapshot := &snapshots[i]
		view, viewErr := b.domainSnapshotView(snapshot, vmName)
		_ = snapshot.Free()
		if viewErr != nil {
			return nil, viewErr
		}
		items = append(items, *view)
	}
	return items, nil
}

func (b *libvirtBackend) DeleteDomainSnapshot(_ context.Context, vmName, snapshotName string) (*model.ActionResult, error) {
	snapshot, err := b.lookupDomainSnapshot(vmName, snapshotName)
	if err != nil {
		return nil, err
	}
	defer snapshot.Free()
	if err := snapshot.Delete(0); err != nil {
		return nil, wrapLibvirt(err, "delete domain snapshot")
	}
	return &model.ActionResult{Message: "domain snapshot deleted"}, nil
}

func (b *libvirtBackend) RevertDomainSnapshot(_ context.Context, vmName, snapshotName string) (*model.SnapshotView, error) {
	snapshot, err := b.lookupDomainSnapshot(vmName, snapshotName)
	if err != nil {
		return nil, err
	}
	defer snapshot.Free()
	if err := snapshot.RevertToSnapshot(0); err != nil {
		return nil, wrapLibvirt(err, "revert domain snapshot")
	}
	return b.domainSnapshotView(snapshot, vmName)
}

func (b *libvirtBackend) ListPools(_ context.Context) ([]model.StoragePoolView, error) {
	pools, err := b.conn.ListAllStoragePools(0)
	if err != nil {
		return nil, wrapLibvirt(err, "list storage pools")
	}
	items := make([]model.StoragePoolView, 0, len(pools))
	for i := range pools {
		pool := &pools[i]
		view, viewErr := b.poolView(pool)
		_ = pool.Free()
		if viewErr != nil {
			return nil, viewErr
		}
		items = append(items, *view)
	}
	return items, nil
}

func (b *libvirtBackend) GetPool(_ context.Context, name string) (*model.StoragePoolView, error) {
	pool, err := b.conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", name)
	}
	defer pool.Free()
	return b.poolView(pool)
}

func (b *libvirtBackend) DefinePool(_ context.Context, spec model.PoolDefineSpec) (*model.StoragePoolView, error) {
	xml, err := xmlbuild.BuildPool(spec)
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build storage pool xml")
	}
	pool, err := b.conn.StoragePoolDefineXML(xml, 0)
	if err != nil {
		return nil, wrapLibvirt(err, "define storage pool")
	}
	defer pool.Free()
	if spec.Autostart {
		if err := pool.SetAutostart(true); err != nil {
			return nil, wrapLibvirt(err, "set pool autostart")
		}
	}
	return b.poolView(pool)
}

func (b *libvirtBackend) StartPool(_ context.Context, name string) (*model.StoragePoolView, error) {
	pool, err := b.conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", name)
	}
	defer pool.Free()
	if err := pool.Create(0); err != nil {
		return nil, wrapLibvirt(err, "start storage pool")
	}
	return b.poolView(pool)
}

func (b *libvirtBackend) DestroyPool(_ context.Context, name string) (*model.StoragePoolView, error) {
	pool, err := b.conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", name)
	}
	defer pool.Free()
	if err := pool.Destroy(); err != nil {
		return nil, wrapLibvirt(err, "destroy storage pool")
	}
	return b.poolView(pool)
}

func (b *libvirtBackend) SetPoolAutostart(_ context.Context, name string, autostart bool) (*model.StoragePoolView, error) {
	pool, err := b.conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", name)
	}
	defer pool.Free()
	if err := pool.SetAutostart(autostart); err != nil {
		return nil, wrapLibvirt(err, "set pool autostart")
	}
	return b.poolView(pool)
}

func (b *libvirtBackend) UndefinePool(_ context.Context, name string) (*model.ActionResult, error) {
	pool, err := b.conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", name)
	}
	defer pool.Free()
	if err := pool.Undefine(); err != nil {
		return nil, wrapLibvirt(err, "undefine storage pool")
	}
	return &model.ActionResult{Message: "storage pool undefined"}, nil
}

func (b *libvirtBackend) RefreshPool(_ context.Context, name string) (*model.StoragePoolView, error) {
	pool, err := b.conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", name)
	}
	defer pool.Free()
	if err := pool.Refresh(0); err != nil {
		return nil, wrapLibvirt(err, "refresh storage pool")
	}
	return b.poolView(pool)
}

func (b *libvirtBackend) ListVolumes(_ context.Context, poolName string) ([]model.VolumeView, error) {
	pool, err := b.conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", poolName)
	}
	defer pool.Free()
	volumes, err := pool.ListAllStorageVolumes(0)
	if err != nil {
		return nil, wrapLibvirt(err, "list storage volumes")
	}
	items := make([]model.VolumeView, 0, len(volumes))
	for i := range volumes {
		vol := &volumes[i]
		view, viewErr := b.volumeView(poolName, vol)
		_ = vol.Free()
		if viewErr != nil {
			return nil, viewErr
		}
		items = append(items, *view)
	}
	return items, nil
}

func (b *libvirtBackend) GetVolume(_ context.Context, poolName, name string) (*model.VolumeView, error) {
	vol, err := b.lookupVolume(poolName, name)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	return b.volumeView(poolName, vol)
}

func (b *libvirtBackend) CreateVolume(ctx context.Context, spec model.VolumeCreateSpec) (*model.VolumeView, error) {
	return b.ensureVolume(ctx, spec)
}

func (b *libvirtBackend) DeleteVolume(ctx context.Context, poolName, name string) (*model.ActionResult, error) {
	vol, err := b.lookupVolume(poolName, name)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	path, _ := vol.GetPath()
	attached, err := b.volumeAttachedDomains(ctx, poolName, name, path)
	if err != nil {
		return nil, err
	}
	if len(attached) > 0 {
		return nil, model.Errorf(model.ErrPrecondition, "volume is attached to domains: %s", strings.Join(attached, ", "))
	}
	if err := vol.Delete(libvirt.STORAGE_VOL_DELETE_NORMAL); err != nil {
		return nil, wrapLibvirt(err, "delete storage volume")
	}
	return &model.ActionResult{Message: "storage volume deleted"}, nil
}

func (b *libvirtBackend) ResizeVolume(_ context.Context, spec model.VolumeResizeSpec) (*model.VolumeView, error) {
	vol, err := b.lookupVolume(spec.Pool, spec.Name)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	if err := vol.Resize(spec.CapacityBytes, 0); err != nil {
		return nil, wrapLibvirt(err, "resize storage volume")
	}
	return b.volumeView(spec.Pool, vol)
}

func (b *libvirtBackend) CreateVolumeSnapshot(ctx context.Context, spec model.VolumeSnapshotCreateSpec) (*model.SnapshotView, error) {
	vol, err := b.lookupVolume(spec.Pool, spec.Volume)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	view, err := b.volumeView(spec.Pool, vol)
	if err != nil {
		return nil, err
	}
	dataset, err := zfs.DatasetForPath(view.Path)
	if err != nil {
		return nil, err
	}
	resume, err := b.suspendAttachedDomains(ctx, spec.Pool, spec.Volume, view.Path)
	if err != nil {
		return nil, err
	}
	defer resume()
	if err := b.zfs.CreateSnapshot(ctx, dataset, spec.Name); err != nil {
		return nil, err
	}
	_, _ = b.RefreshPool(ctx, spec.Pool)
	return &model.SnapshotView{
		Kind:      model.SnapshotKindVolume,
		Name:      spec.Name,
		Pool:      spec.Pool,
		Volume:    spec.Volume,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (b *libvirtBackend) ListVolumeSnapshots(ctx context.Context, poolName, volumeName string) ([]model.SnapshotView, error) {
	vol, err := b.lookupVolume(poolName, volumeName)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	view, err := b.volumeView(poolName, vol)
	if err != nil {
		return nil, err
	}
	dataset, err := zfs.DatasetForPath(view.Path)
	if err != nil {
		return nil, err
	}
	snapshots, err := b.zfs.ListSnapshots(ctx, dataset)
	if err != nil {
		return nil, err
	}
	items := make([]model.SnapshotView, 0, len(snapshots))
	for _, snap := range snapshots {
		items = append(items, model.SnapshotView{
			Kind:      model.SnapshotKindVolume,
			Name:      snap.Name,
			Pool:      poolName,
			Volume:    volumeName,
			CreatedAt: snap.CreatedAt,
		})
	}
	return items, nil
}

func (b *libvirtBackend) DeleteVolumeSnapshot(ctx context.Context, poolName, volumeName, snapshotName string) (*model.ActionResult, error) {
	vol, err := b.lookupVolume(poolName, volumeName)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	view, err := b.volumeView(poolName, vol)
	if err != nil {
		return nil, err
	}
	dataset, err := zfs.DatasetForPath(view.Path)
	if err != nil {
		return nil, err
	}
	if err := b.zfs.DestroySnapshot(ctx, dataset, snapshotName); err != nil {
		return nil, err
	}
	_, _ = b.RefreshPool(ctx, poolName)
	return &model.ActionResult{Message: "volume snapshot deleted"}, nil
}

func (b *libvirtBackend) RollbackVolumeSnapshot(ctx context.Context, poolName, volumeName, snapshotName string) (*model.SnapshotView, error) {
	vol, err := b.lookupVolume(poolName, volumeName)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	view, err := b.volumeView(poolName, vol)
	if err != nil {
		return nil, err
	}
	dataset, err := zfs.DatasetForPath(view.Path)
	if err != nil {
		return nil, err
	}
	resume, err := b.suspendAttachedDomains(ctx, poolName, volumeName, view.Path)
	if err != nil {
		return nil, err
	}
	defer resume()
	if err := b.zfs.RollbackSnapshot(ctx, dataset, snapshotName); err != nil {
		return nil, err
	}
	_, _ = b.RefreshPool(ctx, poolName)
	return &model.SnapshotView{
		Kind:      model.SnapshotKindVolume,
		Name:      snapshotName,
		Pool:      poolName,
		Volume:    volumeName,
		CreatedAt: time.Now().UTC(),
		State:     "rolled_back",
	}, nil
}

func (b *libvirtBackend) ListNetworks(_ context.Context) ([]model.NetworkView, error) {
	networks, err := b.conn.ListAllNetworks(0)
	if err != nil {
		return nil, wrapLibvirt(err, "list networks")
	}
	items := make([]model.NetworkView, 0, len(networks))
	for i := range networks {
		netw := &networks[i]
		view, viewErr := b.networkView(netw)
		_ = netw.Free()
		if viewErr != nil {
			return nil, viewErr
		}
		items = append(items, *view)
	}
	return items, nil
}

func (b *libvirtBackend) GetNetwork(_ context.Context, name string) (*model.NetworkView, error) {
	netw, err := b.conn.LookupNetworkByName(name)
	if err != nil {
		return nil, wrapLookup(err, "network", name)
	}
	defer netw.Free()
	return b.networkView(netw)
}

func (b *libvirtBackend) DefineNetwork(_ context.Context, spec model.NetworkDefineSpec) (*model.NetworkView, error) {
	xml, err := xmlbuild.BuildNetwork(spec)
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build network xml")
	}
	netw, err := b.conn.NetworkDefineXML(xml)
	if err != nil {
		return nil, wrapLibvirt(err, "define network")
	}
	defer netw.Free()
	if spec.Autostart {
		if err := netw.SetAutostart(true); err != nil {
			return nil, wrapLibvirt(err, "set network autostart")
		}
	}
	return b.networkView(netw)
}

func (b *libvirtBackend) StartNetwork(_ context.Context, name string) (*model.NetworkView, error) {
	netw, err := b.conn.LookupNetworkByName(name)
	if err != nil {
		return nil, wrapLookup(err, "network", name)
	}
	defer netw.Free()
	if err := netw.Create(); err != nil {
		return nil, wrapLibvirt(err, "start network")
	}
	return b.networkView(netw)
}

func (b *libvirtBackend) DestroyNetwork(_ context.Context, name string) (*model.NetworkView, error) {
	netw, err := b.conn.LookupNetworkByName(name)
	if err != nil {
		return nil, wrapLookup(err, "network", name)
	}
	defer netw.Free()
	if err := netw.Destroy(); err != nil {
		return nil, wrapLibvirt(err, "destroy network")
	}
	return b.networkView(netw)
}

func (b *libvirtBackend) SetNetworkAutostart(_ context.Context, name string, autostart bool) (*model.NetworkView, error) {
	netw, err := b.conn.LookupNetworkByName(name)
	if err != nil {
		return nil, wrapLookup(err, "network", name)
	}
	defer netw.Free()
	if err := netw.SetAutostart(autostart); err != nil {
		return nil, wrapLibvirt(err, "set network autostart")
	}
	return b.networkView(netw)
}

func (b *libvirtBackend) UndefineNetwork(_ context.Context, name string) (*model.ActionResult, error) {
	netw, err := b.conn.LookupNetworkByName(name)
	if err != nil {
		return nil, wrapLookup(err, "network", name)
	}
	defer netw.Free()
	if err := netw.Undefine(); err != nil {
		return nil, wrapLibvirt(err, "undefine network")
	}
	return &model.ActionResult{Message: "network undefined"}, nil
}

func (b *libvirtBackend) ListInterfaces(_ context.Context) ([]model.InterfaceView, error) {
	ifaces, err := b.conn.ListAllInterfaces(0)
	if err != nil {
		return nil, wrapLibvirt(err, "list interfaces")
	}
	items := make([]model.InterfaceView, 0, len(ifaces))
	for i := range ifaces {
		iface := &ifaces[i]
		view, viewErr := b.interfaceView(iface)
		_ = iface.Free()
		if viewErr != nil {
			return nil, viewErr
		}
		items = append(items, *view)
	}
	return items, nil
}

func (b *libvirtBackend) GetInterface(_ context.Context, name string) (*model.InterfaceView, error) {
	iface, err := b.conn.LookupInterfaceByName(name)
	if err != nil {
		return nil, wrapLookup(err, "interface", name)
	}
	defer iface.Free()
	return b.interfaceView(iface)
}

func (b *libvirtBackend) DefineInterface(_ context.Context, spec model.InterfaceDefineSpec) (*model.InterfaceView, error) {
	xml, err := xmlbuild.BuildInterface(spec)
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build interface xml")
	}
	iface, err := b.conn.InterfaceDefineXML(xml, 0)
	if err != nil {
		return nil, wrapLibvirt(err, "define interface")
	}
	defer iface.Free()
	return b.interfaceView(iface)
}

func (b *libvirtBackend) StartInterface(_ context.Context, name string) (*model.InterfaceView, error) {
	iface, err := b.conn.LookupInterfaceByName(name)
	if err != nil {
		return nil, wrapLookup(err, "interface", name)
	}
	defer iface.Free()
	if err := iface.Create(0); err != nil {
		return nil, wrapLibvirt(err, "start interface")
	}
	return b.interfaceView(iface)
}

func (b *libvirtBackend) DestroyInterface(_ context.Context, name string) (*model.InterfaceView, error) {
	iface, err := b.conn.LookupInterfaceByName(name)
	if err != nil {
		return nil, wrapLookup(err, "interface", name)
	}
	defer iface.Free()
	if err := iface.Destroy(0); err != nil {
		return nil, wrapLibvirt(err, "destroy interface")
	}
	return b.interfaceView(iface)
}

func (b *libvirtBackend) UndefineInterface(_ context.Context, name string) (*model.ActionResult, error) {
	iface, err := b.conn.LookupInterfaceByName(name)
	if err != nil {
		return nil, wrapLookup(err, "interface", name)
	}
	defer iface.Free()
	if err := iface.Undefine(); err != nil {
		return nil, wrapLibvirt(err, "undefine interface")
	}
	return &model.ActionResult{Message: "interface undefined"}, nil
}

func (b *libvirtBackend) withDomainAction(ctx context.Context, name string, action func(*libvirt.Domain) error) (*model.VMView, error) {
	dom, err := b.conn.LookupDomainByName(name)
	if err != nil {
		return nil, wrapLookup(err, "domain", name)
	}
	defer dom.Free()
	if err := action(dom); err != nil {
		return nil, wrapLibvirt(err, "domain action")
	}
	return b.domainView(ctx, dom)
}

func (b *libvirtBackend) domainView(_ context.Context, dom *libvirt.Domain) (*model.VMView, error) {
	name, err := dom.GetName()
	if err != nil {
		return nil, wrapLibvirt(err, "get domain name")
	}
	uuid, _ := dom.GetUUIDString()
	state, _, _ := dom.GetState()
	info, _ := dom.GetInfo()
	persistent, _ := dom.IsPersistent()
	autostart, _ := dom.GetAutostart()
	xmlDesc, err := dom.GetXMLDesc(0)
	if err != nil {
		return nil, wrapLibvirt(err, "get domain xml")
	}
	var xmlDomain libvirtxml.Domain
	if err := xmlDomain.Unmarshal(xmlDesc); err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "parse domain xml")
	}
	view := &model.VMView{
		Name:       name,
		UUID:       uuid,
		State:      mapDomainState(state),
		Persistent: persistent,
		Autostart:  autostart,
	}
	if info != nil {
		view.MemoryMiB = uint64(info.MaxMem) / 1024
		view.CurrentMemoryMiB = uint64(info.Memory) / 1024
		view.VCPU = uint32(info.NrVirtCpu)
		view.CPUTimeNanos = uint64(info.CpuTime)
	}
	if xmlDomain.Devices != nil {
		for _, disk := range xmlDomain.Devices.Disks {
			view.Disks = append(view.Disks, model.VMDiskView{
				Device:   disk.Device,
				Bus:      safeDiskBus(disk.Target),
				Target:   safeDiskTarget(disk.Target),
				Source:   sourceString(disk.Source),
				Format:   safeDiskFormat(disk.Driver),
				Pool:     sourcePool(disk.Source),
				Volume:   sourceVolume(disk.Source),
				ReadOnly: disk.ReadOnly != nil,
			})
		}
		for _, nic := range xmlDomain.Devices.Interfaces {
			view.NICs = append(view.NICs, model.VMNICView{
				Alias:   aliasName(nic.Alias),
				MAC:     macAddress(nic.MAC),
				Model:   interfaceModel(nic.Model),
				Network: interfaceNetwork(nic.Source),
				Bridge:  interfaceBridge(nic.Source),
			})
		}
		for _, graphics := range xmlDomain.Devices.Graphics {
			switch {
			case graphics.VNC != nil:
				view.Graphics = append(view.Graphics, model.VMGraphicsView{
					Type:      "vnc",
					Listen:    graphics.VNC.Listen,
					Port:      int32(graphics.VNC.Port),
					Websocket: int32(graphics.VNC.WebSocket),
				})
			case graphics.Spice != nil:
				view.Graphics = append(view.Graphics, model.VMGraphicsView{
					Type:    "spice",
					Listen:  graphics.Spice.Listen,
					Port:    int32(graphics.Spice.Port),
					TLSPort: int32(graphics.Spice.TLSPort),
				})
			}
		}
	}
	return view, nil
}

func (b *libvirtBackend) poolView(pool *libvirt.StoragePool) (*model.StoragePoolView, error) {
	name, err := pool.GetName()
	if err != nil {
		return nil, wrapLibvirt(err, "get pool name")
	}
	uuid, _ := pool.GetUUIDString()
	active, _ := pool.IsActive()
	autostart, _ := pool.GetAutostart()
	info, _ := pool.GetInfo()
	xmlDesc, err := pool.GetXMLDesc(0)
	if err != nil {
		return nil, wrapLibvirt(err, "get pool xml")
	}
	var xmlPool libvirtxml.StoragePool
	if err := xmlPool.Unmarshal(xmlDesc); err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "parse pool xml")
	}
	view := &model.StoragePoolView{
		Name:       name,
		UUID:       uuid,
		Type:       xmlPool.Type,
		Active:     active,
		Autostart:  autostart,
		TargetPath: safePoolTarget(xmlPool.Target),
		SourceName: safePoolSourceName(xmlPool.Source),
	}
	if info != nil {
		view.CapacityBytes = uint64(info.Capacity)
		view.AllocationBytes = uint64(info.Allocation)
		view.AvailableBytes = uint64(info.Available)
	}
	return view, nil
}

func (b *libvirtBackend) volumeView(poolName string, vol *libvirt.StorageVol) (*model.VolumeView, error) {
	name, err := vol.GetName()
	if err != nil {
		return nil, wrapLibvirt(err, "get volume name")
	}
	key, _ := vol.GetKey()
	path, _ := vol.GetPath()
	info, _ := vol.GetInfo()
	xmlDesc, err := vol.GetXMLDesc(0)
	if err != nil {
		return nil, wrapLibvirt(err, "get volume xml")
	}
	var xmlVol libvirtxml.StorageVolume
	if err := xmlVol.Unmarshal(xmlDesc); err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "parse volume xml")
	}
	view := &model.VolumeView{
		Pool:        poolName,
		Name:        name,
		Key:         key,
		Path:        path,
		Type:        xmlVol.Type,
		Format:      safeVolumeFormat(xmlVol.Target),
		BackingPath: safeVolumeBacking(xmlVol.BackingStore),
	}
	if info != nil {
		view.CapacityBytes = uint64(info.Capacity)
		view.AllocationBytes = uint64(info.Allocation)
	}
	return view, nil
}

func (b *libvirtBackend) networkView(netw *libvirt.Network) (*model.NetworkView, error) {
	name, err := netw.GetName()
	if err != nil {
		return nil, wrapLibvirt(err, "get network name")
	}
	uuid, _ := netw.GetUUIDString()
	active, _ := netw.IsActive()
	autostart, _ := netw.GetAutostart()
	xmlDesc, err := netw.GetXMLDesc(0)
	if err != nil {
		return nil, wrapLibvirt(err, "get network xml")
	}
	var xmlNet libvirtxml.Network
	if err := xmlNet.Unmarshal(xmlDesc); err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "parse network xml")
	}
	view := &model.NetworkView{
		Name:        name,
		UUID:        uuid,
		Bridge:      safeNetworkBridge(xmlNet.Bridge),
		ForwardMode: safeNetworkForward(xmlNet.Forward),
		Domain:      safeNetworkDomain(xmlNet.Domain),
		Active:      active,
		Autostart:   autostart,
	}
	if len(xmlNet.IPs) > 0 {
		view.IPv4CIDR = networkCIDR(xmlNet.IPs[0])
		if xmlNet.IPs[0].DHCP != nil && len(xmlNet.IPs[0].DHCP.Ranges) > 0 {
			view.DHCPStart = xmlNet.IPs[0].DHCP.Ranges[0].Start
			view.DHCPEnd = xmlNet.IPs[0].DHCP.Ranges[0].End
		}
	}
	return view, nil
}

func (b *libvirtBackend) interfaceView(iface *libvirt.Interface) (*model.InterfaceView, error) {
	name, err := iface.GetName()
	if err != nil {
		return nil, wrapLibvirt(err, "get interface name")
	}
	active, _ := iface.IsActive()
	mac, _ := iface.GetMACString()
	xmlDesc, err := iface.GetXMLDesc(0)
	if err != nil {
		return nil, wrapLibvirt(err, "get interface xml")
	}
	var xmlIface libvirtxml.Interface
	if err := xmlIface.Unmarshal(xmlDesc); err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "parse interface xml")
	}
	view := &model.InterfaceView{
		Name:      name,
		MAC:       mac,
		Type:      inferInterfaceType(&xmlIface),
		StartMode: safeInterfaceStart(xmlIface.Start),
		Active:    active,
	}
	if xmlIface.MTU != nil {
		view.MTU = uint32(xmlIface.MTU.Size)
	}
	if xmlIface.Bridge != nil {
		for _, member := range xmlIface.Bridge.Interfaces {
			view.BridgeMembers = append(view.BridgeMembers, member.Name)
		}
	}
	if xmlIface.Bond != nil {
		for _, member := range xmlIface.Bond.Interfaces {
			view.BridgeMembers = append(view.BridgeMembers, member.Name)
		}
	}
	for _, protocol := range xmlIface.Protocol {
		for _, ip := range protocol.IPs {
			view.Addresses = append(view.Addresses, fmt.Sprintf("%s/%d", ip.Address, ip.Prefix))
		}
	}
	return view, nil
}

func (b *libvirtBackend) ensureVolume(_ context.Context, spec model.VolumeCreateSpec) (*model.VolumeView, error) {
	pool, err := b.conn.LookupStoragePoolByName(spec.Pool)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", spec.Pool)
	}
	defer pool.Free()
	xml, err := xmlbuild.BuildVolume(spec)
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build storage volume xml")
	}
	vol, err := pool.StorageVolCreateXML(xml, 0)
	if err != nil {
		return nil, wrapLibvirt(err, "create storage volume")
	}
	defer vol.Free()
	return b.volumeView(spec.Pool, vol)
}

func (b *libvirtBackend) prepareLaunchDisk(ctx context.Context, spec model.VMLaunchSpec) (model.DomainDiskSpec, error) {
	switch spec.Mode {
	case model.LaunchModeDirect:
		if spec.ImagePath != "" {
			return model.DomainDiskSpec{
				Device: "disk",
				Bus:    "virtio",
				Target: "vda",
				Source: spec.ImagePath,
				Format: detectFormat(spec.ImagePath, spec.TargetFormat),
			}, nil
		}
		pool, vol, err := splitVolumeRef(spec.SourceVolume)
		if err != nil {
			return model.DomainDiskSpec{}, err
		}
		return model.DomainDiskSpec{
			Device: "disk",
			Bus:    "virtio",
			Target: "vda",
			Pool:   pool,
			Volume: vol,
		}, nil
	case model.LaunchModeClone:
		if spec.ImagePath != "" {
			view, err := b.importImageAsVolume(ctx, spec)
			if err != nil {
				return model.DomainDiskSpec{}, err
			}
			return model.DomainDiskSpec{
				Device: "disk",
				Bus:    "virtio",
				Target: "vda",
				Pool:   view.Pool,
				Volume: view.Name,
				Format: view.Format,
			}, nil
		}
		view, err := b.cloneVolume(ctx, spec)
		if err != nil {
			return model.DomainDiskSpec{}, err
		}
		return model.DomainDiskSpec{
			Device: "disk",
			Bus:    "virtio",
			Target: "vda",
			Pool:   view.Pool,
			Volume: view.Name,
			Format: view.Format,
		}, nil
	default:
		return model.DomainDiskSpec{}, model.Errorf(model.ErrInvalidArgument, "unsupported launch mode %q", spec.Mode)
	}
}

func (b *libvirtBackend) importImageAsVolume(ctx context.Context, spec model.VMLaunchSpec) (*model.VolumeView, error) {
	info, err := os.Stat(spec.ImagePath)
	if err != nil {
		return nil, model.Wrap(model.ErrNotFound, err, "stat image path")
	}
	volView, err := b.ensureVolume(ctx, model.VolumeCreateSpec{
		Pool:            spec.TargetPool,
		Name:            spec.TargetVolumeName,
		CapacityBytes:   uint64(info.Size()),
		AllocationBytes: uint64(info.Size()),
		Format:          detectFormat(spec.ImagePath, spec.TargetFormat),
	})
	if err != nil {
		return nil, err
	}
	vol, err := b.lookupVolume(volView.Pool, volView.Name)
	if err != nil {
		return nil, err
	}
	defer vol.Free()
	stream, err := b.conn.NewStream(0)
	if err != nil {
		return nil, wrapLibvirt(err, "create upload stream")
	}
	defer stream.Free()
	if err := vol.Upload(stream, 0, uint64(info.Size()), 0); err != nil {
		return nil, wrapLibvirt(err, "begin volume upload")
	}
	file, err := os.Open(spec.ImagePath)
	if err != nil {
		_ = stream.Abort()
		return nil, model.Wrap(model.ErrNotFound, err, "open source image")
	}
	defer file.Close()
	if err := copyToStream(file, stream); err != nil {
		_ = stream.Abort()
		return nil, err
	}
	if err := stream.Finish(); err != nil {
		return nil, wrapLibvirt(err, "finish volume upload")
	}
	return b.GetVolume(ctx, volView.Pool, volView.Name)
}

func (b *libvirtBackend) cloneVolume(ctx context.Context, spec model.VMLaunchSpec) (*model.VolumeView, error) {
	srcPoolName, srcVolName, err := splitVolumeRef(spec.SourceVolume)
	if err != nil {
		return nil, err
	}
	srcVol, err := b.lookupVolume(srcPoolName, srcVolName)
	if err != nil {
		return nil, err
	}
	defer srcVol.Free()
	srcInfo, err := srcVol.GetInfo()
	if err != nil {
		return nil, wrapLibvirt(err, "get source volume info")
	}
	dstPool, err := b.conn.LookupStoragePoolByName(spec.TargetPool)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", spec.TargetPool)
	}
	defer dstPool.Free()
	xml, err := xmlbuild.BuildVolume(model.VolumeCreateSpec{
		Pool:            spec.TargetPool,
		Name:            spec.TargetVolumeName,
		CapacityBytes:   uint64(srcInfo.Capacity),
		AllocationBytes: uint64(srcInfo.Allocation),
		Format:          defaultString(spec.TargetFormat, "raw"),
	})
	if err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "build cloned volume xml")
	}
	clone, err := dstPool.StorageVolCreateXMLFrom(xml, srcVol, 0)
	if err != nil {
		return nil, wrapLibvirt(err, "clone storage volume")
	}
	defer clone.Free()
	return b.GetVolume(ctx, spec.TargetPool, spec.TargetVolumeName)
}

func (b *libvirtBackend) lookupVolume(poolName, name string) (*libvirt.StorageVol, error) {
	pool, err := b.conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return nil, wrapLookup(err, "storage pool", poolName)
	}
	defer pool.Free()
	vol, err := pool.LookupStorageVolByName(name)
	if err != nil {
		return nil, wrapLookup(err, "storage volume", poolName+"/"+name)
	}
	return vol, nil
}

func (b *libvirtBackend) lookupDomainSnapshot(vmName, snapshotName string) (*libvirt.DomainSnapshot, error) {
	dom, err := b.conn.LookupDomainByName(vmName)
	if err != nil {
		return nil, wrapLookup(err, "domain", vmName)
	}
	defer dom.Free()
	snapshot, err := dom.SnapshotLookupByName(snapshotName, 0)
	if err != nil {
		return nil, wrapLookup(err, "snapshot", vmName+"/"+snapshotName)
	}
	return snapshot, nil
}

func (b *libvirtBackend) domainSnapshotView(snapshot *libvirt.DomainSnapshot, vmName string) (*model.SnapshotView, error) {
	xmlDesc, err := snapshot.GetXMLDesc(0)
	if err != nil {
		return nil, wrapLibvirt(err, "get domain snapshot xml")
	}
	var xmlSnap libvirtxml.DomainSnapshot
	if err := xmlSnap.Unmarshal(xmlDesc); err != nil {
		return nil, model.Wrap(model.ErrInternal, err, "parse domain snapshot xml")
	}
	createdAt := time.Time{}
	if xmlSnap.CreationTime != "" {
		if unix, convErr := strconv.ParseInt(xmlSnap.CreationTime, 10, 64); convErr == nil {
			createdAt = time.Unix(unix, 0).UTC()
		}
	}
	return &model.SnapshotView{
		Kind:        model.SnapshotKindDomain,
		Name:        xmlSnap.Name,
		Parent:      safeSnapshotParent(xmlSnap.Parent),
		Domain:      vmName,
		Description: xmlSnap.Description,
		CreatedAt:   createdAt,
		State:       xmlSnap.State,
	}, nil
}

func (b *libvirtBackend) volumeAttachedDomains(ctx context.Context, poolName, volumeName, path string) ([]string, error) {
	vms, err := b.ListVMs(ctx)
	if err != nil {
		return nil, err
	}
	var attached []string
	for _, vm := range vms {
		for _, disk := range vm.Disks {
			if (disk.Pool == poolName && disk.Volume == volumeName) || disk.Source == path {
				attached = append(attached, vm.Name)
				break
			}
		}
	}
	return attached, nil
}

func (b *libvirtBackend) suspendAttachedDomains(ctx context.Context, poolName, volumeName, path string) (func(), error) {
	names, err := b.volumeAttachedDomains(ctx, poolName, volumeName, path)
	if err != nil {
		return nil, err
	}
	resumed := make([]*libvirt.Domain, 0, len(names))
	for _, name := range names {
		dom, err := b.conn.LookupDomainByName(name)
		if err != nil {
			return nil, wrapLookup(err, "domain", name)
		}
		state, _, _ := dom.GetState()
		if mapDomainState(state) != model.VMStateRunning {
			_ = dom.Free()
			continue
		}
		if err := dom.Suspend(); err != nil {
			_ = dom.Free()
			return nil, wrapLibvirt(err, "suspend attached domain")
		}
		resumed = append(resumed, dom)
	}
	return func() {
		for _, dom := range resumed {
			_ = dom.Resume()
			_ = dom.Free()
		}
	}, nil
}

func openConnection(opts Options) (*libvirt.Connect, error) {
	if opts.Username == "" && opts.Password == "" {
		conn, err := libvirt.NewConnect(opts.URI)
		if err != nil {
			return nil, wrapLibvirt(err, "open libvirt connection")
		}
		return conn, nil
	}
	auth := &libvirt.ConnectAuth{
		CredType: []libvirt.ConnectCredentialType{
			libvirt.CRED_AUTHNAME,
			libvirt.CRED_PASSPHRASE,
		},
		Callback: func(creds []*libvirt.ConnectCredential) {
			for _, cred := range creds {
				switch cred.Type {
				case libvirt.CRED_AUTHNAME:
					cred.Result = opts.Username
				case libvirt.CRED_PASSPHRASE:
					cred.Result = opts.Password
				}
			}
		},
	}
	conn, err := libvirt.NewConnectWithAuth(opts.URI, auth, 0)
	if err != nil {
		return nil, wrapLibvirt(err, "open libvirt connection")
	}
	return conn, nil
}

func wrapLibvirt(err error, action string) error {
	if err == nil {
		return nil
	}
	var lvErr libvirt.Error
	if ok := errorAs(err, &lvErr); ok {
		switch lvErr.Code {
		case libvirt.ERR_NO_DOMAIN, libvirt.ERR_NO_STORAGE_POOL, libvirt.ERR_NO_STORAGE_VOL, libvirt.ERR_NO_NETWORK, libvirt.ERR_NO_INTERFACE, libvirt.ERR_NO_DOMAIN_SNAPSHOT:
			return model.Wrap(model.ErrNotFound, err, action)
		case libvirt.ERR_NO_SUPPORT:
			return model.Wrap(model.ErrUnsupported, err, action)
		case libvirt.ERR_OPERATION_INVALID, libvirt.ERR_OPERATION_FAILED:
			return model.Wrap(model.ErrPrecondition, err, action)
		}
	}
	return model.Wrap(model.ErrInternal, err, action)
}

func wrapLookup(err error, kind, name string) error {
	if err == nil {
		return nil
	}
	return model.Wrap(model.ErrNotFound, err, "%s %q not found", kind, name)
}

func errorAs(err error, target any) bool {
	return errors.As(err, target)
}

func buildDomainSnapshotXML(spec model.VMSnapshotCreateSpec) (string, error) {
	snapshot := libvirtxml.DomainSnapshot{
		Name:        spec.Name,
		Description: spec.Description,
	}
	if spec.IncludeMemory {
		snapshot.Memory = &libvirtxml.DomainSnapshotMemory{Snapshot: "internal"}
	} else {
		snapshot.Memory = &libvirtxml.DomainSnapshotMemory{Snapshot: "no"}
	}
	return snapshot.Marshal()
}

func interfaceDeviceXML(spec model.VMNICSpec, idx int) (string, error) {
	xml, err := xmlbuild.BuildDomain(model.DomainSpec{
		Name:      "tmp",
		MemoryMiB: 1,
		VCPU:      1,
		Disks: []model.DomainDiskSpec{{
			Device: "disk",
			Bus:    "virtio",
			Target: "vda",
			Source: "/tmp/placeholder",
			Format: "raw",
		}},
		Networks: []model.VMNICSpec{spec},
	})
	if err != nil {
		return "", err
	}
	var dom libvirtxml.Domain
	if err := dom.Unmarshal(xml); err != nil {
		return "", model.Wrap(model.ErrInternal, err, "parse interface helper xml")
	}
	if dom.Devices == nil || len(dom.Devices.Interfaces) == 0 {
		return "", model.Errorf(model.ErrInternal, "generated interface XML is empty")
	}
	return dom.Devices.Interfaces[0].Marshal()
}

func (b *libvirtBackend) domainModifyFlags(dom *libvirt.Domain) libvirt.DomainDeviceModifyFlags {
	active, err := dom.IsActive()
	if err == nil && active {
		return libvirt.DOMAIN_DEVICE_MODIFY_CONFIG | libvirt.DOMAIN_DEVICE_MODIFY_LIVE
	}
	return libvirt.DOMAIN_DEVICE_MODIFY_CONFIG
}

func copyToStream(r io.Reader, stream *libvirt.Stream) error {
	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			offset := 0
			for offset < n {
				written, sendErr := stream.Send(buf[offset:n])
				if sendErr != nil {
					return model.Wrap(model.ErrInternal, sendErr, "stream send failed")
				}
				offset += written
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return model.Wrap(model.ErrInternal, err, "read source image")
		}
	}
}

func splitVolumeRef(ref string) (string, string, error) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", model.Errorf(model.ErrInvalidArgument, "source volume must be in pool/name format")
	}
	return parts[0], parts[1], nil
}

func detectFormat(path, fallback string) string {
	if fallback != "" {
		return fallback
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".qcow2":
		return "qcow2"
	default:
		return "raw"
	}
}

func mapDomainState(state libvirt.DomainState) model.VMState {
	switch state {
	case libvirt.DOMAIN_RUNNING:
		return model.VMStateRunning
	case libvirt.DOMAIN_PAUSED:
		return model.VMStatePaused
	case libvirt.DOMAIN_SHUTOFF:
		return model.VMStateShutoff
	case libvirt.DOMAIN_CRASHED:
		return model.VMStateCrashed
	case libvirt.DOMAIN_BLOCKED:
		return model.VMStateBlocked
	case libvirt.DOMAIN_PMSUSPENDED:
		return model.VMStatePMSuspend
	default:
		return model.VMStateUnknown
	}
}

func gibToBytes(v uint64) uint64 { return v * 1024 * 1024 * 1024 }

func sourceString(source *libvirtxml.DomainDiskSource) string {
	switch {
	case source == nil:
		return ""
	case source.File != nil:
		return source.File.File
	case source.Volume != nil:
		return fmt.Sprintf("%s/%s", source.Volume.Pool, source.Volume.Volume)
	case source.Block != nil:
		return source.Block.Dev
	default:
		return ""
	}
}

func sourcePool(source *libvirtxml.DomainDiskSource) string {
	if source != nil && source.Volume != nil {
		return source.Volume.Pool
	}
	return ""
}

func sourceVolume(source *libvirtxml.DomainDiskSource) string {
	if source != nil && source.Volume != nil {
		return source.Volume.Volume
	}
	return ""
}

func safeDiskBus(target *libvirtxml.DomainDiskTarget) string {
	if target == nil {
		return ""
	}
	return target.Bus
}

func safeDiskTarget(target *libvirtxml.DomainDiskTarget) string {
	if target == nil {
		return ""
	}
	return target.Dev
}

func safeDiskFormat(driver *libvirtxml.DomainDiskDriver) string {
	if driver == nil {
		return ""
	}
	return driver.Type
}

func aliasName(alias *libvirtxml.DomainAlias) string {
	if alias == nil {
		return ""
	}
	return alias.Name
}

func macAddress(mac *libvirtxml.DomainInterfaceMAC) string {
	if mac == nil {
		return ""
	}
	return mac.Address
}

func interfaceModel(model *libvirtxml.DomainInterfaceModel) string {
	if model == nil {
		return ""
	}
	return model.Type
}

func interfaceNetwork(source *libvirtxml.DomainInterfaceSource) string {
	if source != nil && source.Network != nil {
		return source.Network.Network
	}
	return ""
}

func interfaceBridge(source *libvirtxml.DomainInterfaceSource) string {
	if source != nil {
		if source.Network != nil && source.Network.Bridge != "" {
			return source.Network.Bridge
		}
		if source.Bridge != nil {
			return source.Bridge.Bridge
		}
	}
	return ""
}

func safePoolTarget(target *libvirtxml.StoragePoolTarget) string {
	if target == nil {
		return ""
	}
	return target.Path
}

func safePoolSourceName(source *libvirtxml.StoragePoolSource) string {
	if source == nil {
		return ""
	}
	return source.Name
}

func safeVolumeFormat(target *libvirtxml.StorageVolumeTarget) string {
	if target == nil || target.Format == nil {
		return ""
	}
	return target.Format.Type
}

func safeVolumeBacking(backing *libvirtxml.StorageVolumeBackingStore) string {
	if backing == nil {
		return ""
	}
	return backing.Path
}

func safeNetworkBridge(bridge *libvirtxml.NetworkBridge) string {
	if bridge == nil {
		return ""
	}
	return bridge.Name
}

func safeNetworkForward(forward *libvirtxml.NetworkForward) string {
	if forward == nil {
		return ""
	}
	return forward.Mode
}

func safeNetworkDomain(domain *libvirtxml.NetworkDomain) string {
	if domain == nil {
		return ""
	}
	return domain.Name
}

func networkCIDR(ip libvirtxml.NetworkIP) string {
	if ip.Address == "" {
		return ""
	}
	if ip.Prefix > 0 {
		return fmt.Sprintf("%s/%d", ip.Address, ip.Prefix)
	}
	return ip.Address
}

func inferInterfaceType(iface *libvirtxml.Interface) string {
	switch {
	case iface == nil:
		return ""
	case iface.Bridge != nil:
		return "bridge"
	case iface.Bond != nil:
		return "bond"
	case iface.VLAN != nil:
		return "vlan"
	default:
		return "ethernet"
	}
}

func safeInterfaceStart(start *libvirtxml.InterfaceStart) string {
	if start == nil {
		return ""
	}
	return start.Mode
}

func safeSnapshotParent(parent *libvirtxml.DomainSnapshotParent) string {
	if parent == nil {
		return ""
	}
	return parent.Name
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
