package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset                 *kubernetes.Clientset
	config                    *rest.Config
	err                       error
	deploymentCPULimit        int
	deploymentCPUThreshold    int
	deploymentMemoryLimit     int
	deploymentMemoryThreshold int
	containerNameToMatch      string
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz"
)

func authOutCluster() error {
	if os.Getenv("KUBECONFIG") == "" {
		return fmt.Errorf("KUBECONFIG not set in env")
	}

	kubeconfig := os.Getenv("KUBECONFIG")

	log.Println("Using", kubeconfig, "as config file")

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		return err
	}

	clientset, _ = kubernetes.NewForConfig(config)

	return nil
}

func main() {

	var inCluster = flag.Bool("incluster", true, "-incluster=false to run outside of a k8s cluster")
	var namespace = flag.String("ns", "", "Namespace to search in. Example: -ns=default")
	var deploymentName = flag.String("dep", "", "Deployment name to watch. Example: -dep=deployment-name")
	var cpuThreshold = flag.Int("cpu", 50, "At what percentage of CPU limit should we scale? Example: -cpu=40")
	var memThreshold = flag.Int("mem", 70, "At what percentage of memory limit should we scale? Example: -mem=30")

	flag.Parse()

	log.Println("Starting SanePA")

	if *inCluster {
		log.Println("Running in cluster")
	} else {
		log.Println("Running outside of cluster")
		err = authOutCluster()
		if err != nil {
			log.Println("Error authenticating", err.Error())
		}
	}

	podMetrics, err := getPodMetrics(*namespace)

	if err != nil {
		log.Println("Error gathering pod metrics", err.Error())
	}

	deploymentInfo, err := getDeploymentInfo(*namespace, *deploymentName)

	if err != nil {
		log.Println("Error gathering deployment metrics", err.Error())
	}

	for k := range deploymentInfo.Spec.Template.Spec.Containers {
		deploymentCPULimit = parseCPULimit(deploymentInfo.Spec.Template.Spec.Containers[k].Resources.Limits.CPU)
		deploymentCPUThreshold = generateThreshold(deploymentCPULimit, *cpuThreshold)
		deploymentMemoryLimit = parseMemoryLimit(deploymentInfo.Spec.Template.Spec.Containers[k].Resources.Limits.Memory)
		deploymentMemoryThreshold = generateThreshold(deploymentMemoryLimit, *memThreshold)
		containerNameToMatch = deploymentInfo.Spec.Template.Spec.Containers[k].Name
		log.Println("CPU limit is", deploymentCPULimit, "milliCPU for deployment:", deploymentInfo.Spec.Template.Spec.Containers[k].Name)
		log.Println("Scaling CPU threshold is", deploymentCPUThreshold)
		log.Println("Memory limit is", deploymentMemoryLimit, "Mibibytes for deployment:", deploymentInfo.Spec.Template.Spec.Containers[k].Name)
		log.Println("Scaling memory threshold is", deploymentMemoryThreshold)

		for k := range podMetrics.Items {
			containerName := podMetrics.Items[k].Metadata.Name

			for key := range podMetrics.Items[k].Containers {
				// Convert CPU readings
				cpuInt, cpuUnit, err := parseCPUReading(podMetrics.Items[k].Containers[key].Usage.CPU)
				cpuConverted, friendlyUnit := convertCPUWrapper(cpuInt, cpuUnit)

				if podMetrics.Items[k].Containers[key].Name != containerNameToMatch {
					log.Println("Skipping", podMetrics.Items[k].Containers[key].Name, "as it is not a member of the deployment", *deploymentName)
				} else {

					if err != nil {
						log.Println("Received error parsing CPU")
					}
					// Convert memory readings
					memoryInt, memoryUnit, err := parseMemoryReading(podMetrics.Items[k].Containers[key].Usage.Memory)

					memInMibi := convertMemoryToMibiWrapper(memoryInt, memoryUnit)

					if err != nil {
						log.Println("Received error parsing memory")
					}
					log.Println("Container", containerName, "is using", memInMibi, "Mib memory and", cpuConverted, friendlyUnit)

					if memInMibi > deploymentMemoryThreshold {
						log.Println("ISSUE: Container", containerName, "is over the memory limit. Adding another replica")
					}
					if cpuConverted > deploymentCPUThreshold {
						log.Println("ISSUE: Container", containerName, "is over the CPU limit. Adding another replica")
					}
				}
			}
		}
	}
}
