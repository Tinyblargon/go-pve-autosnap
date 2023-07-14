package filter

import (
	pxAPI "github.com/Telmate/proxmox-api-go/proxmox"
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
				if inArray(guest.Id, step.GuestId) {
					mark = step.Add
				}
			case Name:
				if inArray(guest.Name, step.GuestName) {
					mark = step.Add
				}
			case Node:
				if inArray(guest.Node, step.Node) {
					mark = step.Add
				}
			case Pool:
				if step.Add != mark {
					if inArray(guest.Pool, step.Pool) {
						mark = step.Add
					}
				}
			case Tag:
				for _, tag := range guest.Tags {
					if inArray(tag, step.Tag) {
						mark = step.Add
						break
					}
				}
			}
		}
	}
	return mark
}

func inArray[k comparable](item k, array []k) bool {
	for _, e := range array {
		if item == e {
			return true
		}
	}
	return false
}
