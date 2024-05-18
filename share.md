`template.yaml`
```
AWSTemplateFormatVersion: '2010-09-09'
Transform: 'AWS::Serverless-2016-10-31'
Resources:
  WebhookFunction:
    Type: 'AWS::Serverless::Function'
    Properties:
      Handler: webhook
      Runtime: provided.al2
      CodeUri: functions/webhook/
      Events:
        Api:
          Type: Api
          Properties:
            Path: /webhooks
            Method: POST
  ProcessQueueFunction:
    Type: 'AWS::Serverless::Function'
    Properties:
      Handler: processQueue
      Runtime: provided.al2
      CodeUri: functions/processQueue/
      Events:
        SQSEvent:
          Type: SQS
          Properties:
            Queue: !GetAtt WebhookQueue.Arn
  WebhookQueue:
    Type: 'AWS::SQS::Queue'
  ProcessQueuePolicy:
    Type: 'AWS::IAM::Policy'
    Properties:
      PolicyName: 'processQueuePolicy'
      PolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: 'Allow'
            Action:
              - 'sqs:ReceiveMessage'
              - 'sqs:DeleteMessage'
              - 'sqs:GetQueueAttributes'
            Resource: !GetAtt WebhookQueue.Arn
      Roles:
        - !GetAtt ProcessQueueFunctionRole.Arn
  WebhookFunctionRole:
    Type: 'AWS::IAM::Role'
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: 'Allow'
            Principal:
              Service: 'lambda.amazonaws.com'
            Action: 'sts:AssumeRole'
      Policies:
        - PolicyName: 'WebhookPolicy'
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: 'Allow'
                Action:
                  - 'sqs:SendMessage'
                Resource: !GetAtt WebhookQueue.Arn
  ProcessQueueFunctionRole:
    Type: 'AWS::IAM::Role'
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: 'Allow'
            Principal:
              Service: 'lambda.amazonaws.com'
            Action: 'sts:AssumeRole'
      Policies:
        - PolicyName: 'ProcessQueuePolicy'
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: 'Allow'
                Action:
                  - 's3:PutObject'
                Resource: 'arn:aws:s3:::*'

```
_____________________________________________

`functions/processQueue/main.go`
```
package main

import (
    "fmt"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

func handler(event events.SQSEvent) error {
    sess := session.Must(session.NewSession(&aws.Config{
        Region:   aws.String("us-east-1"),
        Endpoint: aws.String("http://localstack:4566"),
    }))
    svc := s3.New(sess)

    bucket := "localstack-bucket"
    key := "message.txt"

    for _, record := range event.Records {
        _, err := svc.PutObject(&s3.PutObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(key),
            Body:   aws.ReadSeekCloser(strings.NewReader(record.Body)),
        })
        if err != nil {
            return fmt.Errorf("failed to upload data to %s/%s, %s", bucket, key, err)
        }
    }

    return nil
}

func main() {
    lambda.Start(handler)
}

```
_____________________________________

`functions/webhook/main.go`

```
package main

import (
    "encoding/json"
    "fmt"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
)

type RequestBody struct {
    Body string `json:"body"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    var reqBody RequestBody
    err := json.Unmarshal([]byte(request.Body), &reqBody)
    if err != nil {
        return events.APIGatewayProxyResponse{StatusCode: 400}, err
    }

    sess := session.Must(session.NewSession(&aws.Config{
        Region:   aws.String("us-east-1"),
        Endpoint: aws.String("http://localstack:4566"),
    }))
    svc := sqs.New(sess)

    queueUrl := "http://localstack:4566/000000000000/WebhookQueue"

    _, err = svc.SendMessage(&sqs.SendMessageInput{
        MessageBody: aws.String(reqBody.Body),
        QueueUrl:    aws.String(queueUrl),
    })
    if err != nil {
        return events.APIGatewayProxyResponse{StatusCode: 500}, err
    }

    return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
    lambda.Start(handler)
}

```
_______________________________________________
`functions/*/Makefile`

```
build:
    GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
    zip function.zip bootstrap
```
________________________________________________

`samconfig.toml`
```
version = 0.1
[default]
[default.deploy]
[default.deploy.parameters]
stack_name = "localstack-golang-sam"
s3_bucket = "localstack-sam-bucket"
region = "us-east-1"
capabilities = "CAPABILITY_IAM"
endpoint-url = "http://localhost:4566"
```

______________________________________________
`scripts/create_bucket.sh`
```
#!/bin/bash

BUCKET_NAME=$1

awslocal s3 mb s3://$BUCKET_NAME
```
____________________________________________

`Makefile`
```
LOCALSTACK_BUCKET_NAME=localstack-sam-bucket

.PHONY: all build deploy create-bucket

all: create-bucket build deploy

create-bucket:
    @sh ./scripts/create_bucket.sh $(LOCALSTACK_BUCKET_NAME)

build:
    sam build

deploy:
    sam deploy --config-file samconfig.toml --stack-name localstack-golang-sam

```
