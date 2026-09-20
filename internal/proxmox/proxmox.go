package proxmox

import (
	"context"
	"sort"
	"strings"
	"time"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
)

type Snapshot struct {
	Description  string
	NamePrefix   string
	TimeFormat   string
	AmountToKeep int
	State        bool
	DryRun       bool
}

func (snap Snapshot) AddAndRemove(ctx context.Context, c pve.ClientNew, vmr *pve.VmRef, state bool) (*Output, error) {
	hasRequiredFeature, err := c.Guest.HasFeature(ctx, *vmr, pve.GuestFeatureSnapshot)
	if err != nil {
		return nil, err
	}
	if !hasRequiredFeature {
		return nil, nil
	}
	o := Output{
		Create:    pve.SnapshotName(snap.NamePrefix + time.Now().Format(snap.TimeFormat)),
		GuestID:   vmr.VmId(),
		GuestType: vmr.GetVmType(),
	}
	if !snap.DryRun {
		switch o.GuestType {
		case pve.GuestLxc:
			err = c.Snapshot.CreateLxcNoCheck(ctx, *vmr, o.Create, snap.Description)
		case pve.GuestQemu:
			err = c.Snapshot.CreateQemuNoCheck(ctx, *vmr, o.Create, snap.Description, state)
		}
		if err != nil {
			return nil, err
		}
	}
	// skip this error, it will keep adding snapshots without being able to destroy old ones. due to the vm getting locked.
	// better to leave old snapshots to remove for the next run.
	oPtr := &o
	if snap.AmountToKeep > 0 {
		snap.ensureAmount(ctx, c.Snapshot, vmr, oPtr)
	}
	return oPtr, nil
}

func (snap Snapshot) EnsureAmount(ctx context.Context, c pve.SnapshotInterface, vmr *pve.VmRef) (*Output, error) {
	o := &Output{
		GuestID:   vmr.VmId(),
		GuestType: vmr.GetVmType(),
	}
	err := snap.ensureAmount(ctx, c, vmr, o)
	return o, err
}

func (snap Snapshot) ensureAmount(ctx context.Context, c pve.SnapshotInterface, vmr *pve.VmRef, o *Output) error {
	raw, err := c.List(ctx, *vmr)
	if err != nil {
		return err
	}
	toRemove := snap.toRemove(raw.AsArray())
	if snap.DryRun {
		o.Remove = toRemove
		return nil
	}
	return removeSnapshots(ctx, c, vmr, toRemove)
}

// returns a list of snapshot names to remove.
func (snap Snapshot) toRemove(allSnapshots []pve.RawSnapshotInfo) []pve.SnapshotName {
	type snapshot struct {
		name pve.SnapshotName
		Time time.Time
	}
	if len(allSnapshots) <= 1 { // we only have the current snapshot
		return nil
	}
	snapshotsToSort := make([]snapshot, 0, len(allSnapshots)-1)
	var snapName string
	for i := range allSnapshots {
		if allSnapshots[i].GetDescription() == snap.Description {
			snapshot := snapshot{
				name: allSnapshots[i].GetName(),
			}
			snapName = snapshot.name.String()
			if len(snapName) >= len(snap.TimeFormat) {
				nameWithoutTime := snapshot.name.String()[0 : len(snapName)-len(snap.TimeFormat)]
				if nameWithoutTime == snap.NamePrefix {
					var err error
					snapshot.Time, err = time.Parse(snap.TimeFormat, snapName[len(nameWithoutTime):])
					if err == nil {
						snapshotsToSort = append(snapshotsToSort, snapshot)
					}
				}
			}
		}
	}
	if len(snapshotsToSort) <= int(snap.AmountToKeep) {
		return nil
	}
	sort.SliceStable(snapshotsToSort, func(i, j int) bool {
		return snapshotsToSort[i].Time.Before(snapshotsToSort[j].Time)
	})
	toRemoveSnapshots := make([]pve.SnapshotName, len(snapshotsToSort)-int(snap.AmountToKeep))
	for i := range toRemoveSnapshots {
		toRemoveSnapshots[i] = snapshotsToSort[i].name
	}
	return toRemoveSnapshots
}

func removeSnapshots(ctx context.Context, c pve.SnapshotInterface, vmr *pve.VmRef, snapshots []pve.SnapshotName) (err error) {
	for i := range snapshots {
		if _, err = c.Delete(ctx, *vmr, snapshots[i]); err != nil {
			return
		}
	}
	return
}

func (snap Snapshot) ListSnapshots(ctx context.Context, c pve.SnapshotInterface, vmr *pve.VmRef) (uint, error) {
	raw, err := c.List(ctx, *vmr)
	if err != nil {
		return 0, err
	}
	info := raw.AsArray()
	var snapshots uint
	for i := range info {
		if info[i].GetDescription() != snap.Description {
			continue
		}
		if stringWithoutPrefix, ok := strings.CutPrefix(info[i].GetName().String(), snap.NamePrefix); ok {
			if _, err = time.Parse(snap.TimeFormat, stringWithoutPrefix); err == nil {
				snapshots++
			}
		}
	}
	return snapshots, nil
}

func ListNodes(ctx context.Context, c pve.NodeInterface) ([]pve.NodeName, error) {
	raw, err := c.List(ctx)
	if err != nil {
		return nil, err
	}
	raws := raw.AsArray()
	nodes := make([]pve.NodeName, len(raws))
	for i := range raws {
		nodes[i] = raws[i].GetName()
	}
	return nodes, nil
}

type Output struct {
	GuestID   pve.GuestID        `json:"guest_id"`
	GuestType pve.GuestType      `json:"guest_type,omitempty"`
	Create    pve.SnapshotName   `json:"snapshot_create,omitempty"`
	Remove    []pve.SnapshotName `json:"snapshots_prune,omitempty"`
}
