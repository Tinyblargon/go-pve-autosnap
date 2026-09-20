package tracker

import (
	"sync"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
)

type GuestSummary struct {
	mutex         sync.Mutex
	guestsLxc     uint
	guestsQemu    uint
	snapshotsLxc  uint
	snapshotsQemu uint
}

func (s *GuestSummary) Add(kind pve.GuestType, snapshots uint) {
	switch kind { // Locking inside the switch to ensure absolute minimal locked time.
	case pve.GuestLxc:
		s.mutex.Lock()
		s.guestsLxc++
		s.snapshotsLxc += snapshots
		s.mutex.Unlock()
	case pve.GuestQemu:
		s.mutex.Lock()
		s.guestsQemu++
		s.snapshotsQemu += snapshots
		s.mutex.Unlock()
	}
}

func (s *GuestSummary) Return() (guestsLxc, guestsQemu, snapshotsLxc, snapshotsQemu uint) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.guestsLxc, s.guestsQemu, s.snapshotsLxc, s.snapshotsQemu
}
