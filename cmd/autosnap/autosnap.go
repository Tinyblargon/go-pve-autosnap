package autosnap

import (
	"go-pve-autosnap/internal/filter"

	pxAPI "github.com/Telmate/proxmox-api-go/proxmox"
)

// convert a list of proxmox guests to a list of guests for the filter
// the difference is that the filter.Guest type is way smaller helping with memory usage
func convertToGuestList(guests []pxAPI.GuestResource) []filter.Guest {
	var numberOfTemplates int
	for _, e := range guests {
		if e.Template {
			numberOfTemplates++
		}
	}
	guestList := make([]filter.Guest, len(guests)-numberOfTemplates)
	var offset int
	for i, e := range guests {
		if e.Template {
			offset++
			continue
		}
		guestList[i-offset] = filter.Guest{
			Id:   e.Id,
			Name: e.Name,
			Node: e.Node,
			Pool: e.Pool,
			Tags: e.Tags,
			Type: e.Type,
		}
	}
	return guestList
}
