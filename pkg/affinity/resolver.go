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
type Resolver interface {
	Start(close <-chan interface{})
	Resolve(key string) *Endpoint
}
