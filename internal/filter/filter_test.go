package filter

import (
	"strconv"
	"testing"

	pxAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func Test_Apply(t *testing.T) {
	testInput := func() []Guest {
		return []Guest{
			{Id: 100, Name: "abcde", Type: pxAPI.GuestLXC, Node: "pve1", Tags: []string{"no-snapshot"}},
			{Id: 200, Name: "thing", Type: pxAPI.GuestLXC, Node: "test1", Pool: "dev"},
			{Id: 300, Name: "vm45", Type: pxAPI.GuestQemu, Node: "pve1", Pool: "prod", Tags: []string{"staging"}},
			{Id: 400, Name: "ct400", Type: pxAPI.GuestLXC, Node: "pve1", Pool: "prod", Tags: []string{"no-snapshot", "automation", "staging"}},
			{Id: 600, Name: "copy-of-ct400", Type: pxAPI.GuestLXC, Node: "pve1", Pool: "dev", Tags: []string{"no-snapshot"}},
			{Id: 700, Name: "700", Type: pxAPI.GuestQemu, Node: "pve2", Tags: []string{"no-snapshot"}},
			{Id: 800, Name: "test1", Type: pxAPI.GuestQemu, Node: "pve2"},
			{Id: 900, Name: "vm73", Type: pxAPI.GuestQemu, Node: "pve2", Tags: []string{"automation"}},
			{Id: 10000, Name: "ct45", Type: pxAPI.GuestLXC, Node: "test1", Pool: "staging", Tags: []string{"automation"}},
			{Id: 1729, Name: "test1", Type: pxAPI.GuestQemu, Node: "pve1"},
		}
	}
	tests := []struct {
		name   string
		filter *FilterSteps
		input  []Guest
		output []bool
	}{
		// Add
		{name: "Add GuestId",
			filter: &FilterSteps{Steps: []Step{{Add: true, GuestId: []uint{300, 800, 900}, Type: Id}}},
			input:  testInput(),
			output: []bool{false, false, true, false, false, false, true, true, false, false},
		},
		{name: "Add GuestName",
			filter: &FilterSteps{Steps: []Step{{Add: true, GuestName: []string{"ct400", "test1"}, Type: Name}}},
			input:  testInput(),
			output: []bool{false, false, false, true, false, false, true, false, false, true},
		},
		{name: "Add Node",
			filter: &FilterSteps{Steps: []Step{{Add: true, Node: []string{"pve1", "pve2"}, Type: Node}}},
			input:  testInput(),
			output: []bool{true, false, true, true, true, true, true, true, false, true},
		},
		{name: "Add Pool",
			filter: &FilterSteps{Steps: []Step{{Add: true, Pool: []string{"prod"}, Type: Pool}}},
			input:  testInput(),
			output: []bool{false, false, true, true, false, false, false, false, false, false},
		},
		{name: "Add Tag",
			filter: &FilterSteps{Steps: []Step{{Add: true, Tag: []string{"automation", "staging"}, Type: Tag}}},
			input:  testInput(),
			output: []bool{false, false, true, true, false, false, false, true, true, false},
		},
		// Remove
		{name: "Remove GuestId",
			filter: &FilterSteps{All: true, Steps: []Step{{GuestId: []uint{400, 200, 700}, Type: Id}}},
			input:  testInput(),
			output: []bool{true, false, true, false, true, false, true, true, true, true},
		},
		{name: "Remove GuestName",
			filter: &FilterSteps{All: true, Steps: []Step{{GuestName: []string{"test1", "thing"}, Type: Name}}},
			input:  testInput(),
			output: []bool{true, false, true, true, true, true, false, true, true, false},
		},
		{name: "Remove Node",
			filter: &FilterSteps{All: true, Steps: []Step{{Node: []string{"test1"}, Type: Node}}},
			input:  testInput(),
			output: []bool{true, false, true, true, true, true, true, true, false, true},
		},
		{name: "Remove Pool",
			filter: &FilterSteps{All: true, Steps: []Step{{Pool: []string{"dev", "staging"}, Type: Pool}}},
			input:  testInput(),
			output: []bool{true, false, true, true, false, true, true, true, false, true},
		},
		{name: "Remote Tag",
			filter: &FilterSteps{All: true, Steps: []Step{{Tag: []string{"no-snapshot"}, Type: Tag}}},
			input:  testInput(),
			output: []bool{false, true, true, false, false, false, true, true, true, true},
		},
		// All
		{name: "@all only",
			filter: &FilterSteps{All: true},
			input:  testInput(),
			output: []bool{true, true, true, true, true, true, true, true, true, true},
		},
		{name: "full test with @all",
			filter: &FilterSteps{All: true, Steps: []Step{
				{GuestId: []uint{100, 300, 600}, Type: Id},
				{Add: true, GuestName: []string{"abcde", "vm45", "700"}, Type: Name},
				{Node: []string{"test1"}, Type: Node},
				{Add: true, Pool: []string{"prod"}, Type: Pool},
				{Tag: []string{"no-snapshot", "staging"}, Type: Tag},
				{Add: true, Tag: []string{"automation"}, Type: Tag},
				{Pool: []string{"dev", "staging"}, Type: Pool},
				{Add: true, Node: []string{"pve1", "pve2"}, Type: Node},
				{GuestName: []string{"test1"}, Type: Name},
				{Add: true, GuestId: []uint{1729}, Type: Id},
			}},
			input:  testInput(),
			output: []bool{true, false, true, true, true, true, false, true, false, true},
		},
		{name: "full test without @all",
			filter: &FilterSteps{Steps: []Step{
				{Add: true, GuestId: []uint{100, 300, 600}, Type: Id},
				{GuestName: []string{"abcde", "vm45", "700"}, Type: Name},
				{Add: true, Node: []string{"test1"}, Type: Node},
				{Pool: []string{"prod"}, Type: Pool},
				{Add: true, Tag: []string{"no-snapshot", "staging"}, Type: Tag},
				{Tag: []string{"automation"}, Type: Tag},
				{Add: true, Pool: []string{"dev", "staging"}, Type: Pool},
				{Node: []string{"pve1", "pve2"}, Type: Node},
				{Add: true, GuestName: []string{"test1"}, Type: Name},
				{GuestId: []uint{1729}, Type: Id},
			}},
			input:  testInput(),
			output: []bool{false, true, false, false, false, false, true, false, true, false},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			for i := range test.input {
				tmpOutput := test.filter.Apply(test.input[i])
				require.Equal(t, test.output[i], tmpOutput, test.name+" "+strconv.Itoa(int(test.input[i].Id)))
			}
		})
	}
}
