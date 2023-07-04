package sqs_listener

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type SQSListener struct {
	Conf          boilerplate.SQSConfiguration
	Svc           *sqs.SQS
	Callback      func(m string) error
	stopCh        chan struct{}
	ExecutionData ExecutionData
	ServiceName   string
}

func (i *SQSListener) StartListener() {
	go func(url string) {
		defer i.StopListener()
		for {
			select {
			case <-i.stopCh:
				return
			default:
				// Call the ReceiveMessage API
				result, err := i.Svc.ReceiveMessage(&sqs.ReceiveMessageInput{
					QueueUrl:            aws.String(i.Conf.Url),
					MaxNumberOfMessages: aws.Int64(i.Conf.MaxMessages), // Receive up to 10 messages
				})
				if err != nil {
					log.Println("Expected : ", url)
					log.Printf("\nError receiving message : %s ; \nRetrying in 5 seconds; \nRegion : %s \nQueue URL : %s", err.Error(), i.Conf.Region, i.Conf.Url)
					time.Sleep(5 * time.Second) // Pause for 5 seconds before retrying
					continue
				}

				// Process each received message
				for _, msg := range result.Messages {
					i.Callback(*msg.Body)

					if err == nil {
						// Delete the message from the queue
						_, err := i.Svc.DeleteMessage(&sqs.DeleteMessageInput{
							QueueUrl:      aws.String(i.Conf.Url),
							ReceiptHandle: msg.ReceiptHandle,
						})
						if err != nil {
							log.Println("Error deleting message:", err)
						}
					}
				}

				// Pause for a short duration before polling again
				time.Sleep(1 * time.Second)
			}
		}
	}(i.Conf.Url)
}

func (i *SQSListener) StopListener() {
	close(i.stopCh)
}
