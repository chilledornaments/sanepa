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

	//pods, _ := clientset.CoreV1().Pods(*namespace).List(v1.ListOptions{})
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	podMetrics, err := getPodMetrics(*namespace)

	if err != nil {
		log.Println("Error gathering pod metrics", err.Error())
	}

	for k := range podMetrics.Items {
		containerName := podMetrics.Items[k].Metadata.Name

		for key := range podMetrics.Items[k].Containers {
			log.Println("Container", containerName, "is using", podMetrics.Items[k].Containers[key].Usage.CPU, "CPU and", podMetrics.Items[k].Containers[key].Usage.Memory, "memory")
		}
	}
	deploymentInfo, err = getDeploymentInfo(*namespace, *deploymentName)

	if err != nil {
		log.Println("Error gathering deployment metrics", err.Error())

	}

}
