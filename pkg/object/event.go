package object

type Event struct {
	Metadata Metadata    `yaml:"metadata" json:"metadata"`
	Kind     string      `yaml:"kind"     json:"kind"`
	Schedule string      `yaml:"schedule" json:"schedule"` // 事件触发周期字符串，例如"*/1 * * * *"表示每分钟触发一次
	JsonData string      `yaml:"jsonData" json:"jsonData"` // 事件触发时，写入事件的Json数据
	Target   EventTarget `yaml:"target"   json:"target"`   // 事件触发时，写入事件的目标
}
type EventTarget struct {
	Kind      string `yaml:"kind"      json:"kind"`
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}
