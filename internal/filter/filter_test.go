package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Apply(t *testing.T) {
	testInput := func() []Guest {
		return []Guest{
			{Id: 100, Name: "abcde", Node: "pve1", Tag: []string{"no-snapshot"}},
			{Id: 200, Name: "thing", Node: "test1", Pool: "dev"},
			{Id: 300, Name: "vm45", Node: "pve1", Pool: "prod", Tag: []string{"staging"}},
			{Id: 400, Name: "ct400", Node: "pve1", Pool: "prod", Tag: []string{"no-snapshot", "automation", "staging"}},
			{Id: 600, Name: "copy-of-ct400", Node: "pve1", Pool: "dev", Tag: []string{"no-snapshot"}},
			{Id: 700, Name: "700", Node: "pve2", Tag: []string{"no-snapshot"}},
			{Id: 800, Name: "test1", Node: "pve2"},
			{Id: 900, Name: "vm73", Node: "pve2", Tag: []string{"automation"}},
			{Id: 10000, Name: "ct45", Node: "test1", Pool: "staging", Tag: []string{"automation"}},
			{Id: 1729, Name: "test1", Node: "pve1"},
		}
	}
	tests := []struct {
		name   string
		filter *FilterSteps
		input  []Guest
		output []uint
	}{
		// Add
		{name: "Add GuestId",
			filter: &FilterSteps{Steps: []Step{{Add: true, GuestId: []uint{300, 800, 900}, Type: Id}}},
			input:  testInput(),
			output: []uint{300, 800, 900},
		},
		{name: "Add GuestName",
			filter: &FilterSteps{Steps: []Step{{Add: true, GuestName: []string{"ct400", "test1"}, Type: Name}}},
			input:  testInput(),
			output: []uint{400, 800, 1729},
		},
		{name: "Add Node",
			filter: &FilterSteps{Steps: []Step{{Add: true, Node: []string{"pve1", "pve2"}, Type: Node}}},
			input:  testInput(),
			output: []uint{100, 300, 400, 600, 700, 800, 900, 1729},
		},
		{name: "Add Pool",
			filter: &FilterSteps{Steps: []Step{{Add: true, Pool: []string{"prod"}, Type: Pool}}},
			input:  testInput(),
			output: []uint{300, 400},
		},
		{name: "Add Tag",
			filter: &FilterSteps{Steps: []Step{{Add: true, Tag: []string{"automation", "staging"}, Type: Tag}}},
			input:  testInput(),
			output: []uint{300, 400, 900, 10000},
		},
		// Remove
		{name: "Remove GuestId",
			filter: &FilterSteps{All: true, Steps: []Step{{GuestId: []uint{400, 200, 700}, Type: Id}}},
			input:  testInput(),
			output: []uint{100, 300, 600, 800, 900, 10000, 1729},
		},
		{name: "Remove GuestName",
			filter: &FilterSteps{All: true, Steps: []Step{{GuestName: []string{"test1", "thing"}, Type: Name}}},
			input:  testInput(),
			output: []uint{100, 300, 400, 600, 700, 900, 10000},
		},
		{name: "Remove Node",
			filter: &FilterSteps{All: true, Steps: []Step{{Node: []string{"test1"}, Type: Node}}},
			input:  testInput(),
			output: []uint{100, 300, 400, 600, 700, 800, 900, 1729},
		},
		{name: "Remove Pool",
			filter: &FilterSteps{All: true, Steps: []Step{{Pool: []string{"dev", "staging"}, Type: Pool}}},
			input:  testInput(),
			output: []uint{100, 300, 400, 700, 800, 900, 1729},
		},
		{name: "Remote Tag",
			filter: &FilterSteps{All: true, Steps: []Step{{Tag: []string{"no-snapshot"}, Type: Tag}}},
			input:  testInput(),
			output: []uint{200, 300, 800, 900, 10000, 1729},
		},
		// All
		{name: "@all only",
			filter: &FilterSteps{All: true},
			input:  testInput(),
			output: []uint{100, 200, 300, 400, 600, 700, 800, 900, 10000, 1729},
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
			output: []uint{100, 300, 400, 600, 700, 900, 1729},
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
			output: []uint{200, 800, 10000},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			tmpOutput := Apply(test.input, test.filter)
			require.Equal(t, test.output, tmpOutput, test.name)
		})
	}
}

func Test_countMarkedGuests(t *testing.T) {
	tests := []struct {
		name   string
		input  []Guest
		output uint
	}{
		{name: "Valid",
			input: []Guest{
				{Id: 100, Mark: true},
				{Id: 101, Mark: false},
				{Id: 111, Mark: true},
				{Id: 200, Mark: true},
				{Id: 202, Mark: false},
				{Id: 222, Mark: false},
				{Id: 900, Mark: true},
				{Id: 909, Mark: true},
				{Id: 999, Mark: true},
				{Id: 1000, Mark: false},
				{Id: 10000, Mark: false},
				{Id: 100000, Mark: false},
			},
			output: 6,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, countMarkedGuests(test.input), test.name)
		})
	}
}
