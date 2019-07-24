package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JeremyLoy/config"
	"github.com/alexflint/go-arg"
	"github.com/jc21/route53-ddns/pkg/logger"
)

// Populated at build time using ldflags
var gitCommit string
var appVersion string

const defaultConfigFile = "~/.aws/route53-ddns.json"

// AppConfig is data used for displaying command usage and specifying configuration options
type AppConfig struct {
	Setup      bool   `arg:"-s" help:"Setup wizard"`
	ConfigFile string `arg:"-c" help:"Config File to use (default: ~/.aws/route53-ddns.json)"`
}

// AWSConfig is the settings that are saved for use in updating
type AWSConfig struct {
	AWSKeyID     string `survey:"aws_key_id"`
	AWSKeySecret string `survey:"aws_key_secret"`
	ZoneID       string `survey:"zone_id"`
	Recordset    string `survey:"recordset"`
}

// Version returns the build version and git commit
func (AppConfig) Version() string {
	return "v" + appVersion + " (" + gitCommit + ")"
}

// Description returns a simple description of the command
func (AppConfig) Description() string {
	return "Update route53 DNS record with your current IP address"
}

// GetConfig returns the AppConfig
func GetConfig() AppConfig {
	var AppArguments AppConfig

	config.FromEnv().To(&AppArguments)
	arg.MustParse(&AppArguments)

	return AppArguments
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
	var answers AWSConfig
	err := survey.Ask(questions, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	logger := logger.Get()
	logger.Trace("Answers: %+v", answers)

	success := writeAwsConfig(answers)
	if !success {
		os.Exit(1)
	}
}

func writeAwsConfig(awsConfig AWSConfig) bool {
	logger := logger.Get()
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
	var filename string
	logger := logger.Get()

	usr, err := user.Current()
	if err != nil {
		logger.Error(err.Error())
	}

	appConfig := GetConfig()
	if appConfig.ConfigFile != "" {
		filename = appConfig.ConfigFile
	} else {
		var strs []string
		strs = append(strs, usr.HomeDir)
		strs = append(strs, "/")

		filename = strings.ReplaceAll(defaultConfigFile, "~/", strings.Join(strs, ""))
	}

	return filename
}

// GetAWSConfig returns the configuration as read from a file
func GetAWSConfig() AWSConfig {
	var awsConfig AWSConfig
	logger := logger.Get()
	filename := getAwsConfigFilename()

	// Make sure file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Error("Configuration not found, run again with -s")
		os.Exit(1)
	}

	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
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
