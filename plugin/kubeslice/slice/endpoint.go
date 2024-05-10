package slice

type ServicePort struct {
	Name     string
	Port     int32
	Protocol string
}

// Endpoint corresponds to a dns entry for an endpint in the slice
type Endpoint struct {
	Host        string
	IP          string
	Ports       []ServicePort
	TargetStrip int
}
