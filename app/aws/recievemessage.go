package aws
import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/borlinp/amazon-sns-sqs/common"
)
func NewQueue (queue string) (string){
	sess :=BuildSession()
	svc := sqs.New(sess)

    result, err := svc.CreateQueue(&sqs.CreateQueueInput{
        QueueName: queue,
        Attributes: map[string]*string{
            "DelaySeconds":           aws.String("60"),
            "MessageRetentionPeriod": aws.String("86400"),
        },
    })
	return *result.QueueUrl
}
func subscribe(queueUrl string, cancel <-chan os.Signal) {
	awsSession := common.BuildSession()
	svc := sqs.New(awsSession, nil)

	for {
		messages := receiveMessages(svc, queueUrl)

		for _, msg := range messages {
			if msg == nil {
				continue
			}
			fmt.Println(*msg.Body)
			go DeleteMessage(svc, queueUrl, msg.ReceiptHandle)
		}

		select {
		case <-cancel:
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func receiveMessages(svc *sqs.SQS, queueUrl string) []*sqs.Message {

	receiveMessagesInput := &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(10), // max 10
		WaitTimeSeconds:     aws.Int64(3),  // max 20
		VisibilityTimeout:   aws.Int64(20), // max 20
	}

	receiveMessageOutput, err :=
		svc.ReceiveMessage(receiveMessagesInput)

	if err != nil {
		fmt.Println("Error: ", err)
		return nil
	}

	if receiveMessageOutput == nil || len(receiveMessageOutput.Messages) == 0 {
		return nil
	}

	return receiveMessageOutput.Messages
}

func DeleteMessage(svc *sqs.SQS, queueUrl string, handle *string) {
	delInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: handle,
	}
	_, err := svc.DeleteMessage(delInput)

	if err != nil {
		fmt.Println("Delete Error", err)
		return
	}
}

func SubscribeSNS(session *session.Session, topic string) {
	svc := sns.New(session)

	_, err := svc.Subscribe(&sns.SubscribeInput{
		// Attributes:            nil,
		Endpoint: aws.String("myname@mydomain.com"),
		Protocol: aws.String("email"),
		// ReturnSubscriptionArn: nil,
		TopicArn: aws.String(topic),
	})
	if err != nil {
		fmt.Println(err)
	}
}
