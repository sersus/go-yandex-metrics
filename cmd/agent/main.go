package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

const (
	serverURL      = "http://localhost:8080"
	pollInterval   = 2
	reportInterval = 5 // 5*2 = 10
)

func main() {
	gaugeMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}
	storage := MemStorage{counters: make(map[string]int64), gauges: make(map[string]float64)}
	var m runtime.MemStats

	i := 0
	for {
		collectMetrics(&m, gaugeMetrics, &storage)
		i++
		time.Sleep(pollInterval * time.Second)
		if i%int(reportInterval) == 0 {
			sendMetricsToServer(&storage)
			i = 0
		}
	}
}

func sendMetricsToServer(metrics *MemStorage) {
	for name, value := range metrics.gauges {
		sendGaugeToServer(name, value)
	}
	for name, value := range metrics.counters {
		sendCounterToServer(name, value)
	}
}

func collectMetrics(m *runtime.MemStats, gaugeMetrics []string, storage *MemStorage) {
	runtime.ReadMemStats(m)
	for _, metricName := range gaugeMetrics {
		value := reflect.ValueOf(*m).FieldByName(metricName)
		if value.IsValid() {
			// fmt.Println("Metric " + metricName + " equals " + value.String())
			if value.CanFloat() {
				storage.gauges[metricName] = value.Float()
			} else if value.CanUint() {
				storage.gauges[metricName] = float64(value.Uint())
			}
		} else {
			fmt.Printf("Metric named %s was not found in MemStats", metricName)
		}
	}
	storage.counters["PollCount"]++
	storage.gauges["RandomValue"] = rand.Float64()
}

func sendGaugeToServer(metricName string, metricValue float64) {
	url := fmt.Sprintf("%s/update/gauge/%s/%f", serverURL, metricName, metricValue)
	resp, err := http.Post(url, "text/plain", http.NoBody)
	if err != nil {
		fmt.Printf("Failed to send metric %s to server: %v\n", metricName, err)
	}
	resp.Body.Close()
}

func sendCounterToServer(metricName string, metricValue int64) {
	url := fmt.Sprintf("%s/update/counter/%s/%d", serverURL, metricName, metricValue)
	resp, err := http.Post(url, "text/plain", http.NoBody)
	if err != nil {
		fmt.Printf("Failed to send metric %s to server: %v\n", metricName, err)
	}
	resp.Body.Close()
}
