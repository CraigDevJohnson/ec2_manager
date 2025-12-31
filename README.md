# ec2_manager

A Go binary designed to manage AWS EC2 instances through AWS Lambda. This service is intended to be called from a web application hosted on AWS Amplify, enabling users to control EC2 instances via simple button clicks.

## Features

- **Start Instance**: Power on a stopped EC2 instance
- **Stop Instance**: Power off a running EC2 instance
- **Restart Instance**: Stop and then start an EC2 instance
- **Change Instance Type**: Modify the instance type of an EC2 instance (automatically stops the instance if needed)

## Architecture

```
Web Page (AWS Amplify) → AWS Lambda (ec2_manager) → AWS EC2 API
```

The application receives JSON requests via Lambda, performs the requested EC2 operation using AWS SDK v2, and returns a JSON response indicating success or failure.

## Request Format

The Lambda function expects a JSON payload with the following structure:

```json
{
  "action": "start|stop|restart|change_type",
  "instance_id": "i-1234567890abcdef0",
  "instance_type": "t3.medium"
}
```

### Fields

- `action` (required): The operation to perform. Valid values:
  - `start` - Start a stopped instance
  - `stop` - Stop a running instance
  - `restart` - Stop and then start an instance
  - `change_type` - Change the instance type
  
- `instance_id` (required): The EC2 instance ID (e.g., `i-1234567890abcdef0`)

- `instance_type` (optional): Required only for `change_type` action. The new instance type (e.g., `t3.medium`, `t3.large`, etc.)

## Response Format

The Lambda function returns a JSON response:

```json
{
  "success": true,
  "message": "Instance i-1234567890abcdef0 started successfully",
  "error": ""
}
```

### Fields

- `success` (boolean): Whether the operation succeeded
- `message` (string): Human-readable message describing the result
- `error` (string): Error details if the operation failed (empty on success)

## Building

### Prerequisites

- Go 1.21+ installed
- AWS credentials configured (for testing with real AWS resources)
- `make` utility (optional, but recommended)

### Build for AWS Lambda

```bash
make build
```

This creates a `bootstrap` binary compiled for Linux AMD64 (Lambda's runtime environment) and packages it into `ec2_manager.zip` ready for Lambda deployment.

**Note**: AWS Lambda also supports ARM64 (Graviton2) which offers better price-performance. To build for ARM64, modify the Makefile `build` target to use `GOARCH=arm64` and change the runtime to `provided.al2023` in the deployment.

### Build for Local Testing

```bash
make build-local
```

This creates an `ec2_manager` binary for your local platform.

### Manual Build

If you prefer not to use Make:

```bash
# For Lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
zip ec2_manager.zip bootstrap

# For local
go build -o ec2_manager main.go
```

## Testing

Run the test suite:

```bash
make test
```

Run tests with coverage:

```bash
make test-coverage
```

This generates `coverage.html` which you can open in a browser to view detailed coverage information.

## Development

### Install Dependencies

```bash
make deps
```

### Format Code

```bash
make fmt
```

### Lint Code

Requires [golangci-lint](https://golangci-lint.run/usage/install/) to be installed:

```bash
make lint
```

### Clean Build Artifacts

```bash
make clean
```

## Deployment to AWS Lambda

1. Build the deployment package:
   ```bash
   make build
   ```

2. Create a new Lambda function in the AWS Console or via AWS CLI:
   ```bash
   aws lambda create-function \
     --function-name ec2-manager \
     --runtime provided.al2 \
     --role arn:aws:iam::YOUR_ACCOUNT:role/YOUR_LAMBDA_ROLE \
     --handler bootstrap \
     --timeout 360 \
     --zip-file fileb://ec2_manager.zip
   ```

   **Important**: Set the Lambda timeout to at least 360 seconds (6 minutes) to accommodate instance state transitions, which use 4-minute waiters internally.

3. Configure the Lambda function with appropriate IAM permissions (see below)

4. Set up an API Gateway or Lambda Function URL to make it accessible from your Amplify web application

5. **Configure CORS**: If using Lambda Function URL, enable CORS in the function configuration. If using API Gateway, configure CORS settings to allow requests from your Amplify domain. The Lambda function returns appropriate CORS headers in responses.

### Required IAM Permissions

The Lambda function's execution role needs the following EC2 permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:StartInstances",
        "ec2:StopInstances",
        "ec2:DescribeInstances",
        "ec2:ModifyInstanceAttribute"
      ],
      "Resource": "*"
    }
  ]
}
```

For production, consider restricting the `Resource` field to specific instance ARNs or using condition keys for additional security.

## Example Usage from Web Application

### Using Fetch API

```javascript
async function manageInstance(action, instanceId, instanceType = null) {
  const payload = {
    action: action,
    instance_id: instanceId
  };
  
  if (instanceType) {
    payload.instance_type = instanceType;
  }
  
  try {
    const response = await fetch('YOUR_LAMBDA_URL', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload)
    });
    
    const result = await response.json();
    
    if (result.success) {
      console.log('Success:', result.message);
    } else {
      console.error('Error:', result.error);
    }
  } catch (error) {
    console.error('Request failed:', error);
  }
}

// Example button handlers
document.getElementById('startBtn').addEventListener('click', () => {
  manageInstance('start', 'i-1234567890abcdef0');
});

document.getElementById('stopBtn').addEventListener('click', () => {
  manageInstance('stop', 'i-1234567890abcdef0');
});

document.getElementById('restartBtn').addEventListener('click', () => {
  manageInstance('restart', 'i-1234567890abcdef0');
});

document.getElementById('changeTypeBtn').addEventListener('click', () => {
  manageInstance('change_type', 'i-1234567890abcdef0', 't3.medium');
});
```

## Security Considerations

1. **Authentication**: Implement proper authentication/authorization in your API Gateway or Lambda authorizer before calling this function
2. **Instance Access**: Consider implementing instance-level access control based on user identity
3. **Rate Limiting**: Implement rate limiting to prevent abuse
4. **Logging**: All operations are logged via CloudWatch Logs for audit purposes
5. **Least Privilege**: Grant only necessary EC2 permissions and consider restricting to specific instances

## Future Enhancements

This codebase is designed to be extensible. Potential future features include:

- Reboot instance
- Terminate instance
- Create instance snapshot
- Attach/detach volumes
- Update security groups
- View instance metrics
- Schedule instance start/stop times

## License

See [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Submit a pull request
