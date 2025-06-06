package function

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/gpu"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/mholt/archiver"
)

// func controller 负责根据etcd中的内容创建function的pod
type FucntionController struct {
	cache map[string]object.Function
	ci    *apiserver.APIClient
}

func NewFucntionController() *FucntionController {
	return &FucntionController{
		cache: make(map[string]object.Function),
		ci:    apiserver.NewAPIClient(""),
	}
}

func (fc *FucntionController) Start() {
	fc = NewFucntionController()

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for range ticker.C {
			fc.CheckFunctions()
		}
	}()
}
func (fc *FucntionController) CheckFunctions() {
	funcs, err := fc.ci.GetAllFunction()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("开始检查Functions : %d\n", len(funcs))

	cur := make(map[string]bool)

	for _, f := range funcs {
		cur[f.Metadata.Namespace+"/"+f.Metadata.Name] = true

		if _, ok := fc.cache[f.Metadata.Namespace+"/"+f.Metadata.Name]; !ok {
			fmt.Printf(
				"Create function %s %s\n",
				f.Metadata.Namespace,
				f.Metadata.Name,
			)

			fc.cache[f.Metadata.Namespace+"/"+f.Metadata.Name] = f
			fc.CreateFunctionAndReplicaset(f) //这里replicaset方便进行动态伸缩
			fc.CreateService(f)
		}
	}
	//处理etcd中已经删除的记录
	for k, f := range fc.cache {
		fmt.Println(k)

		if _, ok := cur[f.Metadata.Namespace+"/"+f.Metadata.Name]; !ok {
			fmt.Println("DeleteFunction", f.Metadata.Namespace, f.Metadata.Name)
			fc.DeleteFunctionAndReplicaset(f)
			fc.DeleteService(f)
			delete(fc.cache, k)
		}
	}
}
func (fc *FucntionController) DeleteFunctionAndReplicaset(f object.Function) {
	var cli *client.Client

	var err error
	cli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer func() {
		if cerr := cli.Close(); cerr != nil {
			// Log the error if closing the response body fails.
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	err = fc.ci.DeleteReplicaset(f.Metadata.Name)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

// 只是会创建镜像，但是不会运行pod（replicaset为0）
func (fc *FucntionController) CreateFunctionAndReplicaset(f object.Function) {
	//创建docker容器的挂载目录
	fmt.Println("开始创建挂载目录")

	FunctionFilePath := gpu.WorkDir + "/assets/allfunctions/" +
		f.Metadata.Namespace + "/" + f.Metadata.Name
	err := os.RemoveAll(FunctionFilePath)

	if err != nil {
		fmt.Println(err)
	}

	err = os.MkdirAll(FunctionFilePath, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	//先检查zip文件是否存在，如果存在，则删除
	err = os.RemoveAll(FunctionFilePath + "/function.zip")
	if err != nil {
		fmt.Println(err)
	}

	err = os.RemoveAll(FunctionFilePath + "/function")
	if err != nil {
		fmt.Println(err)
	}
	//创建zip文件
	fmt.Println("开始创建zip文件")

	err = os.WriteFile(
		FunctionFilePath+"/function.zip",
		f.Spec.UserUploadFile,
		0777,
	)

	if err != nil {
		fmt.Println(err)
		return
	}
	//解压zip文件
	//将解压后的文件放入新文件夹
	fmt.Println("开始解压zip文件")

	z := archiver.NewZip()
	z.OverwriteExisting = true
	err = z.Unarchive(FunctionFilePath+"/function.zip", FunctionFilePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	//删除压缩包
	err = os.Remove(FunctionFilePath + "/function.zip")
	if err != nil {
		fmt.Println(err)
	}

	//创建dockerfile
	fmt.Println("开始创建dockfile")

	err = os.Remove(FunctionFilePath + "/Dockerfile")

	if err != nil {
		fmt.Println(err)
	}

	dockerfile, err := os.Create(FunctionFilePath + "/Dockerfile")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if cerr := dockerfile.Close(); cerr != nil {
			// Log the error if closing the response body fails.
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	_, err = dockerfile.WriteString(
		"FROM " + gpu.ImageRegistryURL + ":" + strconv.Itoa(
			gpu.ImageRegistryPort,
		) + "/baseserver:latest\n",
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	funcpath := f.Metadata.Name + "/"
	_, err = dockerfile.WriteString("COPY " + funcpath + " /app\n") //???????

	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("aaaaaaaaaaaaaaaaaa")

		return
	}

	_, err = dockerfile.WriteString("EXPOSE 10000\n")
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	fmt.Println("开始打包tar")
	//构建docker上下文，需要将依赖文件打包成tar格式
	z2 := archiver.NewTar()
	z2.OverwriteExisting = true
	err = z2.Archive(
		[]string{FunctionFilePath},
		FunctionFilePath+"/function.tar",
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ctx, err := os.Open(FunctionFilePath + "/function.tar")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//构建docker镜像
	fmt.Println("开始构建镜像")

	var cli *client.Client
	cli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer func() {
		if cerr := cli.Close(); cerr != nil {
			// Log the error if closing the response body fails.
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	resp, err := cli.ImageBuild(
		context.Background(),
		ctx,
		types.ImageBuildOptions{
			Dockerfile: f.Metadata.Name + "/Dockerfile",
			Tags: []string{
				fmt.Sprintf(
					"%s:%d/baseserver/%s/%s:latest",
					gpu.ImageRegistryURL,
					gpu.ImageRegistryPort,
					f.Metadata.Namespace,
					f.Metadata.Name,
				),
			},
			Context: ctx,
			Remove:  true,
		},
	)

	if err != nil {
		s := fmt.Sprintf(
			"%s:%d/baseserver/%s/%s:latest",
			gpu.ImageRegistryURL,
			gpu.ImageRegistryPort,
			f.Metadata.Namespace,
			f.Metadata.Name,
		)
		fmt.Println(s)
		fmt.Println(err.Error())

		return
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Log the error if closing the response body fails.
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//推送docker镜像到docker registry
	fmt.Println("开始推送镜像")

	authEncoded := base64.StdEncoding.EncodeToString(
		[]byte(gpu.Registry_user + ":" + gpu.Registry_password),
	)
	fmt.Println(authEncoded)
	resp2, err := cli.ImagePush(
		context.Background(),
		fmt.Sprintf(
			"%s:%d/baseserver/%s/%s:latest",
			gpu.ImageRegistryURL,
			gpu.ImageRegistryPort,
			f.Metadata.Namespace,
			f.Metadata.Name,
		),
		image.PushOptions{
			RegistryAuth: authEncoded,
			All:          false,
		},
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer func() {
		if cerr := resp2.Close(); cerr != nil {
			// Log the error if closing the response body fails.
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	_, err = io.Copy(os.Stdout, resp2)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fc.CreateReplicas(f)
}

func (fc *FucntionController) CreateReplicas(f object.Function) {
	var rs object.ReplicaSet
	rs.Kind = "Replicaset"
	rs.Metadata.Namespace = f.Metadata.Namespace
	rs.Metadata.Name = f.Metadata.Name
	rs.Spec.Replicas = 0
	rs.Spec.Selector = make(map[string]string)

	rs.Spec.Template.Metadata.Name = f.Metadata.Name

	rs.Spec.Template.Metadata.Labels = make(map[string]string)
	rs.Spec.Selector["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	rs.Spec.Template.Metadata.Labels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	rs.Spec.Template.Spec.Containers = make([]object.Container, 1)
	rs.Spec.Template.Spec.Containers[0].Ports = make([]int, 1)
	rs.Spec.Template.Spec.Containers[0].Ports[0] = 10000
	rs.Spec.Template.Spec.Containers[0].Name = f.Metadata.Name
	rs.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf(
		"%s:%d/baseserver/%s/%s:latest",
		gpu.ImageRegistryURL,
		gpu.ImageRegistryPort,
		f.Metadata.Namespace,
		f.Metadata.Name,
	)
	err := fc.ci.CreateReplicaset(&rs)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (fc *FucntionController) CreateService(f object.Function) {
	var s object.Service
	s.Kind = "Service"
	s.Type = object.SERVICE_TYPE_CLUSTERIP_STR
	s.Metadata.Labels = make(map[string]string)
	s.Metadata.Labels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	s.Metadata.Name = f.Metadata.Name + "-service"
	s.Metadata.Namespace = f.Metadata.Namespace
	s.Spec.Selector = make(map[string]string)
	s.Spec.Selector["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	s.Spec.Ports = make([]object.ServicePort, 1)
	s.Spec.Ports[0].Name = "http"
	s.Spec.Ports[0].TargetPort = 10000
	s.Spec.Ports[0].NodePort = 30001
	s.Spec.Ports[0].Port = 81

	err := fc.ci.CreateService(&s)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
func (fc *FucntionController) DeleteService(f object.Function) {
	services, err := fc.ci.GetServices()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var match_s object.Service

	existFlag := 0

	for _, s := range services {
		if s.Metadata.Name == f.Metadata.Name+"-service" {
			match_s = s
			existFlag = 1

			break
		}
	}

	if existFlag == 0 {
		fmt.Println("function service not found")
		return
	}

	err = fc.ci.DeleteService(&match_s)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
