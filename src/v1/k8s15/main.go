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
	graylogEnabled = flag.Bool("gl-enabled", false, "Enable logging to Graylog. Example: -gl-enabled=true")
	graylogUDPWriter = flag.String("gl-server", "", "IP:PORT of Graylog server. UDP only. Required if -gl-enabled=true. Example: -gl-server=10.10.5.44:11411")

	flag.Parse()

	if *graylogEnabled {
		initGraylog()
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

	watcher := time.Tick(10 * time.Second)

	for range watcher {
		monitorAndScale()
	}

}

func monitorAndScale() {

	fmt.Println("********************************************************")

	currentContainerCount := 0

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
							logInfo(fmt.Sprintf("Skipping %s as it is not a member of the deployment %s", podMetrics.Items[k].Containers[key].Name, string(*deploymentName)))
						} else {

							// Convert CPU readings
							cpuInt, cpuUnit, err := parseCPUReading(podMetrics.Items[k].Containers[key].Usage.CPU)
							cpuConverted, friendlyUnit := convertCPUWrapper(cpuInt, cpuUnit)

							if err != nil {
								logError("Received error parsing CPU", err)
							}
							// Convert memory readings
							memoryInt, memoryUnit, err := parseMemoryReading(podMetrics.Items[k].Containers[key].Usage.Memory)

							memInMibi := convertMemoryToMibiWrapper(memoryInt, memoryUnit)

							if err != nil {
								logError("Received error parsing memory", err)
							}

							logInfo(fmt.Sprintf("Container %s is using %d Mib memory and %d %s", containerName, memInMibi, cpuConverted, friendlyUnit))

							if memInMibi >= deploymentMemoryThreshold {
								logWarning(fmt.Sprintf("Container %s is over the memory limit. Scale up trigger count is %d", containerName, scaleUpOkCount))
								scaleUpOkCount++
								shouldScaleUp = true
							} else if cpuConverted >= deploymentCPUThreshold {
								logWarning(fmt.Sprintf("Container %s is over the CPU limit. Scale up trigger count is %d", containerName, scaleUpOkCount))
								scaleUpOkCount++
								shouldScaleUp = true
							} else {
								logInfo("Containers are below thresholds")
								if hasScaled {
									scaleDownOkCount++
								}
								if (scaleDownOkCount >= *scaleDownOkPeriods) && hasScaled {
									logDebug(fmt.Sprintf("scaleDownOkCount: %d scaleDownOkPeriods: %d", scaleDownOkCount, *scaleDownOkPeriods))
									logInfo("Attempting to scale down by one replica")
									err = scaleDownDeployment(*namespace, *deploymentName)
									if err == errScalingLimitReached {
										hasScaled = false
									}
									scaleDownOkCount = 0
									// We've scaled down, reset hasScaled
									hasScaled = false
									time.Sleep(time.Duration(*cooldownInSeconds) * time.Second)
								}
								shouldScaleUp = false
							}

							if shouldScaleUp && (scaleUpOkCount >= *scaleUpOkPeriods) {
								logInfo("Scale up started")
								err = scaleUpDeployment(*namespace, *deploymentName)
								if err != nil {
									hasScaled = false
								}
								logInfo(fmt.Sprintf("Waiting %d seconds for cooldown", *cooldownInSeconds))
								time.Sleep(time.Duration(*cooldownInSeconds) * time.Second)
								shouldScaleUp = false
								hasScaled = true
								scaleUpOkCount = 0
								scaleDownOkCount = 0
							}

						}
					}
				}
			}
		}
	}
}
