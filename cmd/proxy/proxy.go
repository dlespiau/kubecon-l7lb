package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dlespiau/kubecon-l7lb/pkg/affinity"
	k8s "github.com/dlespiau/kubecon-l7lb/pkg/kubernetes"
)

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

type proxyStats struct {
	sync.Mutex
	// Maps key:endpoint -> number of requests
	requests map[string]int
}

type proxy struct {
	resolver  affinity.Resolver
	header    string
	reverse   httputil.ReverseProxy
	noForward bool
	stats     proxyStats
}

func (p *proxy) printStats() {
	type stat struct {
		key   string
		value int
	}
	var results []stat

	p.stats.Lock()
	for key, requests := range p.stats.requests {
		results = append(results, stat{key, requests})
	}
	p.stats.Unlock()

	sort.Slice(results, func(i, j int) bool {
		return results[i].key < results[j].key
	})
	for i := range results {
		fmt.Printf("%s: %d\n", results[i].key, results[i].value)
	}
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get(p.header)
	if key == "" {
		http.Error(w, fmt.Sprintf("unable to find %s header", p.header), http.StatusBadRequest)
		return
	}

	endpoint := p.resolver.Resolve(key)

	fmt.Printf("%s -> %s\n", key, endpoint.Address)

	// Update stats. XXX make it faster!
	p.stats.Lock()
	p.stats.requests[fmt.Sprintf("%s-%s", key, endpoint.Address)]++
	p.stats.Unlock()

	// Forward request to endpoint.
	if p.noForward {
		p.resolver.Release(endpoint)
		return
	}

	r.Host = endpoint.Address
	r.URL.Host = endpoint.Address
	r.URL.Scheme = "http"

	p.reverse.ServeHTTP(w, r)

	p.resolver.Release(endpoint)

}

func main() {
	kubeconfig := flag.String("k8s.kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	namespace := flag.String("k8s.namespace", "default", "namespace of the service to load balance")
	service := flag.String("k8s.service", "", "name of the service to load balance")
	listen := flag.String("proxy.listen", ":8081", "address the proxy should listen on")
	header := flag.String("proxy.header", "X-Affinity", "name of the HTTP header taken as input")
	keepAlive := flag.Bool("proxy.keep-alive", true, "whether the proxy should keep its connections to endpoints alive")
	noForward := flag.Bool("proxy.no-forward", false, "don't forward request downstream (debug)")
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

	resolver := affinity.NewBoundedLoadRing()
	watcher := affinity.EndpointWatcher{
		Client: client,
		Service: affinity.Service{
			Namespace: *namespace,
			Name:      *service,
			Port:      "8080",
		},
		Receiver: resolver,
	}

	watcher.Start(make(<-chan interface{}))

	proxy := &proxy{
		noForward: *noForward,
		resolver:  resolver,
		header:    *header,
		reverse: httputil.ReverseProxy{
			Director:  proxyDirector,
			Transport: proxyTransport(*keepAlive),
		},
		stats: proxyStats{
			requests: make(map[string]int),
		},
	}

	// Print stats on SIGUSR1
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1)
	go func() {
		for range signals {
			proxy.printStats()
		}
	}()

	log.Fatal(http.ListenAndServe(*listen, proxy))
}
