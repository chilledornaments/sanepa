package main

import (
	"encoding/json"
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

type metricsStruct struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink string `json:"selfLink"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
		Window     string `json:"window"`
		Containers []struct {
			Name  string `json:"string"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

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

	err = getPodMetrics(*namespace)

	if err != nil {
		panic(err)
	}

}

func getPodMetrics(namespace string) error {

	var podURL string

	if namespace == "" {
		podURL = "apis/metrics.k8s.io/v1beta1/pods"
	} else {
		podURL = fmt.Sprintf("apis/metrics.k8s.io/v1beta1/namespaces/%s/pods", namespace)
	}

	pods := &metricsStruct{}

	data, err := clientset.RESTClient().Get().AbsPath(podURL).DoRaw()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &pods)

	if err != nil {
		return err
	}

	log.Println(pods)
	return nil
}
