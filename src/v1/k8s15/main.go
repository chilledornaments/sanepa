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
	scaleDownOkCount          int
	shouldScaleDown           bool
	scaleDownOkPeriods        *int
	hasScaled                 bool
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz"
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
	flag.Parse()

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

	podMetrics, err := getPodMetrics(*namespace)

	if err != nil {
		logError("Error gathering pod metrics", err)
	}

	deploymentInfo, err := getDeploymentInfo(*namespace, *deploymentName)

	if err != nil {
		logError("Error gathering deployment metrics", err)
	}

	logInfo(fmt.Sprintf("Minimum replicas is %d", *deploymentMinReplicas))
	logInfo(fmt.Sprintf("Maximum replicas is %d", *deploymentMaxReplicas))
	logInfo(fmt.Sprintf("Cooldown period is %d seconds", *cooldownInSeconds))
	logInfo(fmt.Sprintf("Scale down ok periods is %d seconds", *scaleDownOkPeriods))

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
				// Convert CPU readings
				cpuInt, cpuUnit, err := parseCPUReading(podMetrics.Items[k].Containers[key].Usage.CPU)
				cpuConverted, friendlyUnit := convertCPUWrapper(cpuInt, cpuUnit)

				if podMetrics.Items[k].Containers[key].Name != containerNameToMatch {
					logInfo(fmt.Sprintf("Skipping %s as it is not a member of the deployment %s", podMetrics.Items[k].Containers[key].Name, string(*deploymentName)))
				} else {

					if err != nil {
						logInfo("Received error parsing CPU")
					}
					// Convert memory readings
					memoryInt, memoryUnit, err := parseMemoryReading(podMetrics.Items[k].Containers[key].Usage.Memory)

					memInMibi := convertMemoryToMibiWrapper(memoryInt, memoryUnit)

					if err != nil {
						logError("Received error parsing memory", err)
					}
					logInfo(fmt.Sprintf("Container %s is using %d Mib memory and %d %s", containerName, memInMibi, cpuConverted, friendlyUnit))

					if memInMibi > deploymentMemoryThreshold {
						logWarning(fmt.Sprintf("Container %s is over the memory limit. Adding another replica", containerName))
						shouldScaleUp = true
					} else if cpuConverted > deploymentCPUThreshold {
						logWarning(fmt.Sprintf("Container %s is over the CPU limit. Adding another replica", containerName))
						shouldScaleUp = true
					} else {
						logInfo("Containers are below thresholds")
						scaleDownOkCount++
						if scaleDownOkCount >= *scaleDownOkPeriods && hasScaled {
							logInfo("Attempting to scale down by one replica")
							err = scaleDownDeployment(*namespace, *deploymentName)
							if err == errScalingLimitReached {
								hasScaled = false
							}
							time.Sleep(30 * time.Second)
						}
						shouldScaleUp = false
					}
					if shouldScaleUp {
						logInfo("Scaling started")
						err = scaleUpDeployment(*namespace, *deploymentName)
						if err != nil {
							hasScaled = false
						}
						logInfo(fmt.Sprintf("Waiting %d seconds for cooldown", *cooldownInSeconds))
						time.Sleep(time.Duration(*cooldownInSeconds) * time.Second)
						shouldScaleUp = false
						hasScaled = true
					}

				}
			}
		}
	}
}
