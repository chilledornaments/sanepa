package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

func authInCluster() error {
	config, err := rest.InClusterConfig()

	if err != nil {
		return err
	}

	clientset, err = kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}

	return nil
}

func main() {

	shouldScaleUp = false
	scaleDownOkCount = 0
	scaleUpOkCount = 0
	hasScaled = false
	shouldScaleUpCounter = 0
	shouldScaleDownCounter = 0
	thresholdBreachesCounter = 0
	metricState = make(map[string]metricReadings)

	inCluster = flag.Bool("incluster", true, "-incluster=false to run outside of a k8s cluster")
	namespace = flag.String("ns", "", "Namespace to search in. Example: -ns=default")
	deploymentName = flag.String("dep", "", "Deployment name to watch. Example: -dep=deployment-name")
	cpuThreshold = flag.Int("cpu", 50, "At what percentage of CPU limit should we scale? Example: -cpu=40")
	memThreshold = flag.Int("mem", 70, "At what percentage of memory limit should we scale? Example: -mem=30")
	deploymentMaxReplicas = flag.Int("max", 5, "The maximum number of replicas the deployment can have. Example: -max=5")
	deploymentMinReplicas = flag.Int("min", 1, "The minimum number of replicas in the deployment. Example: -min=2")
	cooldownInSeconds = flag.Int("cooldown", 30, "The number of seconds to wait after scaling. If your application takes 120 seconds to become ready, set this to 120. Example: -cooldown=10")
	scaleDownOkPeriods = flag.Int("scaledownok", 3, "For how many consecutive periods of time must the containers be under threshold until we scale down? Example: -scaledownok=3")
	scaleUpOkPeriods = flag.Int("scaleupok", 2, "How many consecutive periods must be a pod be above threshold before scaling?. Example -scaleupok=5")
	breachpercentthreshold = flag.Int("breachpercentthreshold", 50, "What percentage of pods must be in breaching state before scaling up?")
	graylogEnabled = flag.Bool("gl-enabled", false, "Enable logging to Graylog. Example: -gl-enabled=true")
	graylogUDPWriter = flag.String("gl-server", "", "IP:PORT of Graylog server. UDP only. Required if -gl-enabled=true. Example: -gl-server=10.10.5.44:11411")

	flag.Parse()

	if *graylogEnabled {
		initGraylog()
	} else {
		logInfo("Not logging to Graylog")
	}

	logInfo("Starting SanePA")

	if *inCluster {
		logInfo("Running in cluster")
		err = authInCluster()
		if err != nil {
			logError("Error authenticating in cluster", err)
			os.Exit(1)
		}
	} else {
		logInfo("Running outside of cluster")
		err = authOutCluster()
		if err != nil {
			logError("Error authenticating outside of cluster", err)
			os.Exit(1)
		}
	}

	watcher := time.Tick(40 * time.Second)

	for range watcher {
		monitorAndScale()
	}

}

func monitorAndScale() {

	fmt.Println("********************************************************")

	currentContainerCount := 0
	metricParseError = false

	podMetrics, err := getPodMetrics(*namespace)

	if err != nil {
		logError("Error gathering pod metrics. Backing off for 10 seconds. Error:", err)
		time.Sleep(10)
	} else {

		deploymentInfo, err := getDeploymentInfo(*namespace, *deploymentName)

		if err != nil {
			logError("Error gathering deployment metrics. Backing off for 10 seconds. Error:", err)
			time.Sleep(10)
		} else {

			logInfo(fmt.Sprintf("Minimum replicas is %d", *deploymentMinReplicas))
			logInfo(fmt.Sprintf("Maximum replicas is %d", *deploymentMaxReplicas))
			logInfo(fmt.Sprintf("Cooldown period is %d seconds", *cooldownInSeconds))
			logInfo(fmt.Sprintf("Scale down ok periods is %d", *scaleDownOkPeriods))
			logInfo(fmt.Sprintf("Scale up ok periods is %d", *scaleUpOkPeriods))

			for k := range deploymentInfo.Spec.Template.Spec.Containers {
				deploymentCPULimit = parseCPULimit(deploymentInfo.Spec.Template.Spec.Containers[k].Resources.Limits.CPU)
				deploymentCPUThreshold = generateThreshold(deploymentCPULimit, *cpuThreshold)
				deploymentMemoryLimit = parseMemoryLimit(deploymentInfo.Spec.Template.Spec.Containers[k].Resources.Limits.Memory)
				deploymentMemoryThreshold = generateThreshold(deploymentMemoryLimit, *memThreshold)
				containerNameToMatch = deploymentInfo.Spec.Template.Spec.Containers[k].Name
				logInfo(fmt.Sprintf("CPU limit is %d milliCPU for deployment: %s", deploymentCPULimit, deploymentInfo.Spec.Template.Spec.Containers[k].Name))
				logInfo(fmt.Sprintf("Scaling CPU threshold is %d milliCPU", deploymentCPUThreshold))
				logInfo(fmt.Sprintf("Memory limit is %d Mibibytes for deployment: %s percent", deploymentMemoryLimit, deploymentInfo.Spec.Template.Spec.Containers[k].Name))
				logInfo(fmt.Sprintf("Scaling memory threshold is %d mibibytes", deploymentMemoryThreshold))

				for k := range podMetrics.Items {
					containerName := podMetrics.Items[k].Metadata.Name

					for key := range podMetrics.Items[k].Containers {
						// Increment counter
						currentContainerCount++

						// Handle pods in same namespace from different deployment
						if podMetrics.Items[k].Containers[key].Name != containerNameToMatch {
							logInfo(fmt.Sprintf("Skipping %s as it is not a member of the deployment %s", podMetrics.Items[k].Containers[key].Name, *deploymentName))
						} else {
							err := storeMetricData(containerName, podMetrics.Items[k].Containers[key].Usage.CPU, podMetrics.Items[k].Containers[key].Usage.Memory)
							if err != nil {
								metricParseError = true
							} else {
								metricParseError = false
							}
						}
					}
				}
			}
			if !metricParseError {
				checkMetricThresholds()
				checkIfShouldScale()
			}

		}
	}
}

func storeMetricData(containerName string, cpu string, memory string) error {
	cpuInt, s, err := parseCPUReading(cpu)
	if err != nil {
		logError("Error parsing CPU reading", err)
		return err
	}

	c, _ := convertCPUWrapper(cpuInt, s)

	memInt, s, err := parseMemoryReading(memory)
	if err != nil {
		logError("Error parsing memory reading", err)
		return err
	}

	m := convertMemoryToMibiWrapper(memInt, s)
	mr := metricReadings{CPU: c, Memory: m}

	metricState[containerName] = mr
	return nil
}

func checkMetricThresholds() {

	for k, v := range metricState {
		if v.CPU > deploymentCPUThreshold {
			logInfo(fmt.Sprintf("%s is breaching CPU: %d mCPU used", k, v.CPU))
			thresholdBreachesCounter++
		} else if v.Memory > deploymentMemoryThreshold {
			logInfo(fmt.Sprintf("%s is breaching memory: %d MiB used", k, v.Memory))
			thresholdBreachesCounter++
		} else {
			logInfo(fmt.Sprintf("%s is not breaching", k))
			if thresholdBreachesCounter > 0 {
				thresholdBreachesCounter--
			}

		}
	}
}

func checkIfShouldScale() bool {
	/*
		If we have 15 containers running and 4 are breaching, 26% of the containers are breaching
		If this percent is greater than the breachpercentthreshold, we'll increment the shouldScaleUpCounter counter
		If the shouldScaleUpCounter counter is greater than the scaleupok value, we'll scale up
	*/
	logInfo(fmt.Sprintf("ScaleDownCounter=%d ScaleUpCounter=%d NumberOfBreachingContainers=%d NumberOfPods=%d", shouldScaleDownCounter, shouldScaleUpCounter, thresholdBreachesCounter, len(metricState)))

	// Enough containers have been breaching for long enough for us to scale up
	if shouldScaleUpCounter >= *scaleUpOkPeriods {
		logScaleEvent("shouldScaleUpCounter has hit threshold. Attempting to add another replica, resetting scale up counter, and entering cooldown")
		if err := scaleUpDeployment(*namespace, *deploymentName); err != nil {
			return true
		}
		shouldScaleUpCounter = 0
		time.Sleep(time.Duration(*cooldownInSeconds) * time.Second)
		return true
	}

	// Containers have been below thresholds long enough to scale down
	if shouldScaleDownCounter >= *scaleDownOkPeriods {
		logScaleEvent("Reached scale down threshold. Attempting to remove a replica, resetting scale down counter, and entering cooldown")
		if err := scaleDownDeployment(*namespace, *deploymentName); err != nil {
			return true
		}
		shouldScaleDownCounter = 0
		time.Sleep(time.Duration(*cooldownInSeconds) * time.Second)
		return true
	}

	// No container are breaching
	if thresholdBreachesCounter == 0 {
		shouldScaleDownCounter++
		logInfo(fmt.Sprintf("No containers are breaching. Incrementing scale down counter. Counter is now at: %d", shouldScaleDownCounter))
		return false
	}

	breachPercent := (float64(thresholdBreachesCounter) / float64(len(metricState))) * 100

	// Every container is breaching
	if breachPercent == 100 {
		shouldScaleUpCounter++
		logInfo(fmt.Sprintf("All containers are breaching thresholds. Incrementing scale up counter. Counter is now at: %d", shouldScaleUpCounter))
		thresholdBreachesCounter = 0
		return true
	} else if breachPercent >= float64(*breachpercentthreshold) {
		shouldScaleUpCounter++
		logInfo(fmt.Sprintf("Percent of breaching containers passed threshold. Incrementing scale up counter. Breach percent: %g", breachPercent))
		thresholdBreachesCounter = 0
		return true
	} else {
		logInfo(fmt.Sprintf("Breaching container percent is below %g percent threshold", breachPercent))
		return false
	}

}
