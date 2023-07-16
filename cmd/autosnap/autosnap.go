package autosnap

import (
	"fmt"
	"go-pve-autosnap/internal/filter"

	pxAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"golang.org/x/exp/slices"
)

// check if a guest is already in the list
func completed(guest filter.Guest, guests []guestMinimal) bool {
	for i := range guests {
		if guest.Id == guests[i].Id && guest.Type == guests[i].Type {
			return true
		}
	}
	return false
}

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

// run the provided function on all guests that match the filter.
func Execute(client *pxAPI.Client, filterObj *filter.FilterSteps, function func(client *pxAPI.Client, vmr *pxAPI.VmRef) error) error {
	if filterObj == nil {
		return fmt.Errorf("filter is nil")
	}
	trackedGuests := guests{
		Filtered:  make(map[uint]guestNoID),
		Completed: make([]guestMinimal, 0),
	}

	var subroutineSucceeded bool
	for !subroutineSucceeded {
		guestList, err := newGuestList(client)
		if err != nil {
			return err
		}
		subroutineSucceeded = executeSubroutine(guestList, &trackedGuests, client, filterObj, function)
	}
	return nil
}

// subroutine of Execute to extract the testable code.
func executeSubroutine(guestList []filter.Guest, trackedGuests *guests, client *pxAPI.Client, filterObj *filter.FilterSteps, function func(client *pxAPI.Client, vmr *pxAPI.VmRef) error) bool {
	for _, e := range guestList {
		// check if a snapshot has been made of the guest during this run.
		if completed(e, trackedGuests.Completed) {
			continue
		}
		// check if guest has been filtered this run.
		// faster than running the filter on every time.
		if filtered(e, trackedGuests.Filtered) {
			continue
		}
		if filterObj.Apply(e) {
			vmr := pxAPI.NewVmRef(int(e.Id))
			vmr.SetNode(e.Node)
			vmr.SetVmType(string(e.Type))
			err := function(client, vmr)
			if err != nil {
				// if err guest list is out of date, obtain a new one.
				return false
			}
			trackedGuests.Completed = append(trackedGuests.Completed, guestMinimal{Id: e.Id, Type: e.Type})
			continue
		}
		slices.Sort[string](e.Tags)
		trackedGuests.Filtered[e.Id] = guestNoID{
			Name: e.Name,
			Node: e.Node,
			Pool: e.Pool,
			Tags: e.Tags,
			Type: e.Type,
		}
	}
	return true
}

// check if a guest is already in the map
// the tags of the guests map[uint]guestNoID have to be sorted
func filtered(guest filter.Guest, guests map[uint]guestNoID) bool {
	if value, ok := guests[guest.Id]; ok {
		if value.Type == guest.Type && value.Name == guest.Name && value.Node == guest.Node && value.Pool == guest.Pool {
			slices.Sort[string](guest.Tags)
			if slices.Equal(value.Tags, guest.Tags) {
				return true
			}
		}
	}
	return false
}

// obtain a new guest list from the proxmox api
func newGuestList(client *pxAPI.Client) ([]filter.Guest, error) {
	rawGuestList, err := pxAPI.ListGuests(client)
	if err != nil {
		return nil, err
	}
	return convertToGuestList(rawGuestList), nil
}

type guestMinimal struct {
	Id   uint
	Type pxAPI.GuestType
}

type guestNoID struct {
	Name string
	Node string
	Pool string
	Tags []string
	Type pxAPI.GuestType
}

type guests struct {
	Filtered  map[uint]guestNoID
	Completed []guestMinimal
}
