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

type ServerOptions struct {
	Address         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
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

func ParceServerFlags(options *ServerOptions) {
	flag.Parse()
	envAddr, exists := os.LookupEnv("ADDRESS")
	if exists && envAddr != "" {
		options.Address = envAddr
	}
	envStoreInt, exists := os.LookupEnv("STORE_INTERVAL")
	if exists && envStoreInt != "" {
		options.StoreInterval, _ = strconv.Atoi(envStoreInt)
	}
	envFilePath, exists := os.LookupEnv("FILE_STORAGE_PATH")
	if exists && envFilePath != "" {
		options.FileStoragePath = envFilePath
	}
	envRestore, exists := os.LookupEnv("RESTORE")
	if exists && envRestore != "" {
		options.Restore, _ = strconv.ParseBool(envRestore)
	}
}
