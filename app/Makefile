# Makefile for Local Development

# LocalStack Configuration
REGION := eu-central-1
TASK_QUEUE_NAME := manim-task-queue
RESULT_QUEUE_NAME := manim-result-queue
BUCKET_NAME := manim-worker-bucket

# Environment Variables
ENV_FILE := .env
ENV_LOCAL_FILE := .env.local
# Start LocalStack in background
localstack-start:
	LOCALSTACK_HOST=localhost SERVICES=s3,sqs,ssm localstack start -d

# Stop LocalStack background process
localstack-stop:
	localstack stop

# Create SQS Queue
create-task-queue:
	awslocal sqs create-queue \
		--queue-name $(TASK_QUEUE_NAME) \
		--region $(REGION)

create-result-queue:
	awslocal sqs create-queue \
		--queue-name $(RESULT_QUEUE_NAME) \
		--region $(REGION)

create-bucket:
	awslocal s3 mb \
		s3://$(BUCKET_NAME) \
		--region $(REGION)

# Source environment variables
source-env:
	@[ -f $(ENV_FILE) ] && export $$(grep -v '^#' $(ENV_FILE) | xargs) || echo "No .env file found"
source-local-env:
	@[ -f $(ENV_LOCAL_FILE) ] && export $$(grep -v '^#' $(ENV_LOCAL_FILE) | xargs) || echo "No .env file found"
# Run worker locally
run-worker: source-env
	go run ./cmd/worker/main.go -log-level=debug
run-worker-local:
	go run ./cmd/worker/main.go -log-level=debug -localstack
run-api: source-env
	go run ./cmd/api/main.go -log-level=debug
run-api-local: source-local-env
	go run ./cmd/api/main.go -log-level=debug -localstack

# Full local setup
setup: localstack-start sleep create-task-queue create-result-queue create-bucket

sleep:
	sleep 2

send-message:
	echo '{"type": "script", "session_id": "session_123", "status": "success", "content": "from manim import *\\n\\nclass CreateCircle(Scene):\\n    def construct(self):\\n        circle = Circle()  \\n        circle.set_fill(PINK, opacity=0.5)  \\n        self.play(Create(circle))  "}' | \
		awslocal sqs send-message \
		--queue-url http://localhost:4566/000000000000/$(TASK_QUEUE_NAME) \
		--message-body file:///dev/stdin \
		--region $(REGION)


# Clean LocalStack resources
clean: localstack-stop


.PHONY: localstack-start localstack-stop create-task-queue create-result-queue create-bucket source-env run-worker run-api setup sleep send-message clean
