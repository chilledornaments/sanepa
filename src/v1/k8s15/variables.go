package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clientset                 *kubernetes.Clientset
	config                    *rest.Config
	err                       error
	deploymentCPULimit        int
	deploymentCPUThreshold    int
	deploymentMemoryLimit     int
	deploymentMemoryThreshold int
	containerNameToMatch      string
	shouldScaleUp             bool
	scaleStatus               string
	namespace                 *string
	deploymentName            *string
	inCluster                 *bool
	cpuThreshold              *int
	memThreshold              *int
	deploymentMaxReplicas     *int
	deploymentMinReplicas     *int
	cooldownInSeconds         *int
	graylogEnabled            *bool
	scaleDownOkCount          int
	shouldScaleDown           bool
	scaleDownOkPeriods        *int
	scaleUpOkCount            int
	scaleUpOkPeriods          *int
	hasScaled                 bool
	graylogUDPWriter          *string
	breachpercentthreshold    *int
	shouldScaleUpCounter      int
	shouldScaleDownCounter    int
	// thresholdBreachesCounter is the number of containers that are breaching either CPU or memory thresholds
	thresholdBreachesCounter int
	metricState              map[string]metricReadings
	metricParseError         bool
)
