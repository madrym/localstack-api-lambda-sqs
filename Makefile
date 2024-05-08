.PHONY: deploy-local deploy-cloud

deploy-local:
	cdklocal deploy

deploy-cloud:
	cdk deploy

localstack-start:
	docker-compose up

localstack-stop:
	docker-compose down

localstack-health:
	curl http://localhost:4566/_localstack/health