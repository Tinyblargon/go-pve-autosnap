package filter

type FilterSteps struct {
	All   bool
	Steps []Step
}

type Guest struct {
	Id   uint
	Name string
	Node string
	Pool string
	Tag  []string
	Mark bool
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

// use the filter to mark all guests for snapshots
func Apply(guests []Guest, filter *FilterSteps) []uint {
	if filter.All {
		for i := range guests {
			guests[i].Mark = true
		}
	}
	for _, step := range filter.Steps {
		switch step.Type {
		case Id:
			for i := range guests {
				if step.Add != guests[i].Mark {
					if inArray(guests[i].Id, step.GuestId) {
						guests[i].Mark = step.Add
					}
				}
			}
		case Name:
			for i := range guests {
				if step.Add != guests[i].Mark {
					if inArray(guests[i].Name, step.GuestName) {
						guests[i].Mark = step.Add
					}
				}
			}
		case Node:
			for i := range guests {
				if step.Add != guests[i].Mark {
					if inArray(guests[i].Node, step.Node) {
						guests[i].Mark = step.Add
					}
				}
			}
		case Pool:
			for i := range guests {
				if step.Add != guests[i].Mark {
					if inArray(guests[i].Pool, step.Pool) {
						guests[i].Mark = step.Add
					}
				}
			}
		case Tag:
			for i := range guests {
				if step.Add != guests[i].Mark {
					for _, tag := range guests[i].Tag {
						if inArray(tag, step.Tag) {
							guests[i].Mark = step.Add
							break
						}
					}
				}
			}
		}
	}
	listOfGuests := make([]uint, countMarkedGuests(guests))
	var counter int
	for _, e := range guests {
		if e.Mark {
			listOfGuests[counter] = e.Id
			counter++
		}
	}
	return listOfGuests
}

// Counts the number of guests that have `Mark` set to true
func countMarkedGuests(guests []Guest) uint {
	var markedGuests uint
	for _, e := range guests {
		if e.Mark {
			markedGuests++
		}
	}
	return markedGuests
}

func inArray[k comparable](item k, array []k) bool {
	for _, e := range array {
		if item == e {
			return true
		}
	}
	return false
}
