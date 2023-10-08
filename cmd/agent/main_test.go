package main

import (
	"runtime"
	"testing"
)

func Test_collectMetrics(t *testing.T) {
	type args struct {
		m            *runtime.MemStats
		gaugeMetrics []string
		storage      *MemStorage
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normalTest",
			args: args{
				m: new(runtime.MemStats),
				gaugeMetrics: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
					"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
					"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
					"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"},
				storage: &MemStorage{counters: make(map[string]int64), gauges: make(map[string]float64)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectMetrics(tt.args.m, tt.args.gaugeMetrics, tt.args.storage)
		})
	}
}
