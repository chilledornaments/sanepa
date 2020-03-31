package main

import (
	"encoding/json"
	"fmt"
)

func getPodMetrics(namespace string) (*podMetricsStruct, error) {

	var url string

	if namespace == "" {
		url = "apis/metrics.k8s.io/v1beta1/pods"
	} else {
		url = fmt.Sprintf("apis/metrics.k8s.io/v1beta1/namespaces/%s/pods", namespace)
	}

	podMetrics := &podMetricsStruct{}

	data, err := clientset.RESTClient().Get().AbsPath(url).DoRaw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &podMetrics)

	if err != nil {
		return nil, err
	}

	return podMetrics, nil
}

func getDeploymentInfo(namespace string, deploymentName string) (*deploymentStruct, error) {
	var url string

	if namespace == "" {
		url = "apis/extensions/v1beta1/deployments"
	} else {
		url = fmt.Sprintf("apis/extensions/v1beta1/namespaces/%s/deployments/%s", namespace, deploymentName)
	}

	deploymentMetrics := &deploymentStruct{}

	data, err := clientset.RESTClient().Get().AbsPath(url).DoRaw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &deploymentMetrics)

	if err != nil {
		return nil, err
	}

	return deploymentMetrics, nil
}
