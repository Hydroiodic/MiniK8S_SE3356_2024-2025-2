package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/mholt/archiver"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var execCmd = &cobra.Command{
	Use:     "apply",
	Short:   "Create a Kubernetes instance from a YAML file",
	Example: "kubectl apply -f config.yaml",
	Run: func(cmd *cobra.Command, _ []string) {
		// Get the flag value for `-f` or `--file`.
		fileFlag, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Printf("Error retrieving flag value: %v\n", err)
			return
		}

		// If the flag not set or empty, print a message.
		if fileFlag == "" {
			fmt.Println("Usage: kubectl apply -f <yaml-file>")
			return
		}

		// Parse the YAML file and handle the instances.
		fmt.Printf("Using file: %s\n", fileFlag)
		parseYaml(fileFlag)
	},
}

// Add the `-f` flag to the execCmd command.
func init() {
	execCmd.Flags().StringP("file", "f", "", "Specify the configuration file")
	rootCmd.AddCommand(execCmd)
}

// parseYaml reads a YAML file and processes its contents based on the `kind` field.
func parseYaml(fileAddr string) {
	// Read the YAML file from the file system.
	data, err := os.ReadFile(fileAddr)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", fileAddr, err)
		return
	}

	// Unmarshal the YAML data into a struct to extract the `kind` field.
	err = yaml.Unmarshal(data, &kindStruct)
	if err != nil {
		fmt.Printf("Error unmarshaling YAML: %v\n", err)
		return
	}

	// Check the kind and call the appropriate handler.
	switch kindStruct.Kind {
	case PodKind:
		err = handlePodRaw(data)
	case ServiceKind:
		err = handleServiceRaw(data)
	case ReplicaSetKind:
		err = handleReplicaSetRaw(data)
	case DNSKind:
		err = handleDNSConfigRaw(data)
	case HorizontalPodAutoscalerKind:
		err = handleHPARaw(data)
	case GPUJobKind:
		err = handleGPUJOBRaw(data)
	case Function:
		err = handleFunctionRaw(data)
	case Workflow:
		err = handleWorkflowRaw(data)
	case Event:
		err = handleEventRaw(data)
	default:
		fmt.Printf("Unsupported resource kind: %s\n", kindStruct.Kind)
		return
	}

	// If any error occurred during processing, print it.
	if err != nil {
		fmt.Printf(
			"Error occurred during applying %s: %v\n",
			kindStruct.Kind,
			err,
		)

		return
	}
}

func handleEventRaw(rawData []byte) error {
	// Parse the raw YAML data into a Pod object.
	var e object.Event
	if err := yaml.Unmarshal(rawData, &e); err != nil {
		return err
	}

	// Add the Pod configuration.
	err := apiserver.NewAPIClient("").CreateEvent(e)
	if err != nil {
		return err
	}

	return nil
}
func handleWorkflowRaw(rawData []byte) error {
	// Parse the raw YAML data into a Pod object.
	var wf object.Workflow
	if err := yaml.Unmarshal(rawData, &wf); err != nil {
		return err
	}

	// Add the Pod configuration.
	err := apiserver.NewAPIClient("").CreateWorkflow(wf)
	if err != nil {
		return err
	}

	return nil
}

func handleFunctionRaw(rawData []byte) error {
	// Parse the raw YAML data into a GpuJob object.
	var f object.Function
	if err := yaml.Unmarshal(rawData, &f); err != nil {
		return err
	}

	// Check if the upload path exists.
	_, err := os.Stat(f.Spec.UserUploadPath)
	if err != nil {
		return err
	}

	// Get every file in the upload path and compress them into a ZIP file.
	z := archiver.NewZip()
	z.OverwriteExisting = true

	// Get all files in the specified upload path.
	files, err := filepath.Glob(filepath.Join(f.Spec.UserUploadPath, "*"))
	if err != nil {
		return err
	}

	// Compress the files into a ZIP archive.
	err = z.Archive(files, f.Spec.UserUploadPath+".zip")
	if err != nil {
		return err
	}

	// Read the ZIP file into a byte slice.
	fileByte, err := os.ReadFile(f.Spec.UserUploadPath + ".zip")
	if err != nil {
		return err
	}

	f.Spec.UserUploadFile = fileByte
	fmt.Println(f)

	err = apiserver.NewAPIClient("").CreateFunction(f)
	if err != nil {
		return err
	}

	return nil
}

func handleGPUJOBRaw(rawData []byte) error {
	// Parse the raw YAML data into a GpuJob object.
	var s object.Job
	if err := yaml.Unmarshal(rawData, &s); err != nil {
		return err
	}

	// Check if the upload path exists.
	_, err := os.Stat(s.Spec.UploadPath)
	if err != nil {
		return err
	}

	// Get every file in the upload path and compress them into a ZIP file.
	z := archiver.NewZip()
	z.OverwriteExisting = true

	// Get all files in the specified upload path.
	files, err := filepath.Glob(filepath.Join(s.Spec.UploadPath, "*"))
	if err != nil {
		return err
	}

	// Compress the files into a ZIP archive.
	err = z.Archive(files, s.Spec.UploadPath+".zip")
	if err != nil {
		return err
	}

	// Read the ZIP file into a byte slice.
	fileByte, err := os.ReadFile(s.Spec.UploadPath + ".zip")
	if err != nil {
		return err
	}

	s.Spec.UserUploadFile = fileByte
	fmt.Println(s)

	err = apiserver.NewAPIClient("").CreateGPUJob(s)
	if err != nil {
		return err
	}

	return nil
}

func handlePodRaw(rawData []byte) error {
	// Parse the raw YAML data into a Pod object.
	var pod object.Pod
	if err := yaml.Unmarshal(rawData, &pod); err != nil {
		return err
	}

	// Add the Pod configuration.
	err := apiserver.NewAPIClient("").CreatePod(&pod)
	if err != nil {
		return err
	}

	return nil
}

func handleServiceRaw(rawData []byte) error {
	// Parse the raw YAML data into a Service object.
	var s object.Service
	if err := yaml.Unmarshal(rawData, &s); err != nil {
		return err
	}

	// Add the Service configuration.
	err := apiserver.NewAPIClient("").CreateService(&s)
	if err != nil {
		return err
	}

	return nil
}

func handleReplicaSetRaw(rawData []byte) error {
	// Parse the raw YAML data into a ReplicaSet object.
	var r object.ReplicaSet
	if err := yaml.Unmarshal(rawData, &r); err != nil {
		return err
	}

	// Add the ReplicaSet configuration.
	err := apiserver.NewAPIClient("").CreateReplicaset(&r)
	if err != nil {
		return err
	}

	return nil
}

func handleDNSConfigRaw(rawData []byte) error {
	// Parse the raw YAML data into a DNS object.
	var DNS object.DNS
	if err := yaml.Unmarshal(rawData, &DNS); err != nil {
		return err
	}

	// Add the DNS configuration.
	err := apiserver.NewAPIClient("").AddDNS(&DNS)
	if err != nil {
		return err
	}

	return nil
}

func handleHPARaw(rawData []byte) error {
	// Parse the raw YAML data into a HorizontalPodAutoscaler object.
	var hpa object.HorizontalPodAutoscaler
	if err := yaml.Unmarshal(rawData, &hpa); err != nil {
		return err
	}

	// Add the HorizontalPodAutoscaler configuration.
	err := apiserver.NewAPIClient("").CreateHpa(&hpa)
	if err != nil {
		return err
	}

	return nil
}
