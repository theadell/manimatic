#!/bin/bash

# Set up AWS credentials (dummy credentials for LocalStack)
export AWS_ACCESS_KEY_ID=000000000000
export AWS_SECRET_ACCESS_KEY=000000000000
export AWS_DEFAULT_REGION=${REGION:-us-east-1}

DLQ_NAME="TaskQueueDLQ"
echo "Creating Dead Letter Queue: $DLQ_NAME"

DLQ_URL=$(awslocal sqs create-queue \
  --queue-name "$DLQ_NAME" \
  --region "$REGION" \
  --query 'QueueUrl' \
  --output text)

DLQ_ARN=$(awslocal sqs get-queue-attributes \
  --queue-url "$DLQ_URL" \
  --attribute-names QueueArn \
  --query 'Attributes.QueueArn' \
  --output text)


# Create the Task queue with DLQ configuration
echo "Creating Task Queue: $TASK_QUEUE_NAME with DLQ and max receives = 2"
awslocal sqs create-queue \
  --queue-name "$TASK_QUEUE_NAME" \
  --region "$REGION" \
  --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"$DLQ_ARN\\\",\\\"maxReceiveCount\\\":\\\"2\\\"}\"}"

# Create the result queue
echo "Creating Result Queue: $RESULT_QUEUE_NAME"
awslocal sqs create-queue \
  --queue-name "$RESULT_QUEUE_NAME" \
  --region "$REGION"

# Create the S3 bucket
echo "Creating Bucket: $BUCKET_NAME"
awslocal s3 mb \
  s3://"$BUCKET_NAME" \
  --region "$REGION"

echo "Initialization complete."
