package object

type Function struct {
	Kind     string       `yaml:"kind"     json:"kind"`
	Metadata Metadata     `yaml:"metadata" json:"metadata"`
	Spec     FunctionSpec `yaml:"spec"     json:"spec"`
}

type FunctionSpec struct {
	UserUploadFile []byte `yaml:"userUploadFile"     json:"userUploadFile"`     //函数zip文件的byte形式
	UserUploadPath string `yaml:"userUploadFilePath" json:"userUploadFilePath"` //这个函数的源文件所在的路径
}
