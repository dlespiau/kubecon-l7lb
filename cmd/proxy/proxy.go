package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
	log "github.com/sirupsen/logrus"

	"github.com/dlespiau/kubecon-l7lb/pkg/affinity"
	k8s "github.com/dlespiau/kubecon-l7lb/pkg/kubernetes"
)

// xxHash for the ring.
type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func proxyDirector(req *http.Request) {}

func proxyTransport(keepAlive bool) http.RoundTripper {
	return &http.Transport{
		DisableKeepAlives: !keepAlive,

		// Rest are from http.DefaultTransport
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type proxy struct {
	resolver affinity.Resolver
	header   string
	reverse  httputil.ReverseProxy
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get(p.header)
	if key == "" {
		http.Error(w, fmt.Sprintf("unable to find %s header", p.header), http.StatusBadRequest)
		return
	}

	endpoint := p.resolver.Resolve(key)

	fmt.Printf("%s -> %s\n", key, endpoint.Address)

	r.Host = endpoint.Address
	r.URL.Host = endpoint.Address
	r.URL.Scheme = "http"

	p.reverse.ServeHTTP(w, r)
}

func main() {
	kubeconfig := flag.String("k8s.kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	namespace := flag.String("k8s.namespace", "default", "namespace of the service to load balance")
	service := flag.String("k8s.service", "", "name of the service to load balance")
	listen := flag.String("proxy.listen", ":8081", "address the proxy should listen on")
	header := flag.String("proxy.header", "X-Affinity", "name of the HTTP header taken as input")
	keepAlive := flag.Bool("proxy.keep-alive", true, "whether the proxy should keep its connections to endpoints alive")
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	if *kubeconfig == "" {
		*kubeconfig = os.Getenv("KUBECONFIG")
	}

	client, err := k8s.NewClientWithConfig(&k8s.ClientConfig{
		Kubeconfig: *kubeconfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	resolver := affinity.NewBoundedLoadRing(
		client,
		consistent.Config{
			Hasher:            hasher{},
			PartitionCount:    71,
			ReplicationFactor: 20,
			Load:              1.25,
		},
		affinity.Service{
			Namespace: *namespace,
			Name:      *service,
			Port:      "8080",
		})

	resolver.Start(make(<-chan interface{}))

	log.Fatal(http.ListenAndServe(*listen, &proxy{
		resolver: resolver,
		header:   *header,
		reverse: httputil.ReverseProxy{
			Director:  proxyDirector,
			Transport: proxyTransport(*keepAlive),
		},
	}))
}
