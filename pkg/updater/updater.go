package updater

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	externalip "github.com/glendc/go-external-ip"
	ddnsaws "github.com/jc21/route53-ddns/pkg/aws"
	"github.com/jc21/route53-ddns/pkg/config"
	"github.com/jc21/route53-ddns/pkg/logger"
	"github.com/jc21/route53-ddns/pkg/model"
)

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
	state := config.GetRoute53State()
	logger.Trace("STATE: %+v", state)

	if argConfig.Force || awsConfig.ZoneID != state.ZoneID || awsConfig.Recordset != state.Recordset {
		// Update the Route53 IP and save to new state
		logger.Info("Updating IP to %v", ip.String())

		updateErr := updateIP(awsConfig, ip.String())
		if updateErr != nil {
			logger.Error("Could not update Route53: %v", updateErr.Error())
		}

	} else {
		logger.Info("IP %v hasn't changed, not updating Route53", ip.String())
	}
}

func updateIP(awsConfig model.AWSConfig, ip string) error {
	creds := credentials.NewCredentials(&ddns.MyProvider{})
	credValue, err := creds.Get()

	sess := session.Must(session.NewSession())
	svc := route53.New(sess)

	// Create a message
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("CREATE"),
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
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				fmt.Println(route53.ErrCodeNoSuchHostedZone, aerr.Error())
			case route53.ErrCodeNoSuchHealthCheck:
				fmt.Println(route53.ErrCodeNoSuchHealthCheck, aerr.Error())
			case route53.ErrCodeInvalidChangeBatch:
				fmt.Println(route53.ErrCodeInvalidChangeBatch, aerr.Error())
			case route53.ErrCodeInvalidInput:
				fmt.Println(route53.ErrCodeInvalidInput, aerr.Error())
			case route53.ErrCodePriorRequestNotComplete:
				fmt.Println(route53.ErrCodePriorRequestNotComplete, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			return err
		}
	}

	logger.Trace("AWS Result: %v", result)

	return nil
}
