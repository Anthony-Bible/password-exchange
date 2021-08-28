
package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)
//GetQueueURL converts an ARN to a queue url
func GetQueueURL(sess *session.Session, queue *string) (*sqs.GetQueueUrlOutput, error) {
    // Create an SQS service client
    svc := sqs.New(sess)

    result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
        QueueName: queue,
    })
    if err != nil {
        return nil, err
    }

    return result, nil
}
//SendSQS uses the aws sdk to send a message to SQS
func SendSQS(session *session.Session, destination string, message string) {
	svc := sqs.New(session, nil)

	sendInput := &sqs.SendMessageInput{
		MessageBody: aws.String(message),
		QueueUrl:    aws.String(destination),
	}

	_, err := svc.SendMessage(sendInput)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println(output.MessageId)
}

