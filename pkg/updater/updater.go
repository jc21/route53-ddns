package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
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
	state := GetRoute53State(argConfig)
	logger.Trace("STATE: %+v", state)

	// Apply state
	state.ZoneID = awsConfig.ZoneID
	state.Recordset = awsConfig.Recordset
	state.LastUpdateTime = time.Now()

	if awsConfig.ZoneID != state.ZoneID || awsConfig.Recordset != state.Recordset {
		argConfig.Force = true
	}

	awsConfig.Protocols = strings.Trim(awsConfig.Protocols, " ")
	changed := false
	hasError := false

	if awsConfig.Protocols == "IPv4 Only" || awsConfig.Protocols == "Both" || awsConfig.Protocols == "" {
		// Determine IPv4
		consensus.UseIPProtocol(4)
		ipv4, errv4 := consensus.ExternalIP()
		if errv4 == nil {
			changed, errv4 = updateIPProtocol(ipv4, state.LastIPv4, argConfig, awsConfig)
		}

		if errv4 != nil {
			logger.Error(errv4.Error())
			hasError = true
		} else if changed {
			state.LastIPv4 = ipv4.String()
			state.Write(getRoute53StateFilename(argConfig))
		}
	}

	if awsConfig.Protocols == "IPv6 Only" || awsConfig.Protocols == "Both" {
		// Determine IPv6
		consensus.UseIPProtocol(6)
		ipv6, errv6 := consensus.ExternalIP()
		if errv6 == nil {
			changed, errv6 = updateIPProtocol(ipv6, state.LastIPv6, argConfig, awsConfig)
		}

		if errv6 != nil {
			logger.Error(errv6.Error())
			hasError = true
		} else if changed {
			state.LastIPv6 = ipv6.String()
			state.Write(getRoute53StateFilename(argConfig))
		}
	}

	if hasError {
		os.Exit(1)
	}
}

// updateIPProtocol returns: changed, error
func updateIPProtocol(ip net.IP, lastIP string, argConfig model.ArgConfig, awsConfig model.AWSConfig) (bool, error) {
	if ip.String() != lastIP || argConfig.Force {
		// Update the Route53 IP and save to new state
		logger.Info("Updating IP to %v", ip.String())

		// Determine if this is ipv4 or ipv6
		recordType := "A"
		if strings.Contains(ip.String(), ":") {
			recordType = "AAAA"
		}

		updateErr := updateIP(awsConfig, ip.String(), recordType)
		if updateErr != nil {
			logger.Error("Could not update Route53: %v", updateErr.Error())
			return false, updateErr
		} else {
			logger.Info("'%s' record has been updated to %v for %v", recordType, ip.String(), awsConfig.Recordset)

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
				} else {
					logger.Info("Pushover Notification Sent OK")
				}
			}

			return true, nil
		}
	} else {
		logger.Info("IP %v hasn't changed, not updating Route53", ip.String())
	}

	return false, nil
}

func updateIP(awsConfig model.AWSConfig, ip, recordType string) error {
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
						Type: aws.String(recordType),
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
