package affinity

import (
	"fmt"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/buraksezer/consistent"

	client "k8s.io/client-go/kubernetes"
)

// BoundedLoadRing is a service Resolver using consistent hashing with bounded load.
type BoundedLoadRing struct {
	client  *client.Clientset
	service Service

	consistent *consistent.Consistent
}

var _ Resolver = &BoundedLoadRing{}

// NewBoundedLoadRing creates a new BoundedLoadRing resolver.
func NewBoundedLoadRing(client *client.Clientset, config consistent.Config, service Service) *BoundedLoadRing {
	return &BoundedLoadRing{
		client:     client,
		service:    service,
		consistent: consistent.New(nil, config),
	}
}

func (r *BoundedLoadRing) makeEndpoints(subsets []corev1.EndpointSubset) []Endpoint {
	var endpoints []Endpoint

	servicePort, err := strconv.Atoi(r.service.Port)

	for _, subset := range subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				if (err == nil && servicePort == int(port.Port)) ||
					(err != nil && r.service.Port == port.Name) {
					endpoints = append(endpoints, Endpoint{
						Address: net.JoinHostPort(address.IP, strconv.Itoa(int(port.Port))),
					})
				}
			}
		}
	}

	return endpoints
}

// Start implements Resolver.
func (r *BoundedLoadRing) Start(close <-chan interface{}) {
	go func() {
		events, err := r.client.CoreV1().Endpoints(r.service.Namespace).Watch(metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", r.service.Name),
		})
		if err != nil {
			log.Fatal(err)
			return
		}

		for {
			select {
			case <-close:
				log.Info("stop watching endpoints")
				events.Stop()
				return
			case event := <-events.ResultChan():
				endpoints, ok := event.Object.(*corev1.Endpoints)
				if !ok {
					log.Warnf("didn't receive endpoints in event")
					break
				}

				switch event.Type {
				case watch.Added:
					members := r.makeEndpoints(endpoints.Subsets)
					for i := range members {
						log.Debugf("ring: add %s", members[i].Address)
						r.consistent.Add(&members[i])
					}
				case watch.Modified:
					// Build the list of old endpoints.
					var old []string
					for _, member := range r.consistent.GetMembers() {
						old = append(old, member.String())
					}

					// And the list of new ones.
					var new []string
					for _, endpoint := range r.makeEndpoints(endpoints.Subsets) {
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
				case watch.Deleted:
					for _, member := range r.consistent.GetMembers() {
						log.Debugf("ring: remove %s", member.String())
						r.consistent.Remove(member.String())
					}
				case watch.Error:
					// XXX: log a better error. Maybe we lost the connection with the API
					// server? Should probably retry with backoff.
					log.Warnf("received an error during watch")

				}
			}
		}
	}()
}

// Resolve implements Resolver.
func (r *BoundedLoadRing) Resolve(key string) *Endpoint {
	return r.consistent.LocateKey([]byte(key)).(*Endpoint)
}
