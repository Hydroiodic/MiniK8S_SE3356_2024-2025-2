package object

// `DNS` defines the structure of a DNS object in the system.
type DNS struct {
	Kind     string   `yaml:"kind"` // always "DNS"
	Metadata Metadata `yaml:"metadata"`
	Spec     DNSSpec  `yaml:"spec"`
}

// `DNSSpec` represents the specification for the DNS object.
type DNSSpec struct {
	Host  string    `yaml:"host"`  // The domain name for the DNS record
	Paths []DNSPath `yaml:"paths"` // The list of paths associated with the domain
}

// `DNSPath` represents a specific path within a DNS record.
type DNSPath struct {
	Path        string `yaml:"path"`        // The path for the DNS record, e.g., "/api"
	ServiceName string `yaml:"serviceName"` // The name of the service associated with this path
	ServiceIP   string `yaml:"serviceIP"`   // The IP address of the service associated with this path
	ServicePort int    `yaml:"servicePort"` // The port of the service associated with this path
}

// `DNSResolveInfo` is used to store the DNS resolution information.
type DNSResolveInfo struct {
	Host string
	IP   string
}

// `ProxyRule` defines a forwarding rule for the dynamic proxy.
type ProxyRule struct {
	Domain     string // e.g. "example.com"
	PathPrefix string
	Target     string // e.g. "http://10.0.0.2:80" or "https://10.0.0.1:443"
}
