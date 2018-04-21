package affinity

// Service describe a kubernetes service. Port can be a named port or the port
// number.
type Service struct {
	Namespace, Name, Port string
}

// Endpoint is a service endpoint.
type Endpoint struct {
	Address string
}

func (e *Endpoint) String() string {
	return e.Address
}

// Resolver is an interface resolving a key into an Endpoint.
//
// Call Release on the Endpoint returned by Resolve once the request has been
// handled.
type Resolver interface {
	SetEndpoints([]Endpoint)
	Resolve(key string) *Endpoint
	Release(endpoint *Endpoint)
}
