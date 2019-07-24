package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JeremyLoy/config"
	"github.com/alexflint/go-arg"
	"github.com/jc21/route53-ddns/pkg/helper"
	"github.com/jc21/route53-ddns/pkg/logger"
	"github.com/jc21/route53-ddns/pkg/model"
)

// Populated at build time using ldflags
var appArguments model.ArgConfig

const defaultConfigFile = "~/.aws/route53-ddns.json"

// GetConfig returns the ArgConfig
func GetConfig() model.ArgConfig {
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
		logger.Error("Could not write configuration: %v", err.Error())
		os.Exit(1)
	}
}

func writeAwsConfig(awsConfig model.AWSConfig) bool {
	filename := getAwsConfigFilename()
	content, _ := json.MarshalIndent(awsConfig, "", " ")

	// Make sure the ".aws" folder exists
	folder := path.Dir(filename)
	dirErr := os.MkdirAll(folder, os.ModePerm)
	if dirErr != nil {
		logger.Error("Could not create folder: %v", dirErr.Error())
		return false
	}

	logger.Trace("Writing config to: %+v", filename)

	err := ioutil.WriteFile(filename, content, 0600)
	if err != nil {
		logger.Error("Could not save config: %v", err.Error())
		return false
	}

	logger.Info("Wrote config to %v", filename)
	return true
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

	jsonFile, err := os.Open(filename)
	if err != nil {
		logger.Error("Configuration could not be opened: %v", err.Error())
		os.Exit(1)
	}

	defer jsonFile.Close()

	contents, readErr := ioutil.ReadAll(jsonFile)
	if readErr != nil {
		logger.Error("Configuration file could not be read: %v", readErr.Error())
		os.Exit(1)
	}

	unmarshalErr := json.Unmarshal(contents, &awsConfig)
	if unmarshalErr != nil {
		logger.Error("Configuration file looks damaged, run again with -s")
		os.Exit(1)
	}

	return awsConfig
}
