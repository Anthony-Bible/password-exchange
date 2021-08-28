
package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)
//GetQueueURL converts an ARN to a queue url
// func GetQueueURL(sess *session.Session, queue *string) (*sqs.GetQueueUrlOutput, error) {
//     // Create an SQS service client
//     svc := sqs.New(sess)

//     result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
//         QueueName: queue,
//     })
//     if err != nil {
//         return nil, err
//     }

//     return result, nil
// }
//SendSQS uses the aws sdk to send a message to SQS
func SendSNS(session *session.Session, destination string, message string) {
	svc := sns.New(session)

	pubInput := &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(destination),
		MessageGroupId: aws.String("encryption"),
	}

	_, err := svc.Publish(pubInput)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//fmt.Println(output.MessageId)
}
