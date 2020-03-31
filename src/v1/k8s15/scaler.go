package main

func scaleUpDeployment(namespace string) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
}
