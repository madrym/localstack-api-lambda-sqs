package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// RequestBody structure to parse incoming requests
type RequestBody struct {
	Message string `json:"message"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var reqBody RequestBody
	err := json.Unmarshal([]byte(request.Body), &reqBody)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("Error parsing request: %s", err),
		}, nil
	}

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

	// Initialize SQS client
	sqsClient := sqs.New(sess)

	// Send message to SQS
	_, err = sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(os.Getenv("QUEUE_URL")),
		MessageBody: aws.String(reqBody.Message),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to send message: %s", err),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Message sent successfully",
	}, nil
}

func main() {
	lambda.Start(handler)
}
