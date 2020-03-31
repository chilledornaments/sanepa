package main

import (
	"log"
)

func getPodMetrics(namespace string) error {
	podMetricsList, err := clientset.MetricsV1beta1().PodMetricses(namespace).List(metav1.ListOptions{})

	log.Println(string(podMetricsList))

	return nil
}
