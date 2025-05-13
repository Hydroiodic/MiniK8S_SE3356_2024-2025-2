package kubelet

func main() {
	println("Starting kubelet...")
	// Additional initialization code would go here.
	// 进入一个循环，不断检查自身、接收配置……

	// 1. 启动Go Routine监听Pod创建请求

	// 2. 启动Go Routine监听Pod删除请求

	// 3. 启动Go Routine获取最新Pod配置，自我修复 & 发送心跳
}
