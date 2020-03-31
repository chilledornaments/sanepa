package main

import (
	"log"
	"strconv"
)

func parseCPUReading(cpu string) (int, string, error) {
	unit := string(cpu[len(cpu)-1])
	cpuStr := cpu[0 : len(cpu)-1]
	cpuInt, err := strconv.Atoi(cpuStr)

	if err != nil {
		log.Println("Unable to convert", cpuStr, "to int")
		log.Println(err.Error())
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
	}

	// This should never get called
	return cpuUsage, "milliCPU"
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
	converted := cpuUsage / 1000000
	return converted
}

func convertKibiToMibi() {

}
