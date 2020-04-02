package main

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func scaleUpDeployment(namespace string, deploymentName string) error {

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	o := metav1.GetOptions{}

	d, err := deploymentsClient.GetScale(deploymentName, o)

	if err != nil {
		panic(err)
	}

	if int(d.Spec.Replicas) >= *deploymentMaxReplicas {
		logWarning("Scaling upper limit reached. Not scaling up")
		return errScalingLimitReached
	}

	newReplicas := d.Spec.Replicas + 1

	d.Spec.Replicas = newReplicas

	_, err = deploymentsClient.UpdateScale(deploymentName, d)

	if err != nil {
		logError(fmt.Sprintf("Received error when attempting to scale up to %d replicas", newReplicas), err)
		return err
	}

	logScaleEvent(fmt.Sprintf("Successfully scaled %s up to %d replicas", deploymentName, newReplicas))
	return nil

}

func scaleDownDeployment(namespace string, deploymentName string) error {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	o := metav1.GetOptions{}

	d, err := deploymentsClient.GetScale(deploymentName, o)

	if err != nil {
		panic(err)
	}

	if int(d.Spec.Replicas) <= *deploymentMinReplicas {
		logWarning("Scaling lower limit reached. Not scaling down")
		return errScalingLimitReached
	}

	newReplicas := d.Spec.Replicas - 1

	d.Spec.Replicas = newReplicas

	_, err = deploymentsClient.UpdateScale(deploymentName, d)

	if err != nil {
		logError(fmt.Sprintf("Received error when attempting to scale down to %d replicas", newReplicas), err)
		return err
	}

	logScaleEvent(fmt.Sprintf("Successfully scaled %s down to %d replicas", deploymentName, newReplicas))
	return nil

}
