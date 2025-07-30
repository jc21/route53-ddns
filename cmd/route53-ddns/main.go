package main

import (
	"route53-ddns/internal/config"
	"route53-ddns/internal/logger"
	"route53-ddns/internal/updater"
)

func main() {
	argConfig := config.GetConfig()
	log := logger.Init(argConfig)
	log.Trace("Args: %+v", argConfig)

	if argConfig.Setup {
		config.SetupAWSConfig()
	}

	awsConfig := config.GetAWSConfig()
	log.Trace("AWSConfig: %+v", awsConfig)
	updater.Process(argConfig, awsConfig)
}
