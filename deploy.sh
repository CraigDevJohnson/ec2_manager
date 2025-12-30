#!/bin/bash

# EC2 Manager Lambda Deployment Script
# This script helps deploy the ec2_manager to AWS Lambda

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
FUNCTION_NAME="${LAMBDA_FUNCTION_NAME:-ec2-manager}"
RUNTIME="provided.al2"
HANDLER="bootstrap"
TIMEOUT="${LAMBDA_TIMEOUT:-300}"
MEMORY="${LAMBDA_MEMORY:-256}"
REGION="${AWS_REGION:-us-east-1}"

echo -e "${GREEN}EC2 Manager Lambda Deployment${NC}"
echo "============================================"

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

if ! command -v aws &> /dev/null; then
    echo -e "${RED}Error: AWS CLI is not installed${NC}"
    echo "Please install AWS CLI: https://aws.amazon.com/cli/"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go: https://golang.org/doc/install"
    exit 1
fi

if ! command -v make &> /dev/null; then
    echo -e "${RED}Error: make is not installed${NC}"
    exit 1
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}Error: AWS credentials are not configured${NC}"
    echo "Please run: aws configure"
    exit 1
fi

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo -e "${GREEN}✓ AWS Account: ${ACCOUNT_ID}${NC}"

# Check for IAM role
if [ -z "${LAMBDA_ROLE_ARN}" ]; then
    echo -e "\n${YELLOW}Warning: LAMBDA_ROLE_ARN environment variable not set${NC}"
    echo "The Lambda execution role should have the following policies:"
    echo "  - AWSLambdaBasicExecutionRole (for CloudWatch Logs)"
    echo "  - Custom policy for EC2 permissions (StartInstances, StopInstances, DescribeInstances, ModifyInstanceAttribute)"
    echo ""
    read -p "Enter the Lambda execution role ARN: " LAMBDA_ROLE_ARN
    
    if [ -z "${LAMBDA_ROLE_ARN}" ]; then
        echo -e "${RED}Error: Lambda role ARN is required${NC}"
        exit 1
    fi
fi

# Validate Lambda role ARN format
if ! [[ "${LAMBDA_ROLE_ARN}" =~ ^arn:aws:iam::[0-9]{12}:role/.+ ]]; then
    echo -e "${RED}Error: LAMBDA_ROLE_ARN is not a valid IAM role ARN${NC}"
    echo "Expected format: arn:aws:iam::123456789012:role/YourRoleName"
    exit 1
fi
echo -e "${GREEN}✓ Lambda Role: ${LAMBDA_ROLE_ARN}${NC}"

# Build the deployment package
echo -e "\n${YELLOW}Building deployment package...${NC}"
make build

if [ ! -f "ec2_manager.zip" ]; then
    echo -e "${RED}Error: Deployment package not found${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Deployment package created${NC}"

# Check if Lambda function exists
echo -e "\n${YELLOW}Checking if Lambda function exists...${NC}"
if aws lambda get-function --function-name "${FUNCTION_NAME}" --region "${REGION}" &> /dev/null; then
    echo -e "${YELLOW}Function exists, updating...${NC}"
    
    # Update function code
    aws lambda update-function-code \
        --function-name "${FUNCTION_NAME}" \
        --zip-file fileb://ec2_manager.zip \
        --region "${REGION}"
    
    echo -e "${GREEN}✓ Function code updated${NC}"
    
    # Update function configuration
    aws lambda update-function-configuration \
        --function-name "${FUNCTION_NAME}" \
        --timeout "${TIMEOUT}" \
        --memory-size "${MEMORY}" \
        --region "${REGION}" > /dev/null
    
    echo -e "${GREEN}✓ Function configuration updated${NC}"
else
    echo -e "${YELLOW}Function does not exist, creating...${NC}"
    
    # Create function
    aws lambda create-function \
        --function-name "${FUNCTION_NAME}" \
        --runtime "${RUNTIME}" \
        --role "${LAMBDA_ROLE_ARN}" \
        --handler "${HANDLER}" \
        --timeout "${TIMEOUT}" \
        --memory-size "${MEMORY}" \
        --zip-file fileb://ec2_manager.zip \
        --region "${REGION}"
    
    echo -e "${GREEN}✓ Function created${NC}"
fi

# Get function details
FUNCTION_ARN=$(aws lambda get-function --function-name "${FUNCTION_NAME}" --region "${REGION}" --query 'Configuration.FunctionArn' --output text)

echo -e "\n${GREEN}============================================${NC}"
echo -e "${GREEN}Deployment successful!${NC}"
echo -e "${GREEN}============================================${NC}"
echo -e "Function Name: ${FUNCTION_NAME}"
echo -e "Function ARN:  ${FUNCTION_ARN}"
echo -e "Region:        ${REGION}"
echo -e ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Test the function using the AWS Console or AWS CLI"
echo -e "2. Set up API Gateway or Function URL for web access"
echo -e "3. Configure CORS if calling from a web application"
echo -e ""
echo -e "${YELLOW}Example test command:${NC}"
echo -e "aws lambda invoke \\"
echo -e "  --function-name ${FUNCTION_NAME} \\"
echo -e "  --payload '{\"action\":\"start\",\"instance_id\":\"i-xxxxx\"}' \\"
echo -e "  --region ${REGION} \\"
echo -e "  response.json"
