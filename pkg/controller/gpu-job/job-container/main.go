package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/melbahja/goph"
)

// app下的文件是job-controller下的文件
// app/func下的文件是.cu和.slurm文件
func main() {
	//ssh连接交大slurm平台
	fmt.Println(goph.HasAgent())
	auth, err := goph.UseAgent()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := goph.NewUnknown(
		"stu1156",
		"pilogin.hpc.sjtu.edu.cn",
		auth,
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := cli.Close(); err != nil {
			fmt.Printf("Failed to close SSH client: %v\n", err)
		}
	}()
	//上传路径下的所有文件
	dirpath := "./func"
	entries, err := os.ReadDir(dirpath)

	if err != nil {
		panic(err)
	}

	slurmName := "/lustre/home/acct-stu/stu1156/"

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		//如果后缀不是.cu和.slurm,则跳过
		if entry.Name()[len(entry.Name())-3:] != ".cu" &&
			entry.Name()[len(entry.Name())-6:] != ".slurm" {
			continue
		}

		err := cli.Upload(
			dirpath+"/"+entry.Name(),
			"/lustre/home/acct-stu/stu1156/"+entry.Name(),
		)
		if err != nil {
			panic(err)
		}
		//找到后缀为.slurm的文件
		if entry.Name()[len(entry.Name())-6:] == ".slurm" {
			slurmName += entry.Name()
		}
	}
	// 运行slurm脚本
	res, err := cli.Run("sbatch " + slurmName)
	if err != nil {
		panic(err)
	}

	jobname := os.Getenv("JOB_NAME")
	fmt.Printf("Job name: %s\n", jobname)

	jobnamespace := os.Getenv("JOB_NAMESPACE")
	fmt.Printf("Job namespace: %s\n", jobnamespace)

	var jobID string
	_, err = fmt.Sscanf(string(res), "Submitted batch job %s", &jobID)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Job ID: %s\n", jobID)

	workdir := "/lustre/home/acct-stu/stu1156/"
	monitorJob(jobID, jobname, jobnamespace, workdir, *cli)
}

func monitorJob(jobID, jobname, jobnamespace, workdir string, cli goph.Client) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cmd := "sacct -j " + jobID + " | sed -n '3p' | awk '{print $6}'"
		fmt.Println(cmd)

		res, err := cli.Run(cmd)
		if err != nil {
			fmt.Println("Run command error:", err)
			return
		}

		status := strings.TrimSuffix(string(res), "\n")
		fmt.Printf("Job status: %s\n", status)

		switch status {
		case "COMPLETED":
			fmt.Println("Job completed")

			outputfile := os.Getenv("OUTPUT_FILE")
			// outputfile := "44636434.out"
			if err := cli.Download(workdir+outputfile, "/app/"+outputfile); err != nil {
				fmt.Println("Download error:", err)
				return
			}

			file, err := os.ReadFile("/app/" + outputfile)

			if err != nil {
				fmt.Println("Read file error:", err)
				return
			}

			postResult("output", jobID, jobname, jobnamespace, string(file))

			return
		case "CANCELLED+":
			fmt.Println("Job completed")

			// outputfile := os.Getenv("OUTPUT_FILE")
			outputfile := "44636434.out"
			if err := cli.Download(workdir+outputfile, "/app/"+outputfile); err != nil {
				fmt.Println("Download error:", err)
				return
			}

			file, err := os.ReadFile("/app/" + outputfile)

			if err != nil {
				fmt.Println("Read file error:", err)
				return
			}

			postResult("output", jobID, jobname, jobnamespace, string(file))

			return
		case "FAILED":
			fmt.Println("Job failed")

			errorfile := os.Getenv("ERROR_FILE")

			if err := cli.Download(workdir+errorfile, "/app/"+errorfile); err != nil {
				fmt.Println("Download error:", err)
				return
			}

			file, err := os.ReadFile("/app/" + errorfile)
			if err != nil {
				fmt.Println("Read error file error:", err)
				return
			}

			postResult("error", jobID, jobname, jobnamespace, string(file))

			return

		case "CANCELLED":
			fmt.Println("Job cancelled")
			return

		case "PENDING":
			fmt.Println("Job pending")
			// continue polling
		default:
			fmt.Println("Unknown job status:", status)
		}
	}
}

func postResult(resultType, jobID, jobname, jobnamespace, content string) {
	apiserverIp := os.Getenv("API_SERVER_IP&PORT")

	body := map[string]interface{}{
		"job_id":       jobID,
		"jobname":      jobname,
		"jobnamespace": jobnamespace,
	}
	if resultType == "output" {
		body["output"] = content
	} else {
		body["error"] = content
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		fmt.Println("JSON marshal error:", err)
		return
	}

	url := apiserverIp + "/gpujob/updateResult"

	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	if err != nil {
		fmt.Println("HTTP request error:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("HTTP do error:", err)
		return
	}
	defer resp.Body.Close()
}
