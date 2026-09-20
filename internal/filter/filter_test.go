package filter

import (
	"errors"
	"testing"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

type test struct {
	err    error
	filter Filter
	input  string
	name   string
	data   []pve.GuestResource
	output []bool
}

func testDataInvalid() []test {
	return []test{
		{name: `errEmptyFilter`,
			err: errors.New(errEmptyFilter)},
		{name: `errors.New(errMissingOperatorAt0)`,
			input: `a`,
			err:   &ParseError{message: errMissingOperatorAt0}},
		{name: `errors.New(errNoOperatorEnd)`,
			input: `@all`,
			err: &ParseError{
				message:    errNoOperatorEnd,
				charNumber: 4}},
		{name: `errors.New(errNoOperator)`,
			input: `@:`,
			err: &ParseError{
				message:    errNoOperator,
				charNumber: 1}},
		{name: `errors.New(errUndefinedOperator)`,
			input: `@all:@undefined:`,
			err: &ParseError{
				message:    errUndefinedOperator,
				charNumber: 6}},
		{name: `errors.New(errRangeID) -`,
			input: `@id:-`,
			err:   errors.New(errRangeID)},
		{name: `errors.New(errRangeID) -1`,
			input: `@id:-1`,
			err:   errors.New(errRangeID)},
		{name: `errors.New(errRangeID) 1-`,
			input: `@id:1-`,
			err:   errors.New(errRangeID)},
		{name: `errors.New(errRangeID) a-`,
			input: `@id:a-`,
			err:   errors.New(errRangeID)},
		{name: `errors.New(errRangeID) -a`,
			input: `@id:-a`,
			err:   errors.New(errRangeID)},
		{name: `errors.New(errRangeID) 1-a`,
			input: `@id:1-a`,
			err:   errors.New(errRangeID)},
		{name: `errors.New(errRangeID) a-1`,
			input: `@id:a-1`,
			err:   errors.New(errRangeID)},
	}
}

func testDataValid() []test {
	testInput := func() []pve.GuestResource {
		return []pve.GuestResource{
			{ID: 100, Name: "abcde", Type: pve.GuestLxc, Node: "pve1", Tags: []pve.Tag{"no-snapshot"}},
			{ID: 200, Name: "thing", Type: pve.GuestLxc, Node: "test1", Pool: "dev"},
			{ID: 300, Name: "vm45", Type: pve.GuestQemu, Node: "pve1", Pool: "prod", Tags: []pve.Tag{"staging"}},
			{ID: 400, Name: "ct400", Type: pve.GuestLxc, Node: "pve1", Pool: "prod", Tags: []pve.Tag{"no-snapshot", "automation", "staging"}},
			{ID: 600, Name: "copy-of-ct400", Type: pve.GuestLxc, Node: "pve1", Pool: "dev", Tags: []pve.Tag{"no-snapshot"}},
			{ID: 700, Name: "700", Type: pve.GuestQemu, Node: "pve2", Tags: []pve.Tag{"no-snapshot"}},
			{ID: 800, Name: "test1", Type: pve.GuestQemu, Node: "pve2"},
			{ID: 900, Name: "vm73", Type: pve.GuestQemu, Node: "pve2", Tags: []pve.Tag{"automation"}},
			{ID: 10000, Name: "ct45", Type: pve.GuestLxc, Node: "test1", Pool: "staging", Tags: []pve.Tag{"automation"}},
			{ID: 1729, Name: "test1", Type: pve.GuestQemu, Node: "pve1"},
		}
	}
	return []test{
		{name: `@all:`,
			input: `@all:`,
			filter: Filter{
				step:      []FilterStep{filterAll()},
				alignment: []bool{true}},
			data:   testInput(),
			output: []bool{true, true, true, true, true, true, true, true, true, true}},
		{name: "@id:300,`800`,900 1700-1800",
			input: "@id:300,`800`,900 1700-1800",
			filter: Filter{
				step:      []FilterStep{filterID([]pve.GuestID{300, 800, 900}, []idRange{{min: 1700, max: 1800}})},
				alignment: []bool{true}},
			data:   testInput(),
			output: []bool{false, false, true, false, false, false, true, true, false, true}},
		{name: `@name:'ct400',test1`,
			input: `@name:'ct400',test1`,
			filter: Filter{
				step:      []FilterStep{filterName([]pve.GuestName{"ct400", "test1"})},
				alignment: []bool{true}},
			data:   testInput(),
			output: []bool{false, false, false, true, false, false, true, false, false, true}},
		{name: `@node:pve1,pve2`,
			input: `@node:pve1,pve2`,
			filter: Filter{
				step:      []FilterStep{filterNode([]pve.NodeName{"pve1", "pve2"})},
				alignment: []bool{true}},
			data:   testInput(),
			output: []bool{true, false, true, true, true, true, true, true, false, true}},
		{name: `@pool:prod`,
			input: `@pool:prod`,
			filter: Filter{
				step:      []FilterStep{filterPool([]pve.PoolName{"prod"})},
				alignment: []bool{true}},
			data:   testInput(),
			output: []bool{false, false, true, true, false, false, false, false, false, false}},
		{name: `@tag:automation,staging`,
			input: `@tag:automation,staging`,
			filter: Filter{
				step:      []FilterStep{filterTag([]pve.Tag{"automation", "staging"})},
				alignment: []bool{true}},
			data:   testInput(),
			output: []bool{false, false, true, true, false, false, false, true, true, false}},
		{name: `@+all:@-id:400,200,700`,
			input: `@+all:@-id:400,200,700`,
			filter: Filter{
				step: []FilterStep{
					filterAll(),
					filterID([]pve.GuestID{400, 200, 700}, []idRange{})},
				alignment: []bool{true, false}},
			data:   testInput(),
			output: []bool{true, false, true, false, true, false, true, true, true, true}},
		{name: `@+all:@-name:test1,"thing"`,
			input: `@+all:@-name:test1,"thing"`,
			filter: Filter{
				step: []FilterStep{
					filterAll(),
					filterName([]pve.GuestName{"test1", "thing"})},
				alignment: []bool{true, false}},
			data:   testInput(),
			output: []bool{true, false, true, true, true, true, false, true, true, false}},
		{name: `@all:@-node:"test1"`,
			input: `@all:@-node:"test1"`,
			filter: Filter{
				step: []FilterStep{
					filterAll(),
					filterNode([]pve.NodeName{"test1"})},
				alignment: []bool{true, false}},
			data:   testInput(),
			output: []bool{true, false, true, true, true, true, true, true, false, true}},
		{name: `@all:@-pool:dev,staging`,
			input: `@all:@-pool:dev,staging`,
			filter: Filter{
				step: []FilterStep{
					filterAll(),
					filterPool([]pve.PoolName{"dev", "staging"})},
				alignment: []bool{true, false}},
			data:   testInput(),
			output: []bool{true, false, true, true, false, true, true, true, false, true}},
		{name: `@all:@-tag:no-snapshot`,
			input: `@all:@-tag:no-snapshot`,
			filter: Filter{
				step: []FilterStep{
					filterAll(),
					filterTag([]pve.Tag{"no-snapshot"})},
				alignment: []bool{true, false}},
			data:   testInput(),
			output: []bool{false, true, true, false, false, false, true, true, true, true}},
		{name: `@all:@-id:100 300,600@+name:abcde,vm45;700@-node:test1@pool:prod@-tag:no-snapshot,staging@tag:automation@-pool:dev,staging@node:pve1,pve2@-name:test1@id:1700-1800`,
			input: `@all:@-id:100 300,600@+name:abcde,vm45;700@-node:test1@pool:prod@-tag:no-snapshot,staging@tag:automation@-pool:dev,staging@node:pve1,pve2@-name:test1@id:1700-1800`,
			filter: Filter{
				step: []FilterStep{
					filterAll(),
					filterID([]pve.GuestID{100, 300, 600}, []idRange{}),
					filterName([]pve.GuestName{"abcde", "vm45", "700"}),
					filterNode([]pve.NodeName{"test1"}),
					filterPool([]pve.PoolName{"prod"}),
					filterTag([]pve.Tag{"no-snapshot", "staging"}),
					filterTag([]pve.Tag{"automation"}),
					filterPool([]pve.PoolName{"dev", "staging"}),
					filterNode([]pve.NodeName{"pve1", "pve2"}),
					filterName([]pve.GuestName{"test1"}),
					filterID([]pve.GuestID{}, []idRange{{min: 1700, max: 1800}}),
				},
				alignment: []bool{true, false, true, false, true, false, true, false, true, false, true}},
			data:   testInput(),
			output: []bool{true, false, true, true, true, true, false, true, false, true}},
		{name: `@id:100 300,600@-name:abcde,vm45;700@node:test1@-pool:prod@tag:no-snapshot,staging@-tag:automation@pool:dev,staging@-node:pve1,pve2@name:test1@-id:1700-1800`,
			input: `@id:100 300,600@-name:abcde,vm45;700@node:test1@-pool:prod@tag:no-snapshot,staging@-tag:automation@pool:dev,staging@-node:pve1,pve2@name:test1@-id:1700-1800`,
			filter: Filter{
				step: []FilterStep{
					filterID([]pve.GuestID{100, 300, 600}, []idRange{}),
					filterName([]pve.GuestName{"abcde", "vm45", "700"}),
					filterNode([]pve.NodeName{"test1"}),
					filterPool([]pve.PoolName{"prod"}),
					filterTag([]pve.Tag{"no-snapshot", "staging"}),
					filterTag([]pve.Tag{"automation"}),
					filterPool([]pve.PoolName{"dev", "staging"}),
					filterNode([]pve.NodeName{"pve1", "pve2"}),
					filterName([]pve.GuestName{"test1"}),
					filterID([]pve.GuestID{}, []idRange{{min: 1700, max: 1800}}),
				},
				alignment: []bool{true, false, true, false, true, false, true, false, true, false}},
			data:   testInput(),
			output: []bool{false, true, false, false, false, false, true, false, true, false}},
	}
}

func Benchmark_Apply(b *testing.B) {
	tests := testDataValid()
	bench := make([]test, 0, len(tests))
	for i := range tests {
		if tests[i].err == nil {
			bench = append(bench, tests[i])
		}
	}
	for b.Loop() {
		for i := range bench {
			for ii := range bench[i].data {
				bench[i].filter.Apply(&bench[i].data[ii])
			}
		}
	}
}

func Test_Apply(t *testing.T) {
	t.Parallel()
	tests := append(testDataInvalid(), testDataValid()...)
	for _, test := range tests {
		if test.err != nil {
			continue
		}
		t.Run(test.name, func(*testing.T) {
			for i := range test.data {
				output := test.filter.Apply(&test.data[i])
				require.Equal(t, test.output[i], output, i)
			}
		})
	}
}

func Benchmark_Parse(b *testing.B) {
	tests := testDataValid()
	for b.Loop() {
		for i := range tests {
			f := newFilter()
			f.Parse(tests[i].input)
		}
	}
}

func Test_Parse(t *testing.T) {
	t.Parallel()
	tests := append(testDataInvalid(), testDataValid()...)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			f := newFilter()
			filter, err := f.Parse(test.input)
			require.Equal(t, test.err, err)
			for i := range test.data {
				output := filter.Apply(&test.data[i])
				require.Equal(t, test.output[i], output, i)
			}
		})
	}
}

func newFilter() Filters {
	f := New()
	f.Register(ConstructorAll())
	f.Register(ConstructorID())
	f.Register(ConstructorName())
	f.Register(ConstructorNode())
	f.Register(ConstructorPool())
	f.Register(ConstructorTag())
	return f
}
