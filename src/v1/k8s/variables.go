package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	numberBreachingContainers int
	metricState               map[string]metricReadings
	metricParseError          bool
	listenPort                *int

	scaleUpPromCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sanepa_scale_up_events",
		Help: "Total number of times SanePA has scaled a deployment up",
	})

	scaleDownPromCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sanepa_scale_down_events",
		Help: "Total number of times SanePA has scaled a deployment down",
	})

	scaleDownErrPromCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sanepa_scale_down_errors",
		Help: "Total number of times SanePA has tried and failed to scale a deployment down",
	})

	scaleUpErrPromCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sanepa_scale_up_errors",
		Help: "Total number of times SanePA has tried and failed to scale a deployment up",
	})

	collectionErrPromCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sanepa_metric_collection_errors",
		Help: "Total number of times SanePA has failed to collect metrics or retrieve depoyment info",
	})
)
