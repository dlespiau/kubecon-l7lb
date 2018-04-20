package affinity

import (
	"fmt"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	client "k8s.io/client-go/kubernetes"
)

type endpointReceiver interface {
	setEndpoints([]Endpoint)
}

// EndpointWatcher load balances requests.
type EndpointWatcher struct {
	Client   *client.Clientset
	Service  Service
	Receiver endpointReceiver
}

func (w *EndpointWatcher) makeEndpoints(subsets []corev1.EndpointSubset) []Endpoint {
	var endpoints []Endpoint

	servicePort, err := strconv.Atoi(w.Service.Port)

	for _, subset := range subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				if (err == nil && servicePort == int(port.Port)) ||
					(err != nil && w.Service.Port == port.Name) {
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
func (w *EndpointWatcher) Start(close <-chan interface{}) {
	go func() {
		events, err := w.Client.CoreV1().Endpoints(w.Service.Namespace).Watch(metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", w.Service.Name),
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
					log.Warnf("received a %s object", event.Object.GetObjectKind().GroupVersionKind().Kind)
					time.Sleep(1)
					break
				}

				switch event.Type {
				case watch.Added:
					fallthrough
				case watch.Modified:
					members := w.makeEndpoints(endpoints.Subsets)
					w.Receiver.setEndpoints(members)
				case watch.Deleted:
					w.Receiver.setEndpoints(nil)
				case watch.Error:
					// XXX: log a better error. Maybe we lost the connection with the API
					// server? Should probably retry with backoff.
					log.Warnf("received an error during watch")

				}
			}
		}
	}()
}
