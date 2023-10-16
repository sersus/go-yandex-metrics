package config

import (
	"flag"
	"os"
	"strconv"
)

type Options struct {
	Address        string
	ReportInterval int
	PollInterval   int
}

func ParceFlags(options *Options) {
	flag.Parse()
	envAddr, exists := os.LookupEnv("ADDRESS")
	if exists && envAddr != "" {
		options.Address = envAddr
	}
	envPollInt, exists := os.LookupEnv("POLL_INTERVAL")
	if exists && envPollInt != "" {
		options.PollInterval, _ = strconv.Atoi(envPollInt)
	}
	envRepInt, exists := os.LookupEnv("REPORT_INTERVAL")
	if exists && envRepInt != "" {
		options.ReportInterval, _ = strconv.Atoi(envRepInt)
	}
}
