package main

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//scale "k8s.io/client-go/scale"
)

func scaleUpDeployment(namespace string, deploymentName string) {

	//var sg scale.ScalesGetter
	//scaleInterface := sg.Scales(namespace)

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	o := metav1.GetOptions{}

	d, err := deploymentsClient.GetScale(deploymentName, o)

	if err != nil {
		panic(err)
	}

	newReplicas := d.Spec.Replicas + 1

	d.Spec.Replicas = newReplicas

	scaleResult, err := deploymentsClient.UpdateScale(deploymentName, d)

	if err != nil {
		panic(err)
	}

	log.Println(scaleResult)
}
