package main

import (
	"fmt"
	"log"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	log.Println("Starting sanepa")

	if os.Getenv("KUBECONFIG") == "" {
		panic("KUBECONFIG not set in env")
	}
	kubeconfig := os.Getenv("KUBECONFIG")

	log.Println("Using", kubeconfig, "as config file")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		panic(err)
	}

	clientset, _ := kubernetes.NewForConfig(config)
	pods, _ := clientset.CoreV1().Pods("").List(v1.ListOptions{})
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

}
