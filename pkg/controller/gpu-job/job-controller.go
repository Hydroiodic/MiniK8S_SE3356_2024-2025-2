package gpujob

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/mholt/archiver"
)

// jobs/
// └── <namespace>/
//     └── <job-name>/
//         ├── job.zip
//         ├── job.tar
//         ├── Dockerfile
//         ├── <namespace>_<job-name>.slurm
//         ├── <zip解压后内容> (如 .cu文件 等)

const WorkDir = "/home/ubuntu/wang/MiniK8S_SE3356_2024-2025-2"
const registry_user = "useruser"
const registry_password = "HTEQRHOAOLDN"

// cache: 用于缓存已处理的 Job 对象，key 格式为 "namespace/name"
type JobController struct {
	cache map[string]object.Job
	ci    *apiserver.APIClient
}

func NewJobController() *JobController {
	return &JobController{
		cache: make(map[string]object.Job),
		ci:    apiserver.NewAPIClient(""),
	}
}

func (fc *JobController) Start() {
	//每10s进行一次routine操作
	fmt.Println("JobController Run")
	fc = NewJobController()
	ticker := time.NewTicker(10 * time.Second)
	// defer ticker.Stop()
	// 开启一个goroutine执行轮询操作
	go func() {
		for {
			select {
			case <-ticker.C:
				fc.CheckJob()
			}
		}
	}()
}

func (fc *JobController) CheckJob() {
	jobs, err := fc.ci.GetAllGpuJob()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("开始检查gpujob : %d", len(jobs))

	cur := make(map[string]bool)

	for _, job := range jobs {
		cur[job.Metadata.Namespace+"/"+job.Metadata.Name] = true
		if _, ok := fc.cache[job.Metadata.Namespace+"/"+job.Metadata.Name]; !ok {
			fmt.Printf(
				"CreateJob %s %s\n",
				job.Metadata.Namespace,
				job.Metadata.Name,
			)
			fc.cache[job.Metadata.Namespace+"/"+job.Metadata.Name] = job
			fc.CreateJob(job)
		}
	}
}

func (fc *JobController) CreateJob(job object.Job) {
	fmt.Println("检查到未创建gpujob,开始创建")
	// 任务的文件夹路径
	JobFilePath := WorkDir + "/jobs/" + job.Metadata.Namespace + "/" + job.Metadata.Name
	//解析.cu文件
	fmt.Println("开始解析.cu文件")

	err := os.RemoveAll(JobFilePath)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = os.MkdirAll(JobFilePath, 0777)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = os.WriteFile(JobFilePath+"/job.zip", job.Spec.UserUploadFile, 0777)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("开始解压")
	z := archiver.NewZip()
	z.OverwriteExisting = true
	err = z.Unarchive(JobFilePath+"/job.zip", JobFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 写slurm文件
	fmt.Println("开始写slurm文件")
	scriptanme := job.Metadata.Namespace + "_" + job.Metadata.Name + ".slurm"
	slurm, err := os.Create(JobFilePath + "/" + scriptanme)
	// // 逐行写入内容
	// _, err = slurm.WriteString("#!/bin/bash\n\n")
	// _, err = slurm.WriteString("#SBATCH --job-name=aaaaaa" + "\n")
	// _, err = slurm.WriteString("#SBATCH --partition=cpu\n")
	// _, err = slurm.WriteString(
	// 	"#SBATCH --output=" + job.Spec.OutputFile + "\n",
	// ) // 你的需求
	// _, err = slurm.WriteString("#SBATCH --error=" + job.Spec.ErrorFile + "\n")
	// _, err = slurm.WriteString("#SBATCH -n 8\n")
	// _, err = slurm.WriteString("#SBATCH --ntasks-per-node=8\n\n")
	// _, err = slurm.WriteString("ulimit -l unlimited\n")
	// _, err = slurm.WriteString("ulimit -s unlimited\n\n")
	// _, err = slurm.WriteString("module load gcc\n\n")
	// _, err = slurm.WriteString("export OMP_NUM_THREADS=8\n")
	// _, err = slurm.WriteString("./omphello\n")

	// if err != nil {
	// 	panic(err)
	// }

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer slurm.Close()
	slurm.WriteString("#!/bin/bash\n")
	slurm.WriteString(
		"#SBATCH --job-name=" + job.Metadata.Namespace + "_" + job.Metadata.Name + "\n",
	)
	slurm.WriteString("#SBATCH --output=" + job.Spec.OutputFile + "\n")
	slurm.WriteString("#SBATCH --error=" + job.Spec.ErrorFile + "\n")
	slurm.WriteString("#SBATCH --partition=" + job.Spec.Partition + "\n")
	slurm.WriteString("#SBATCH -n " + job.Spec.NTasks + "\n")
	slurm.WriteString(
		"#SBATCH --ntasks-per-node=" + job.Spec.NTasksPerNode + "\n",
	)
	slurm.WriteString("#SBATCH --gres=gpu:" + job.Spec.GPUNums + "\n")
	slurm.WriteString("#SBATCH --cpus-per-task=" + job.Spec.CPUPerTask + "\n")

	for _, v := range job.Spec.RunCommands {
		slurm.WriteString(v + "\n")
	}

	//构建dockerfile
	fmt.Println("开始构建dockerfile")
	dockerfile, err := os.Create(JobFilePath + "/Dockerfile")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dockerfile.Close()
	dockerfile.WriteString(
		"FROM " + apiserver.APIServerIp + ":5000/job-server:latest\n",
	)
	dockerfile.WriteString("ENV OUTPUT_FILE " + job.Spec.OutputFile + "\n")
	dockerfile.WriteString("ENV ERROR_FILE " + job.Spec.ErrorFile + "\n")
	dockerfile.WriteString(
		"ENV API_SERVER_IP&PORT " + apiserver.APIServerUrl + "\n",
	)
	// dockerfile.WriteString(
	// 	"ENV API_SERVER_PORT " + constant.API_SERVER_PORT + "\n",
	// )
	dockerfile.WriteString("ENV JOB_NAME " + job.Metadata.Name + "\n")
	dockerfile.WriteString("ENV JOB_NAMESPACE " + job.Metadata.Namespace + "\n")
	dockerfile.WriteString("COPY ./" + job.Metadata.Name + " /app/func\n")

	//构建docker上下文，需要将依赖文件打包成tar格式
	fmt.Println("开始打包tar")
	z2 := archiver.NewTar()
	z2.OverwriteExisting = true
	err = z2.Archive([]string{JobFilePath}, JobFilePath+"/job.tar")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ctx, err := os.Open(JobFilePath + "/job.tar")
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
	defer cli.Close()

	resp, err := cli.ImageBuild(
		context.Background(),
		ctx,
		types.ImageBuildOptions{
			Dockerfile: job.Metadata.Name + "/Dockerfile",
			Tags: []string{
				fmt.Sprintf(
					"%s:5000/job-server/%s/%s:latest",
					apiserver.APIServerIp,
					job.Metadata.Namespace,
					job.Metadata.Name,
				),
			},

			Context: ctx,
			Remove:  true,
		},
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
	fmt.Println("开始推送镜像")
	authEncoded := base64.StdEncoding.EncodeToString(
		[]byte(registry_user + ":" + registry_password),
	)
	fmt.Println(authEncoded)
	resp2, err := cli.ImagePush(
		context.Background(),
		apiserver.APIServerIp+":5000/job-server/"+job.Metadata.Namespace+"/"+job.Metadata.Name+":latest",
		image.PushOptions{
			RegistryAuth: authEncoded,
			All:          false,
		},
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp2.Close()
	io.Copy(os.Stdout, resp2)

	//创建pod
	fmt.Println("开始创建pod")
	var pod object.Pod
	pod.Kind = "Pod"
	pod.Metadata.Name = "job_pod_" + job.Metadata.Name
	pod.Metadata.Namespace = job.Metadata.Namespace
	pod.Spec.Containers = make([]object.Container, 0)
	pod.Spec.Containers = append(
		pod.Spec.Containers,
		object.Container{
			Name:  "job_container_" + job.Metadata.Name,
			Image: apiserver.APIServerIp + ":5000/job-server/" + job.Metadata.Namespace + "/" + job.Metadata.Name,
		},
	)

	fc.ci.CreatePod(&pod)
}
