package main

import (
	"fmt"
	"log"
	"strconv"
)

func parseCPUReading(cpu string) (int, string, error) {
	logDebug(fmt.Sprintf("parseCPUReading received %s", cpu))
	// We want the last item in the string
	unit := string(cpu[len(cpu)-1])
	cpuStr := cpu[0 : len(cpu)-1]
	cpuInt, err := strconv.Atoi(cpuStr)

	if err != nil {
		logError(fmt.Sprintf("Unable to convert %s to int", cpuStr), err)
		return 0, "", err
	}

	return cpuInt, unit, nil
}

func convertCPUWrapper(cpuUsage int, cpuUnit string) (int, string) {
	switch cpuUnit {
	case "n":
		return convertNanoToMilli(cpuUsage), "milliCPU"
	case "m":
		return cpuUsage, "milliCPU"
	case "u":
		return convertMicroToMilli(cpuUsage), "milliCPU"
	default:
		logDebug(fmt.Sprintf("Encountered unknown CPU unit %s", cpuUnit))
		return cpuUsage, "UNKNOWN"
	}
}

func parseCPULimit(cpuLimit string) int {
	var wrappedCPULimit int
	wrappedCPULimit, err = strconv.Atoi(cpuLimit)

	if err != nil {
		log.Println("Error converting", cpuLimit, "to int. Assuming CPU limit specifies millicpu")
		cpuStr := cpuLimit[0 : len(cpuLimit)-1]
		wrappedCPULimit, err = strconv.Atoi(cpuStr)
	} else {
		wrappedCPULimit = wrappedCPULimit * 1000
	}

	return wrappedCPULimit
}

func convertNanoToMilli(cpuUsage int) int {
	return cpuUsage / 1000000
}

func convertMicroToMilli(cpuUsage int) int {
	return cpuUsage / 1000
}

func parseMemoryReading(memory string) (int, string, error) {
	// Second from last item in the string should be e.g. K/M/G
	unit := string(memory[len(memory)-2])
	memoryStr := memory[0 : len(memory)-2]
	memoryInt, err := strconv.Atoi(memoryStr)
	if err != nil {
		logError(fmt.Sprintf("Unable to convert %s to int", memoryStr), err)
		return 0, "", err
	}

	return memoryInt, unit, nil
}

func parseMemoryLimit(memoryLimit string) int {
	limit, unit, err := parseMemoryReading(memoryLimit)

	if err != nil {
		logError("Error parsing memory limit", err)
		// Placeholder, we shouldn't return 0 because we will always scale
		// We shouldn't return 1000 cause then we'll probably never scale
		return 1000
	}

	return convertMemoryToMibiWrapper(limit, unit)

}

func convertMemoryToMibiWrapper(memoryUsage int, memoryType string) int {
	switch memoryType {
	case "K":
		return memoryUsage / 1024
	case "M":
		return memoryUsage
	case "G":
		return memoryUsage * 1024
	default:
		logDebug(fmt.Sprintf("Received unknown memory unit %s", memoryType))
		// This is probably wrong
		return memoryUsage
	}
}

func generateThreshold(limit int, threshold int) int {
	return limit * threshold / 100
}

func resetScalingCounters() {
	shouldScaleDownCounter = 0
	shouldScaleUpCounter = 0
}
