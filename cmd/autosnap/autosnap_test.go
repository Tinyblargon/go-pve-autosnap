package autosnap

import (
	"errors"
	"go-pve-autosnap/internal/filter"
	"testing"

	pxAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func Test_completed(t *testing.T) {
	tests := []struct {
		name   string
		input  filter.Guest
		inList []guestMinimal
		output bool
	}{
		{name: "Empty list",
			input:  filter.Guest{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"},
			inList: []guestMinimal{},
			output: false,
		},
		{name: "In list",
			input: filter.Guest{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"},
			inList: []guestMinimal{
				{Id: 101, Type: pxAPI.GuestQemu},
				{Id: 102, Type: pxAPI.GuestLXC},
				{Id: 100, Type: pxAPI.GuestLXC},
			},
			output: true,
		},
		{name: "Not in list",
			input: filter.Guest{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"},
			inList: []guestMinimal{
				{Id: 101, Type: pxAPI.GuestQemu},
				{Id: 102, Type: pxAPI.GuestLXC},
				{Id: 103, Type: pxAPI.GuestLXC},
			},
			output: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, completed(test.input, test.inList))
		})
	}
}

func Test_convertToGuestList(t *testing.T) {
	tests := []struct {
		name   string
		input  []pxAPI.GuestResource
		output []filter.Guest
	}{
		{name: "Empty",
			input:  []pxAPI.GuestResource{},
			output: []filter.Guest{},
		},
		{name: "Single",
			input: []pxAPI.GuestResource{
				{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1", Tags: []string{"no-snapshot"}},
			},
			output: []filter.Guest{
				{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1", Tags: []string{"no-snapshot"}},
			},
		},
		{name: "Multiple",
			input: []pxAPI.GuestResource{
				{Id: 101, Name: "test", Type: pxAPI.GuestQemu, Node: "pve2", Tags: []string{"no-snapshot"}},
				{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"},
				{Id: 102, Name: "third-test", Type: pxAPI.GuestLXC, Node: "test", Tags: []string{"no-snapshot", "automated"}},
			},
			output: []filter.Guest{
				{Id: 101, Name: "test", Type: pxAPI.GuestQemu, Node: "pve2", Tags: []string{"no-snapshot"}},
				{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"},
				{Id: 102, Name: "third-test", Type: pxAPI.GuestLXC, Node: "test", Tags: []string{"no-snapshot", "automated"}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, convertToGuestList(test.input))
		})
	}
}

func Test_executeSubroutine(t *testing.T) {
	tests := []struct {
		name          string
		guests        []filter.Guest
		filterObj     *filter.FilterSteps
		trackedGuests *guests
		function      func(client *pxAPI.Client, vmr *pxAPI.VmRef) error
		output        bool
		outputTracked *guests
	}{
		{name: "add to trackedGuests.Completed",
			guests:    []filter.Guest{{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"}},
			filterObj: &filter.FilterSteps{All: true},
			trackedGuests: &guests{
				Filtered:  make(map[uint]guestNoID),
				Completed: []guestMinimal{{Id: 200, Type: pxAPI.GuestQemu}},
			},
			function: func(client *pxAPI.Client, vmr *pxAPI.VmRef) error {
				return nil
			},
			output: true,
			outputTracked: &guests{
				Filtered: make(map[uint]guestNoID),
				Completed: []guestMinimal{
					{Id: 200, Type: pxAPI.GuestQemu},
					{Id: 100, Type: pxAPI.GuestLXC},
				},
			},
		},
		{name: "add to trackedGuests.Filtered",
			guests:    []filter.Guest{{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1", Tags: []string{"no-snapshot", "automated"}}},
			filterObj: &filter.FilterSteps{All: false},
			trackedGuests: &guests{
				Filtered:  map[uint]guestNoID{200: {Name: "test", Type: pxAPI.GuestQemu, Node: "pve"}},
				Completed: make([]guestMinimal, 0),
			},
			function: func(client *pxAPI.Client, vmr *pxAPI.VmRef) error {
				return nil
			},
			output: true,
			outputTracked: &guests{
				Filtered: map[uint]guestNoID{
					// tags have to be sorted
					100: {Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1", Tags: []string{"automated", "no-snapshot"}},
					200: {Name: "test", Type: pxAPI.GuestQemu, Node: "pve"},
				},
				Completed: make([]guestMinimal, 0),
			},
		},
		{name: "return false on error",
			guests:    []filter.Guest{{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"}},
			filterObj: &filter.FilterSteps{All: true},
			trackedGuests: &guests{
				Filtered:  map[uint]guestNoID{200: {Name: "test", Type: pxAPI.GuestQemu, Node: "pve"}},
				Completed: []guestMinimal{{Id: 500, Type: pxAPI.GuestQemu}},
			},
			function: func(client *pxAPI.Client, vmr *pxAPI.VmRef) error {
				return errors.New("test error")
			},
			output: false,
			outputTracked: &guests{
				Filtered:  map[uint]guestNoID{200: {Name: "test", Type: pxAPI.GuestQemu, Node: "pve"}},
				Completed: []guestMinimal{{Id: 500, Type: pxAPI.GuestQemu}},
			},
		},
		{name: "skip trackedGuests.Completed",
			guests:    []filter.Guest{{Id: 100, Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve"}},
			filterObj: &filter.FilterSteps{All: false},
			trackedGuests: &guests{
				Filtered:  make(map[uint]guestNoID),
				Completed: []guestMinimal{{Id: 100, Type: pxAPI.GuestQemu}},
			},
			function: func(client *pxAPI.Client, vmr *pxAPI.VmRef) error {
				require.FailNow(t, "function should not be called")
				return nil
			},
			output: true,
		},
		{name: "skip trackedGuests.Filtered",
			guests:    []filter.Guest{{Id: 100, Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve"}},
			filterObj: &filter.FilterSteps{All: false},
			trackedGuests: &guests{
				Filtered:  map[uint]guestNoID{100: {Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve"}},
				Completed: make([]guestMinimal, 0),
			},
			function: func(client *pxAPI.Client, vmr *pxAPI.VmRef) error {
				require.FailNow(t, "function should not be called")
				return nil
			},
			output: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, executeSubroutine(test.guests, test.trackedGuests, nil, test.filterObj, test.function))
			if test.outputTracked != nil {
				require.Equal(t, test.outputTracked, test.trackedGuests)
			}
		})
	}
}

func Test_filtered(t *testing.T) {
	tests := []struct {
		name   string
		input  filter.Guest
		inMap  map[uint]guestNoID
		output bool
	}{
		{name: "Empty map",
			input:  filter.Guest{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"},
			inMap:  make(map[uint]guestNoID),
			output: false,
		},
		{name: "In map",
			input: filter.Guest{Id: 100, Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve1", Tags: []string{"no-snapshot", "automated"}},
			inMap: map[uint]guestNoID{
				101: {Name: "test", Type: pxAPI.GuestLXC, Node: "pve"},
				100: {Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve1", Tags: []string{"automated", "no-snapshot"}},
				200: {Name: "third-test", Type: pxAPI.GuestLXC, Node: "test", Tags: []string{"no-snapshot"}},
			},
			output: true,
		},
		{name: "Map tags not in order",
			input: filter.Guest{Id: 100, Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve1", Tags: []string{"no-snapshot", "automated"}},
			inMap: map[uint]guestNoID{
				101: {Name: "test", Type: pxAPI.GuestLXC, Node: "pve"},
				100: {Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve1", Tags: []string{"no-snapshot", "automated"}},
				200: {Name: "third-test", Type: pxAPI.GuestLXC, Node: "test", Tags: []string{"no-snapshot"}},
			},
			output: false,
		},
		{name: "Not in map",
			input: filter.Guest{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1"},
			inMap: map[uint]guestNoID{
				101: {Name: "test", Type: pxAPI.GuestLXC, Node: "pve"},
				100: {Name: "abcde", Type: pxAPI.GuestQemu, Node: "pve1", Tags: []string{"automated", "no-snapshot"}},
				200: {Name: "third-test", Type: pxAPI.GuestLXC, Node: "test", Tags: []string{"no-snapshot"}},
			},
			output: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, filtered(test.input, test.inMap))
		})
	}
}
