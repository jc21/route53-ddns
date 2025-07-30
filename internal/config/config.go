package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"route53-ddns/internal/helper"
	"route53-ddns/internal/logger"
	"route53-ddns/internal/model"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JeremyLoy/config"
	"github.com/alexflint/go-arg"
)

// Populated at build time using ldflags
var appArguments model.ArgConfig

const defaultConfigFile = "~/.aws/route53-ddns.json"

// GetConfig returns the ArgConfig
func GetConfig() model.ArgConfig {
	// nolint: gosec, errcheck
	config.FromEnv().To(&appArguments)
	arg.MustParse(&appArguments)

	return appArguments
}

// SetupAWSConfig will ask for setup questions
func SetupAWSConfig() {
	// the questions to ask
	var questions = []*survey.Question{
		{
			Name:     "aws_key_id",
			Prompt:   &survey.Input{Message: "AWS Access Key:"},
			Validate: survey.Required,
		},
		{
			Name:     "aws_key_secret",
			Prompt:   &survey.Input{Message: "AWS Access Key Secret:"},
			Validate: survey.Required,
		},
		{
			Name:     "zone_id",
			Prompt:   &survey.Input{Message: "Route53 Zone ID:"},
			Validate: survey.Required,
		},
		{
			Name:     "recordset",
			Prompt:   &survey.Input{Message: "Route53 Record Set:"},
			Validate: survey.Required,
		},
		{
			Name: "protocols",
			Prompt: &survey.Select{
				Message: "Which IP Protocals do you update?",
				Options: []string{"IPv4 Only", "IPv6 Only ", "Both"},
			},
			Validate: survey.Required,
		},
		{
			Name:   "pushover_user_token",
			Prompt: &survey.Input{Message: "Pushover User Token: (leave blank to disable)"},
		},
	}

	// perform the questions
	var answers model.AWSConfig
	err := survey.Ask(questions, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	logger.Trace("Answers: %+v", answers)

	writeErr := answers.Write(getAwsConfigFilename())
	if writeErr != nil {
		logger.Error("Could not write configuration: %v", writeErr.Error())
		os.Exit(1)
	}
}

func getAwsConfigFilename() string {
	argConfig := GetConfig()
	if argConfig.ConfigFile != "" {
		return argConfig.ConfigFile
	}

	return helper.GetFullFilename(defaultConfigFile)
}

// GetAWSConfig returns the configuration as read from a file
func GetAWSConfig() model.AWSConfig {
	var awsConfig model.AWSConfig
	filename := getAwsConfigFilename()

	// Make sure file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Error("Configuration not found, run again with -s")
		os.Exit(1)
	}

	// nolint: gosec
	jsonFile, err := os.Open(filename)
	if err != nil {
		logger.Error("Configuration could not be opened: %v", err.Error())
		os.Exit(1)
	}

	// nolint: gosec, errcheck
	defer jsonFile.Close()

	contents, readErr := io.ReadAll(jsonFile)
	if readErr != nil {
		logger.Error("Configuration file could not be read: %v", readErr.Error())
		// nolint: gocritic
		os.Exit(1)
	}

	unmarshalErr := json.Unmarshal(contents, &awsConfig)
	if unmarshalErr != nil {
		logger.Error("Configuration file looks damaged, run again with -s")
		os.Exit(1)
	}

	return awsConfig
}
