package autosnap

import (
	"go-pve-autosnap/internal/filter"
	"testing"

	pxAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

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
