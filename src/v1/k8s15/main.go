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
	clientset *kubernetes.Clientset
	config    *rest.Config
	err       error
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
	var namespace = flag.String("ns", "", "-ns=default")
	var deploymentName = flag.String("dep", "", "-dep=deployment-name")

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

	for k := range podMetrics.Items {
		containerName := podMetrics.Items[k].Metadata.Name

		for key := range podMetrics.Items[k].Containers {
			// Convert CPU readings
			cpuInt, cpuUnit, err := parseCPUReading(podMetrics.Items[k].Containers[key].Usage.CPU)
			cpuConverted, friendlyUnit := convertCPUWrapper(cpuInt, cpuUnit)

			if err != nil {
				log.Println("Received error parsing CPU")
			}
			// Convert memory readings
			memoryInt, memoryUnit, err := parseMemoryReading(podMetrics.Items[k].Containers[key].Usage.Memory)

			memInMibi := convertMemoryToMibiWrapper(memoryInt, memoryUnit)

			if err != nil {
				log.Println("Received error parsing memory")
			}
			log.Println("Container", containerName, "is using", memInMibi, "Mib memory and", cpuConverted, friendlyUnit, "CPU")

		}
	}

	deploymentInfo, err := getDeploymentInfo(*namespace, *deploymentName)

	if err != nil {
		log.Println("Error gathering deployment metrics", err.Error())
	}

	for k := range deploymentInfo.Spec.Template.Spec.Containers {
		log.Println("CPU limit is", parseCPULimit(deploymentInfo.Spec.Template.Spec.Containers[k].Resources.Limits.CPU), "milliCPU")
		log.Println("Memory limit is", parseMemoryLimit(deploymentInfo.Spec.Template.Spec.Containers[k].Resources.Limits.Memory), "Mibibytes")
	}
}
