package autosnap

import (
	"context"
	"go-pve-autosnap/internal/filter"
	"go-pve-autosnap/internal/pool"
	"strings"
	"sync/atomic"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
)

type Function func(ctx context.Context, c pve.ClientNew, vmr *pve.VmRef) error

const maxAttempts = maxIndex + 1
const maxIndex = 2

// run the provided function on all guests that match the filter.
func Execute(ctx context.Context, c pve.ClientNew, filterObj *filter.Filter, pool pool.Pool, f Function) error {
	tracked := make(map[pve.GuestID]*[]error)
	var subroutineSucceeded bool
	var firstRun bool = true
	for !subroutineSucceeded {
		guests, err := newGuestList(ctx, c.Guest)
		if err != nil {
			return err
		}
		if firstRun { // Initialize tracker
			for i := range guests {
				tracked[guests[i].GetID()] = new([]error)
			}
			firstRun = false
		}
		subroutineSucceeded = executeSubroutine(ctx, guests, tracked, maxAttempts, c, filterObj, pool, f)
	}
	errs := make([]GuestError, 0)
	for id, e := range tracked {
		if len(*e) == maxAttempts {
			errs = append(errs, GuestError{
				ID:  id,
				Err: (*e)[maxIndex],
			})
		}
	}
	if len(errs) > 0 {
		return &ExecuteError{Errs: errs}
	}
	return nil
}

// subroutine of Execute to extract the testable code.
func executeSubroutine(ctx context.Context,
	guests []pve.RawGuestResource, tracked map[pve.GuestID]*[]error, maxAttempts int,
	c pve.ClientNew, filterObj *filter.Filter, pool pool.Pool, f Function,
) bool {
	var success atomic.Bool
	success.Store(true)
	for i := range guests {
		id := guests[i].GetID()
		attempts, ok := tracked[id]
		if !ok { // Skip guests that where not discovered in the initial run
			continue
		}
		if *attempts != nil { // check if a snapshot has been made of the guest during this run.
			if len(*attempts) == 0 || len(*attempts) == maxAttempts {
				continue
			}
		}
		guest := guests[i].Get()
		if guest.Template {
			continue
		}
		if filterObj.Apply(&guest) {
			vmr := pve.NewVmRef(id)
			vmr.SetNode(string(guest.Node))
			vmr.SetVmType(guest.Type)
			pool.Submit(guest.Node, func() {
				if err := f(ctx, c, vmr); err != nil {
					*attempts = append(*attempts, err)
					success.Store(false)
				} else {
					*attempts = make([]error, 0)
				}
			})
		}
	}
	pool.Wait()
	return success.Load()
}

// obtain a new guest list from the proxmox api
func newGuestList(ctx context.Context, c pve.GuestInterface) ([]pve.RawGuestResource, error) {
	raw, err := c.List(ctx)
	if err != nil {
		return nil, err
	}
	return raw.AsArray(), nil
}

type ExecuteError struct {
	Errs []GuestError
}

func (e *ExecuteError) Error() string {
	var b strings.Builder
	for i := range e.Errs {
		b.WriteByte(',')
		b.WriteString(e.Errs[i].ID.String())
	}
	return "operation failed on guest(s): " + b.String()[1:]
}

type GuestError struct {
	ID  pve.GuestID
	Err error
}

func (e *GuestError) Error() string {
	return e.ID.String() + ": " + e.Err.Error()
}
