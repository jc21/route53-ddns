package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	externalip "github.com/glendc/go-external-ip"
	"github.com/gregdel/pushover"
	"github.com/jc21/route53-ddns/pkg/helper"
	"github.com/jc21/route53-ddns/pkg/logger"
	"github.com/jc21/route53-ddns/pkg/model"
)

const defaultStateFile = "~/.aws/route53-ddns-state.json"

// Process will update the ip address with route53, if forced or changed
func Process(argConfig model.ArgConfig, awsConfig model.AWSConfig) {
	// Determine the current public ip
	consensus := externalip.DefaultConsensus(nil, nil)
	ip, err := consensus.ExternalIP()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Determine if we need to update it
	state := GetRoute53State(argConfig)
	logger.Trace("STATE: %+v", state)

	if argConfig.Force || awsConfig.ZoneID != state.ZoneID || awsConfig.Recordset != state.Recordset {
		// Update the Route53 IP and save to new state
		logger.Info("Updating IP to %v", ip.String())

		updateErr := updateIP(awsConfig, ip.String())
		if updateErr != nil {
			logger.Error("Could not update Route53: %v", updateErr.Error())
		} else {
			logger.Info("IP has been updated to %v for %v", ip.String(), awsConfig.Recordset)

			// Save state
			state.ZoneID = awsConfig.ZoneID
			state.Recordset = awsConfig.Recordset
			state.LastIP = ip.String()
			state.LastUpdateTime = time.Now()
			state.Write(getRoute53StateFilename(argConfig))

			if awsConfig.PushoverUserToken != "" {
				pushoverApp := pushover.New("a4dhut1a7waegz6p2xh7enzegjedgo")
				recipient := pushover.NewRecipient(awsConfig.PushoverUserToken)

				message := &pushover.Message{
					Message:    fmt.Sprintf("For %v", awsConfig.Recordset),
					Title:      fmt.Sprintf("IP updated to %v", ip.String()),
					Priority:   0,
					URL:        "",
					URLTitle:   "",
					Timestamp:  time.Now().Unix(),
					Retry:      60 * time.Second,
					Expire:     time.Hour,
					DeviceName: "",
					Sound:      "",
				}

				// Send the message to the recipient
				_, err := pushoverApp.SendMessage(message, recipient)
				if err != nil {
					logger.Error(err.Error())
					os.Exit(1)
				} else {
					logger.Info("Pushover Notification Sent OK")
				}
			}
		}

	} else {
		logger.Info("IP %v hasn't changed, not updating Route53", ip.String())
	}
}

func updateIP(awsConfig model.AWSConfig, ip string) error {
	session, sessionErr := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(awsConfig.AWSKeyID, awsConfig.AWSKeySecret, ""),
	})

	if sessionErr != nil {
		return sessionErr
	}

	svc := route53.New(session)

	// Create a message
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(awsConfig.Recordset),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip),
							},
						},
						TTL:  aws.Int64(60),
						Type: aws.String("A"),
					},
				},
			},
			Comment: aws.String("Updated by route53-ddns"),
		},
		HostedZoneId: aws.String(awsConfig.ZoneID),
	}

	// Send
	result, err := svc.ChangeResourceRecordSets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return aerr
		} else {
			return err
		}
	}

	logger.Trace("AWS Result: %v", result)

	return nil
}

func getRoute53StateFilename(argConfig model.ArgConfig) string {
	if argConfig.StateFile != "" {
		return argConfig.StateFile
	}

	return helper.GetFullFilename(defaultStateFile)
}

// GetRoute53State returns the configuration as read from a file
func GetRoute53State(argConfig model.ArgConfig) model.Route53State {
	var state model.Route53State
	filename := getRoute53StateFilename(argConfig)

	jsonFile, err := os.Open(filename)
	if err == nil {
		defer jsonFile.Close()
		contents, readErr := ioutil.ReadAll(jsonFile)
		if readErr == nil {
			json.Unmarshal(contents, &state)
		}
	}

	return state
}
