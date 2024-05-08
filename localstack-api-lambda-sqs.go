package main

import (
	stacks "localstack-api-lambda-sqs/lib"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	var account string = "000000000000"  // Dummy account ID for LocalStack
	var region string = "ap-southeast-2" // Default LocalStack region

	// Use environment variables if they are set
	if envAccount, isSet := os.LookupEnv("CDK_DEFAULT_ACCOUNT"); isSet {
		account = envAccount
	}
	if envRegion, isSet := os.LookupEnv("CDK_DEFAULT_REGION"); isSet {
		region = envRegion
	}

	stacks.NewServerlessStack(app, "localstack-api-lambda-sqs", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: aws.String(account),
			Region:  aws.String(region),
		},
	})

	app.Synth(nil)
}
