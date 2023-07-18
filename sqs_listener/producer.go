package sqs_listener

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type Publisher struct {
	Conf boilerplate.SQSConfiguration
	Svc  *sqs.SQS
}

func SendMessage[T any](c *Publisher, data T) (*sqs.SendMessageOutput, error) {
	message, err := marshalMessage(data)
	if err != nil {
		return nil, err
	}

	result, err := c.Svc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(c.Conf.Url),
		MessageBody: aws.String(message),
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("Url : ", c.Conf.Url, "\nMessage Received  : ", message)

	return result, nil
}

func marshalMessage[T any](data T) (string, error) {
	s, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(s), nil
}
