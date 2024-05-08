# Welcome to your CDK Go project!

This is a blank project for CDK development with Go.

The `cdk.json` file tells the CDK toolkit how to execute your app.

## Useful commands

- `cdk deploy` deploy this stack to your default AWS account/region
- `cdk diff` compare deployed stack with current state
- `cdk synth` emits the synthesized CloudFormation template
- `go test` run unit tests

## Localstack

### Install cdklocal

#### Install globally

`npm install -g aws-cdk-local aws-cdk`

#### Verify it installed correctly

`cdklocal --version`

### Run Docker-compose

`docker-compose up`

### Run Bootstrap

`cdklocal bootstap`

### Deploy

`cdklocal deploy`

### Health Check

`make localstack-health`
