package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

var (
	gelfWriter   *gelf.UDPWriter
	gelfHostname string
)

func initGraylog() {
	gelfHostname, _ = os.Hostname()
	gelfWriter, err = gelf.NewUDPWriter(*graylogUDPWriter)

	if err != nil {
		fmt.Println(err.Error())
	}
}

func wrapBuildGraylogMessage(shortMsg string, fullMsg string, level int32, extraFields map[string]interface{}) *gelf.Message {

	m := &gelf.Message{
		Version:  "1.1",
		Host:     gelfHostname,
		Short:    shortMsg,
		Full:     fullMsg,
		TimeUnix: float64(time.Now().Unix()),
		Level:    level,
		Extra:    extraFields,
	}

	return m
}

func sendGraylogMessage(m *gelf.Message) {
	err := gelfWriter.WriteMessage(m)

	if err != nil {
		log.Println("ERROR: Error sending message to Graylog", err.Error())
	}
}

func logInfo(message string) {
	log.Println("INFO:", message)

	if *graylogEnabled {
		_, file, line, _ := runtime.Caller(1)
		extra := map[string]interface{}{"file": file, "line": line, "deployment": *deploymentName}
		m := wrapBuildGraylogMessage("info", message, 6, extra)
		sendGraylogMessage(m)
	}
}

func logWarning(message string) {
	log.Println("WARNING:", message)

	if *graylogEnabled {
		_, file, line, _ := runtime.Caller(1)
		extra := map[string]interface{}{"file": file, "line": line, "deployment": *deploymentName}
		m := wrapBuildGraylogMessage("warning", message, 4, extra)
		sendGraylogMessage(m)
	}
}

func logError(message string, e error) {
	log.Println("ERROR:", message, e.Error())

	if *graylogEnabled {
		_, file, line, _ := runtime.Caller(1)
		extra := map[string]interface{}{"file": file, "line": line, "deployment": *deploymentName}
		m := wrapBuildGraylogMessage("error", fmt.Sprintf("%s err=%s", message, e.Error()), 3, extra)
		sendGraylogMessage(m)
	}
}

func logScaleEvent(message string) {
	log.Println("EVENT:", message)
	if *graylogEnabled {
		_, file, line, _ := runtime.Caller(1)
		extra := map[string]interface{}{"file": file, "line": line, "deployment": *deploymentName}
		m := wrapBuildGraylogMessage("scaleEvent", message, 5, extra)
		sendGraylogMessage(m)
	}
}

func logDebug(message string) {
	log.Println("DEBUG:", message)
	if *graylogEnabled {
		_, file, line, _ := runtime.Caller(1)
		extra := map[string]interface{}{"file": file, "line": line, "deployment": *deploymentName}
		m := wrapBuildGraylogMessage("debug", message, 7, extra)
		sendGraylogMessage(m)
	}
}
