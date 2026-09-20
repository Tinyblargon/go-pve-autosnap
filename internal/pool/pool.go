package pool

import (
	pve "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/alitto/pond/v2"
)

func New(maxConcurrency uint, nodes []pve.NodeName, nodeConcurrency uint) Pool {
	pool := pond.NewPool(int(maxConcurrency))
	if len(nodes) != 0 {
		if nodeConcurrency > maxConcurrency {
			nodeConcurrency = maxConcurrency
		}
		pools := make([]pond.Pool, len(nodes))
		for i := range pools {
			pools[i] = pool.NewSubpool(int(nodeConcurrency))
		}
		poolMap := make(map[pve.NodeName]*pond.TaskGroup, len(nodes))
		for i := range nodes {
			poolMap[nodes[i]] = new(pools[i].NewGroup())
		}
		return &threadsNodes{groups: poolMap}
	}
	return &threads{group: pool.NewGroup()}
}

type Pool interface {
	Submit(pve.NodeName, ...func())
	Wait()
}

type threads struct {
	group pond.TaskGroup
}

var _ Pool = (*threads)(nil)

func (t *threads) Submit(name pve.NodeName, tasks ...func()) { t.group.Submit(tasks...) }

func (t *threads) Wait() { t.group.Wait() }

type threadsNodes struct {
	// Pointer to pond.TaskGroup so interactions with it can't modify the layout of the map.
	// This map is read concurrently.
	groups map[pve.NodeName]*pond.TaskGroup
}

var _ Pool = (*threadsNodes)(nil)

func (t *threadsNodes) Submit(name pve.NodeName, tasks ...func()) { (*t.groups[name]).Submit(tasks...) }

func (t *threadsNodes) Wait() {
	for _, v := range t.groups {
		(*v).Wait()
	}
}
