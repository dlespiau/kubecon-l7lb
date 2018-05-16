package affinity

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/dlespiau/kubecon-consistent"
)

// HashRing is a service Resolver using consistent hashing.
type HashRing struct {
	consistent *consistent.Consistent

	// Protect the endpoints map .
	sync.RWMutex
	// Maps an address to the corresponding *Endpoint.
	endpoints map[string]*Endpoint
}

var _ Resolver = &HashRing{}

// XXX: we want to expose the consistent hashing knobs:
//   - replication count
// XXX: I liked the Member interface of https://github.com/buraksezer/consistent so
// we can get back the actual Endpoint object from the ring without having to
// maintain a local hash table.

// NewHashRing creates a new HashRing resolver.
func NewHashRing() *HashRing {
	return &HashRing{
		consistent: consistent.New(),
		endpoints:  make(map[string]*Endpoint),
	}
}

// SetEndpoints implements Resolver.
func (r *HashRing) SetEndpoints(endpoints []Endpoint) {
	// Build the list of old endpoints.
	old := r.consistent.Hosts()

	// And the list of new ones.
	var new []string
	for _, endpoint := range endpoints {
		new = append(new, endpoint.Address)
	}

	// Adjust ring members.
	for _, chunk := range difference(old, new) {
		switch chunk.operation {
		case add:
			log.Debugf("ring: add %s", chunk.str)
			r.Lock()
			r.endpoints[chunk.str] = &Endpoint{
				Address: chunk.str,
			}
			r.Unlock()
			r.consistent.Add(chunk.str)
		case del:
			log.Debugf("ring: remove %s", chunk.str)
			r.consistent.Remove(chunk.str)
			r.Lock()
			delete(r.endpoints, chunk.str)
			r.Unlock()
		}
	}
}

// Resolve implements Resolver.
func (r *HashRing) Resolve(key string) *Endpoint {
	address, err := r.consistent.Get(key)
	if err != nil {
		return nil
	}

	r.RLock()
	endpoint := r.endpoints[address]
	r.RUnlock()
	return endpoint
}

// Release implements Resolver.
func (r *HashRing) Release(endpoint *Endpoint) {
}
