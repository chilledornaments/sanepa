package main

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func scaleUpDeployment(namespace string, deploymentName string) error {

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	o := metav1.GetOptions{}

	d, err := deploymentsClient.GetScale(deploymentName, o)

	if err != nil {
		panic(err)
	}

	newReplicas := d.Spec.Replicas + 1

	d.Spec.Replicas = newReplicas

	_, err = deploymentsClient.UpdateScale(deploymentName, d)

	if err != nil {
		log.Println("Received error when attempting to scale to", newReplicas, "replicas")
		log.Println(err.Error())
		return err
	}

	log.Println("Successfully scaled", deploymentName, "to", newReplicas, "replicas")
	return nil

}
