package filter

import (
	"errors"
	"strconv"
	"strings"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
)

const (
	errRangeID = "invalid id range"
)

func ConstructorAll() Constructor {
	return Constructor{
		Operator: "all",
		Func: func(args string) (FilterStep, error) {
			if len(args) != 0 {
				return nil, errors.New("operator all must not have any arguments")
			}
			return filterAll(), nil
		}}
}

func ConstructorID() Constructor {
	return Constructor{
		Operator: "id",
		Func: func(args string) (FilterStep, error) {
			if len(args) == 0 {
				return nil, errors.New("operator id must have at least one argument")
			}
			parsedArgs := split(args)
			exactArgs := make([]pve.GuestID, 0)
			rangeArgs := make([]idRange, 0)
			for i := range parsedArgs {
				if v := strings.IndexByte(parsedArgs[i], '-'); v != -1 {
					if v == 0 || v == len(parsedArgs[i])-1 {
						return nil, errors.New(errRangeID)
					}
					tmpMin, err := strconv.ParseUint(parsedArgs[i][:v], 10, 32)
					if err != nil {
						return nil, errors.New(errRangeID)
					}
					tmpMax, err := strconv.ParseUint(parsedArgs[i][v+1:], 10, 32)
					if err != nil {
						return nil, errors.New(errRangeID)
					}
					if tmpMax <= tmpMin {
						return nil, errors.New(errRangeID)
					}
					rangeArgs = append(rangeArgs, idRange{
						min: pve.GuestID(tmpMin),
						max: pve.GuestID(tmpMax)})
					continue
				}
				tmpID, err := strconv.ParseUint(parsedArgs[i], 10, 32)
				if err != nil {
					return nil, errors.New("invalid id")
				}
				id := pve.GuestID(tmpID)
				if err := id.Validate(); err != nil {
					return nil, errors.New("invalid id")
				}
				exactArgs = append(exactArgs, id)
			}
			return filterID(exactArgs, rangeArgs), nil
		}}
}

func ConstructorName() Constructor {
	return Constructor{
		Operator: "name",
		Func: func(args string) (FilterStep, error) {
			if len(args) == 0 {
				return nil, errors.New("operator name must have at least one argument")
			}
			parsedArgs := split(args)
			checkArgs := make([]pve.GuestName, len(parsedArgs))
			for i := range parsedArgs {
				checkArgs[i] = pve.GuestName(parsedArgs[i])
				if err := checkArgs[i].Validate(); err != nil {
					return nil, errors.New("invalid name")
				}
			}
			return filterName(checkArgs), nil
		}}
}

func ConstructorNode() Constructor {
	return Constructor{
		Operator: "node",
		Func: func(args string) (FilterStep, error) {
			if len(args) == 0 {
				return nil, errors.New("operator node must have at least one argument")
			}
			parsedArgs := split(args)
			checkArgs := make([]pve.NodeName, len(parsedArgs))
			for i := range parsedArgs {
				checkArgs[i] = pve.NodeName(parsedArgs[i])
				if err := checkArgs[i].Validate(); err != nil {
					return nil, errors.New("invalid node")
				}
			}
			return filterNode(checkArgs), nil
		}}
}

func ConstructorPool() Constructor {
	return Constructor{
		Operator: "pool",
		Func: func(args string) (FilterStep, error) {
			if len(args) == 0 {
				return nil, errors.New("operator pool must have at least one argument")
			}
			parsedArgs := split(args)
			checkArgs := make([]pve.PoolName, len(parsedArgs))
			for i := range parsedArgs {
				checkArgs[i] = pve.PoolName(parsedArgs[i])
				if err := checkArgs[i].Validate(); err != nil {
					return nil, errors.New("invalid pool")
				}
			}
			return filterPool(checkArgs), nil
		}}
}

func ConstructorTag() Constructor {
	return Constructor{
		Operator: "tag",
		Func: func(args string) (FilterStep, error) {
			if len(args) == 0 {
				return nil, errors.New("operator tag must have at least one argument")
			}
			parsedArgs := split(args)
			checkArgs := make([]pve.Tag, len(parsedArgs))
			for i := range parsedArgs {
				checkArgs[i] = pve.Tag(parsedArgs[i])
				if err := checkArgs[i].Validate(); err != nil {
					return nil, errors.New("invalid tag")
				}
			}
			return filterTag(checkArgs), nil
		}}
}
