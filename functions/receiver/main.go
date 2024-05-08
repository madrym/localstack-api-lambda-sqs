package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func handler(ctx context.Context) {
	// Initialize AWS session
	var sess *session.Session

	if os.Getenv("AWS_ENV") == "local" {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region:           aws.String("ap-southeast-2"),
				Endpoint:         aws.String("http://localhost:4566"),
				S3ForcePathStyle: aws.Bool(true),
				Credentials:      credentials.NewStaticCredentials("test", "test", ""),
			},
		}))
	} else {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region: aws.String("ap-southeast-2"),
			},
		}))
	}

	// Create SQS client
	sqsClient := sqs.New(sess)

	// Receive message from SQS
	result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(os.Getenv("QUEUE_URL")),
		MaxNumberOfMessages: aws.Int64(1),
	})
	if err != nil {
		fmt.Printf("Unable to receive message from queue: %v\n", err)
		return
	}

	for _, message := range result.Messages {
		fmt.Printf("Received message: %s\n", *message.Body)

		// Process the secret
		secretManager := secretsmanager.New(sess)
		secretValue, err := secretManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(os.Getenv("SECRET_ID")),
		})
		if err != nil {
			fmt.Printf("Failed to retrieve secret: %s\n", err)
		} else {
			fmt.Printf("Secret value: %s\n", *secretValue.SecretString)
		}

		// Delete the message from the queue
		_, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(os.Getenv("QUEUE_URL")),
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			fmt.Printf("Failed to delete message: %s\n", err)
		}
	}
}

func main() {
	lambda.Start(handler)
}
