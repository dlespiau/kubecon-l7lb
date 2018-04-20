package affinity

import (
	log "github.com/sirupsen/logrus"

	"github.com/buraksezer/consistent"
)

// BoundedLoadRing is a service Resolver using consistent hashing with bounded load.
type BoundedLoadRing struct {
	consistent *consistent.Consistent
}

var _ Resolver = &BoundedLoadRing{}

// NewBoundedLoadRing creates a new BoundedLoadRing resolver.
func NewBoundedLoadRing(config consistent.Config) *BoundedLoadRing {
	return &BoundedLoadRing{
		consistent: consistent.New(nil, config),
	}
}

// setEndpoints implements endpointReceiver.
func (r *BoundedLoadRing) setEndpoints(endpoints []Endpoint) {
	// Build the list of old endpoints.
	var old []string
	for _, member := range r.consistent.GetMembers() {
		old = append(old, member.String())
	}

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
			r.consistent.Add(&Endpoint{
				Address: chunk.str,
			})
		case del:
			log.Debugf("ring: remove %s", chunk.str)
			r.consistent.Remove(chunk.str)
		}
	}
}

// Resolve implements Resolver.
func (r *BoundedLoadRing) Resolve(key string) *Endpoint {
	return r.consistent.LocateKey([]byte(key)).(*Endpoint)
}
