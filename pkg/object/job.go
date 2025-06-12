package object

type Job struct {
	ApiVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind"       json:"kind"`
	Metadata   Metadata `yaml:"metadata"   json:"metadata"`
	Spec       JobSpec  `yaml:"spec"       json:"spec"`
	Status     string   `yaml:"status"     json:"status"`
}

type JobSpec struct {
	Partition      string   `yaml:"partition"      json:"partition"`
	OutputFile     string   `yaml:"outputFile"     json:"outputFile"`
	ErrorFile      string   `yaml:"errorFile"      json:"errorFile"`
	RunCommands    []string `yaml:"runCommands"    json:"runCommands"`
	GPUNums        string   `yaml:"GPUNums"        json:"GPUNums"`
	CPUPerTask     string   `yaml:"CPUPerTask"     json:"CPUPerTask"`
	UploadPath     string   `yaml:"uploadPath"     json:"uploadPath"`
	NTasks         string   `yaml:"nTasks"         json:"nTasks"`
	NTasksPerNode  string   `yaml:"nTasksPerNode"  json:"nTasksPerNode"`
	UserUploadFile []byte   `yaml:"userUploadFile" json:"userUploadFile"`
}

type JobRequestBody struct {
	JobID        string `json:"job_id"`
	JobName      string `json:"jobname"`
	JobNamespace string `json:"jobnamespace"`
	Output       string `json:"output,omitempty"`
	Error        string `json:"error,omitempty"`
	Status       string `json:"status"`
}
