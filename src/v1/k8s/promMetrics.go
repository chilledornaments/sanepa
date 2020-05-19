package main

import (
	"time"
)

func incScaleUpCounter() {
	go func() {
		scaleUpPromCounter.Inc()
		time.Sleep(1 * time.Second)
	}()
}

func incScaleUpErrCounter() {
	go func() {
		scaleUpErrPromCounter.Inc()
		time.Sleep(1 * time.Second)
	}()
}

func incScaleDownCounter() {
	go func() {
		scaleDownPromCounter.Inc()
		time.Sleep(1 * time.Second)
	}()
}

func incScaleDownErrCounter() {
	go func() {
		scaleDownErrPromCounter.Inc()
		time.Sleep(1 * time.Second)
	}()
}

func incCollectionErrCounter() {
	go func() {
		collectionErrPromCounter.Inc()
		time.Sleep(1 * time.Second)
	}()
}
