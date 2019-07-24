package main

import (
	"github.com/jc21/route53-ddns/pkg/config"
	"github.com/jc21/route53-ddns/pkg/logger"
	"github.com/jc21/route53-ddns/pkg/updater"
)

func main() {
	argConfig := config.GetConfig()
	logger := logger.Init(argConfig)
	logger.Trace("Args: %+v", argConfig)

	if argConfig.Setup {
		config.SetupAWSConfig()
	}

	awsConfig := config.GetAWSConfig()
	logger.Trace("AWSConfig: %+v", awsConfig)
	updater.Process(argConfig, awsConfig)
}
