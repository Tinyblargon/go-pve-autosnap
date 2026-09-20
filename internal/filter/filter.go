package filter

import (
	"errors"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
)

func New() Filters { return Filters{a: make(map[Operator]FilterConstructor)} }

type Constructor struct {
	Operator Operator
	Func     FilterConstructor
}

type FilterConstructor func(args string) (FilterStep, error)

type Operator string

type FilterStep func(*pve.GuestResource) bool

type Filters struct {
	a map[Operator]FilterConstructor
}

func (filters Filters) Register(c Constructor) error {
	if _, registered := filters.a[c.Operator]; registered {
		return errors.New("operator already registered")
	}
	filters.a[c.Operator] = c.Func
	return nil
}

type Filter struct {
	step      []FilterStep
	alignment []bool
}

func (f Filter) Apply(gr *pve.GuestResource) bool {
	var status bool = false
	for i := range f.alignment {
		if status == f.alignment[i] {
			continue
		}
		if f.step[i](gr) {
			status = f.alignment[i]
		}
	}
	return status
}
