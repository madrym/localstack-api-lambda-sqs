package stacks

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewServerlessStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// Define an SQS queue
	queue := awssqs.NewQueue(stack, jsii.String("MyQueue"), nil)

	// Create a new PolicyDocument for Sender Lambda
	senderPolicyDocument := awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{})

	senderLambdaPolicyStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("lambda:InvokeFunction", "sqs:SendMessage", "sqs:GetQueueUrl"),
		Resources: jsii.Strings("*"),
		Effect:    awsiam.Effect_ALLOW,
	})

	senderPolicyDocument.AddStatements(senderLambdaPolicyStatement)

	senderLambdaRole := awsiam.NewRole(stack, jsii.String("SenderLambdaRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaSQSQueueExecutionRole")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")),
		},
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"SecretPolicy": senderPolicyDocument,
		},
	})

	// Get CWD
	cwd, _ := os.Getwd()
	senderLambdaPath := cwd + "/functions/sender"
	receiverLambdaPath := cwd + "/functions/receiver"

	// Define the Lambda function
	senderLambda := awslambda.NewFunction(stack, jsii.String("SenderLambda"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2(),
		Handler: jsii.String("sender.handler"),
		Code:    awslambda.Code_FromAsset(jsii.String(senderLambdaPath), nil),
		Role:    senderLambdaRole,
		Environment: &map[string]*string{
			"QUEUE_URL": queue.QueueUrl(),
		},
	})

	// Define a secret in Secrets Manager
	secret := awssecretsmanager.NewSecret(stack, jsii.String("MySecret"), &awssecretsmanager.SecretProps{
		SecretName: jsii.String("MySecret"),
	})

	// Create a new PolicyDocument
	recieverPolicyDocument := awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{})
	// Create a new PolicyStatement
	getSecretPolicyStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("secretsmanager:GetSecretValue"),
		Resources: jsii.Strings(*secret.SecretArn()),
		Effect:    awsiam.Effect_ALLOW,
	})
	allowSQSAndSecretsPolicyStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("sqs:ReceiveMessage", "sqs:DeleteMessage", "sqs:GetQueueAttributes"),
		Resources: jsii.Strings(*queue.QueueArn()),
		Effect:    awsiam.Effect_ALLOW,
	})

	// Add the PolicyStatement to the PolicyDocument
	recieverPolicyDocument.AddStatements(getSecretPolicyStatement)
	recieverPolicyDocument.AddStatements(allowSQSAndSecretsPolicyStatement)

	// Create the IAM role for the receiver Lambda
	receiverLambdaRole := awsiam.NewRole(stack, jsii.String("ReceiverLambdaRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"SecretPolicy": recieverPolicyDocument,
		},
	})

	// Define the Lambda function that reads from SQS and logs messages
	receiverLambda := awslambda.NewFunction(stack, jsii.String("ReceiverLambda"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2(),
		Handler: jsii.String("receiver.handler"),
		Code:    awslambda.Code_FromAsset(jsii.String(receiverLambdaPath), nil),
		Role:    receiverLambdaRole,
		Environment: &map[string]*string{
			"SECRET_ID": secret.SecretArn(),
			"QUEUE_URL": queue.QueueUrl(),
		},
	})

	// Create an event source mapping between the SQS queue and the receiver Lambda
	awslambda.NewEventSourceMapping(stack, jsii.String("MyQueueEventSource"), &awslambda.EventSourceMappingProps{
		Target:         receiverLambda,
		EventSourceArn: queue.QueueArn(),
		BatchSize:      jsii.Number(10), // Number of records to process per batch
		Enabled:        jsii.Bool(true),
	})

	// API Gateway setup
	api := awsapigateway.NewRestApi(stack, jsii.String("WebhookAPI"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("WebhookService"),
		Description: jsii.String("API Gateway for handling webhooks."),
	})

	// Define a new resource for the webhook endpoint
	webhookResource := api.Root().AddResource(jsii.String("webhook"), &awsapigateway.ResourceOptions{})

	// Add a POST method to the /webhook resource that triggers the senderLambda
	webhookResource.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(senderLambda, &awsapigateway.LambdaIntegrationOptions{
		AllowTestInvoke: jsii.Bool(true),
	}), &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE, // Assuming no auth required; adjust as necessary
	})

	// Allow the sender Lambda to send messages to the SQS queue
	queue.GrantSendMessages(senderLambda)

	return stack
}
