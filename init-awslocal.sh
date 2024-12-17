#!/bin/bash

# Set up AWS credentials (dummy credentials for LocalStack)
export AWS_ACCESS_KEY_ID=000000000000
export AWS_SECRET_ACCESS_KEY=000000000000
export AWS_DEFAULT_REGION=${REGION:-us-east-1}

# Create the task queue
echo "Creating Task Queue: $TASK_QUEUE_NAME"
awslocal sqs create-queue \
  --queue-name "$TASK_QUEUE_NAME" \
  --region "$REGION"

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
