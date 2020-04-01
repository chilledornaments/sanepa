package main

import (
	"log"
)

func logInfo(message string) {
	log.Println("INFO:", message)
}

func logWarning(message string) {
	log.Println("WARNING:", message)
}

func logError(message string, e error) {
	log.Println("ERROR:", message, e.Error())
}

func logScaleEvent(message string) {
	log.Println("EVENT:", message)
}
