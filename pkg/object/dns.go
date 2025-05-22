package object

// DNS 数据结构定义
type DNS struct {
	Kind     string   `yaml:"kind"`     // 固定为 "DNS"
	Metadata Metadata `yaml:"metadata"` // 名称、命名空间、标签
	Spec     DNSSpec  `yaml:"spec"`     // DNS 配置
}

// DNSSpec 表示 DNS 的配置
type DNSSpec struct {
	Host  string    `yaml:"host"`  // 域名主路径
	Paths []DNSPath `yaml:"paths"` // 子路径列表
}

// DNSPath 表示每个子路径的配置
type DNSPath struct {
	Path        string `yaml:"path"`        // 子路径地址
	ServiceName string `yaml:"serviceName"` // 对应的 Service 名称
	ServiceIP   string `yaml:"serviceIP"`   // 对应的 Service IP
}

// DNSResolveInfo 用于消息队列传递消息
type DNSResolveInfo struct {
	Host string // 域名
	IP   string // IP 地址
}

type DnsMsg struct {
	Dns        DNS
	HostConfig []string
}
