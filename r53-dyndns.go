// This entire codebase was written by ChatGPT from OpenAI.
// I have made no modifications myself that were not explicitly
// recommended or suggested by the chatGPT client.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func main() {
	sess := session.Must(session.NewSession())
	r53 := route53.New(sess)

	// Flags
	hostname := flag.String("hostname", "homelab.danmanners.com", "Domain/Record to update.")
	hostedZoneId := flag.String("hosted-zone-id", "HostedZoneGoHere", "AWS Hosted Zone ID to manage.")
	timeCheck := flag.Int("update-interval", 180, "Time between checks/updates.")
	ttl := flag.Int64("ttl", 300, "TTL for the record.")
	ipFetchHost := flag.String("ip-fetch-host", "http://icanhazip.com", "Host to fetch the public IP from.")
	flag.Parse()

	// Confirm that the hostedZoneId and hostname are set
	if *hostedZoneId == "" || *hostname == "" {
		fmt.Println("hostedZoneId and hostname must be set. Exiting.")
		os.Exit(1)
	}

	// Post parsing, set the checkInterval timing correctly
	checkInterval := time.Duration(*timeCheck) * time.Second

	fmt.Printf("Starting up...\n")

	for {
		publicIP, err := getPublicIP(*ipFetchHost)
		if err != nil {
			fmt.Println(err)
			continue
		}

		currentIP, err := getCurrentRecordValue(r53, *hostedZoneId, *hostname)
		if err != nil && err.Error() == fmt.Sprintf("no A record found for hostname %s", *hostname) {
			// If the record does not exist, create it.
			err = createRecord(r53, *hostname, *hostedZoneId, publicIP, *ttl)
			if err != nil {
				fmt.Println(err)
				continue
			}
		} else if err != nil {
			fmt.Println(err)
			break

		}

		if publicIP != currentIP {
			err = updateRecord(r53, *hostname, *hostedZoneId, publicIP, *ttl)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		time.Sleep(checkInterval)
	}
}

func getPublicIP(ipFetchHost string) (string, error) {
	// Get the current time and format it as a string.
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05")

	// Log the URL that is being requested with the timestamp.
	fmt.Printf("[%s] Fetching public IP from %s...\n", timestamp, ipFetchHost)

	// Create a new HTTP client and make a GET request to the third-party service.
	client := http.Client{}
	resp, err := client.Get(ipFetchHost)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body and log the public IP address with the timestamp.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	publicIP := string(body)
	fmt.Printf("[%s] Received public IP: %s\n", timestamp, publicIP)
	return publicIP, nil
}

func getCurrentRecordValue(r53 *route53.Route53, hostedZoneId string, hostname string) (string, error) {
	// Lookup the current value of the A record for the specified hostname.
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(hostedZoneId),
		StartRecordName: aws.String(hostname),
		StartRecordType: aws.String("A"),
	}

	result, err := r53.ListResourceRecordSets(input)
	if err != nil {
		return "", err
	}

	for _, record := range result.ResourceRecordSets {
		if *record.Name == hostname && *record.Type == "A" {
			return *record.ResourceRecords[0].Value, nil
		}
	}

	// If the A record does not exist, return an empty string instead of an error.
	return "", nil
}

func updateRecord(r53 *route53.Route53, hostname, hostedZoneId string, publicIp string, ttl int64) error {
	// Update the A record for the specified hostname with the new value and TTL.
	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneId),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(hostname),
						Type: aws.String("A"),
						TTL:  aws.Int64(ttl),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(publicIp),
							},
						},
					},
				},
			},
		},
	}

	_, err := r53.ChangeResourceRecordSets(input)

	if err != nil {
		return err
	}

	// Log a message indicating that the update was successful.
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] Successfully updated A record for %s to %s\n", timestamp, hostname, publicIp)
	return nil

}

func createRecord(r53 *route53.Route53, hostname string, hostedZoneId string, value string, ttl int64) error {
	// Check if the A record already exists.
	_, err := getCurrentRecordValue(r53, hostedZoneId, hostname)
	if err == nil {
		// If the record already exists, return an error.
		return fmt.Errorf("'A record' for hostname %s already exists", hostname)
	}

	// Create an A record for the specified hostname with the given value and TTL.
	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneId),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("CREATE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(hostname),
						Type: aws.String("A"),
						TTL:  aws.Int64(ttl),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(value),
							},
						},
					},
				},
			},
		},
	}

	_, err = r53.ChangeResourceRecordSets(input)
	return err
}
