package filter

import (
	"slices"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
)

func filterAll() FilterStep {
	return func(gr *pve.GuestResource) bool {
		return true
	}
}

type idRange struct {
	min pve.GuestID
	max pve.GuestID
}

func filterID(e []pve.GuestID, r []idRange) FilterStep {
	return func(gr *pve.GuestResource) bool {
		contains := slices.Contains(e, gr.ID)
		if contains {
			return true
		}
		for i := range r {
			if gr.ID >= r[i].min && gr.ID <= r[i].max {
				return true
			}
		}
		return false
	}
}

func filterName(args []pve.GuestName) FilterStep {
	return func(gr *pve.GuestResource) bool {
		return slices.Contains(args, gr.Name)
	}
}

func filterNode(args []pve.NodeName) FilterStep {
	return func(gr *pve.GuestResource) bool {
		return slices.Contains(args, gr.Node)
	}
}

func filterPool(args []pve.PoolName) FilterStep {
	return func(gr *pve.GuestResource) bool {
		return slices.Contains(args, gr.Pool)
	}
}

func filterTag(args []pve.Tag) FilterStep {
	return func(gr *pve.GuestResource) bool {
		for i := range args {
			if slices.Contains(gr.Tags, args[i]) {
				return true
			}
		}
		return false
	}
}
