package filter

import (
	pxAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"golang.org/x/exp/slices"
)

type FilterSteps struct {
	All   bool
	Steps []Step
}

type Guest struct {
	Id   uint
	Name string
	Node string
	Pool string
	Tags []string
	Type pxAPI.GuestType
}

type StepType uint8

const (
	Empty StepType = iota
	Id
	Name
	Node
	Pool
	Tag
)

type Step struct {
	Add       bool
	GuestId   []uint
	GuestName []string
	Node      []string
	Pool      []string
	Tag       []string
	Type      StepType
}

// use the filter to mark a guest for snapshot
func (filter *FilterSteps) Apply(guest Guest) bool {
	mark := filter.All
	for _, step := range filter.Steps {
		if step.Add != mark {
			switch step.Type {
			case Id:
				if slices.Contains(step.GuestId, guest.Id) {
					mark = step.Add
				}
			case Name:
				if slices.Contains(step.GuestName, guest.Name) {
					mark = step.Add
				}
			case Node:
				if slices.Contains(step.Node, guest.Node) {
					mark = step.Add
				}
			case Pool:
				if step.Add != mark {
					if slices.Contains(step.Pool, guest.Pool) {
						mark = step.Add
					}
				}
			case Tag:
				for _, tag := range guest.Tags {
					if slices.Contains(step.Tag, tag) {
						mark = step.Add
						break
					}
				}
			}
		}
	}
	return mark
}
