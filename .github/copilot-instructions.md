# EC2 Manager - GitHub Copilot Instructions

## Project Overview
This is a Go-based AWS Lambda function that manages EC2 instances through a web application. The service handles start, stop, restart, and instance type change operations via JSON API requests. It's designed to be deployed on AWS Lambda and called from a web application hosted on AWS Amplify.

## Architecture
- **Deployment Target**: AWS Lambda with `provided.al2` runtime
- **Trigger**: API Gateway or Lambda Function URL
- **Client**: Web application (AWS Amplify) making JSON HTTP requests
- **AWS Services**: EC2 API via AWS SDK v2

## Technology Stack
- **Language**: Go 1.23+
- **AWS SDK**: AWS SDK for Go v2 (`github.com/aws/aws-sdk-go-v2`)
- **Lambda Runtime**: `github.com/aws/aws-lambda-go`
- **Build Target**: Linux AMD64 (can also build for ARM64/Graviton2)
- **Testing**: Go standard testing package

## Coding Standards and Conventions

### Go Style
- Follow standard Go conventions and idioms
- Use `go fmt` for code formatting (tabs for indentation)
- Run `golangci-lint` for linting (when available)
- Keep functions focused and single-purpose
- Use meaningful variable names

### Error Handling
- Always wrap errors with context using `fmt.Errorf` with `%w` verb
- Log errors and state changes using the `log` package
- Return user-friendly error messages in API responses
- Never expose sensitive AWS details in error messages

### AWS SDK Usage
- Use AWS SDK v2 (not v1)
- Always pass `context.Context` to AWS API calls
- Use waiters for operations that require state transitions (e.g., `ec2.NewInstanceStoppedWaiter`)
- Set appropriate timeout values for waiters (currently 4 minutes)

### Logging
- Use Go's standard `log` package
- Log all state changes and important operations
- Include relevant IDs (instance ID, action) in log messages
- Example: `log.Printf("Instance %s state changing from %s to %s", instanceID, prevState, currState)`

### JSON Structure
- Use struct tags for JSON serialization: `json:"field_name"`
- Use `omitempty` for optional fields: `json:"field_name,omitempty"`
- Maintain consistent request/response structures as documented in README

### CORS
- All responses must include CORS headers for cross-origin requests from Amplify
- Standard headers: `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods: POST, OPTIONS`, `Access-Control-Allow-Headers: Content-Type`

## Build and Deployment

### Building
- Use `make build` to build for Lambda (Linux AMD64)
- Use `make build-arm64` for ARM64/Graviton2
- Use `make build-local` for local development
- Output: `bootstrap` binary (for Lambda) or `ec2_manager` binary (for local)
- Package: `ec2_manager.zip` for Lambda deployment

### Build Configuration
- Always use `CGO_ENABLED=0` for Lambda builds
- Set `GOOS=linux` and `GOARCH=amd64` (or `arm64`) for Lambda builds
- Binary must be named `bootstrap` for custom runtime

## Testing

### Test Requirements
- All new functionality must include tests
- Use table-driven tests for multiple test cases
- Test validation logic separately from AWS API calls
- Skip tests requiring AWS credentials in CI/CD using: `if os.Getenv("AWS_REGION") == "" { t.Skip(...) }`

### Running Tests
- `make test` - Run all tests
- `make test-coverage` - Generate coverage report (creates `coverage.html`)
- Tests should not require actual AWS resources unless specifically testing AWS integration

### Test Structure
- Use subtests with `t.Run()` for organized test cases
- Test both success and error cases
- Validate response structure (Success, Message, Error fields)
- Test CORS headers are properly set

## Security Considerations

### IAM Permissions
- Lambda execution role needs: `ec2:StartInstances`, `ec2:StopInstances`, `ec2:DescribeInstances`, `ec2:ModifyInstanceAttribute`
- Follow principle of least privilege
- Consider restricting Resource to specific instance ARNs in production

### Input Validation
- Always validate `instance_id` and `action` fields are present
- Validate `instance_type` is present for `change_type` action
- Validate action is one of: `start`, `stop`, `restart`, `change_type`
- Return clear validation error messages

### Best Practices
- Never log AWS credentials or sensitive information
- Implement authentication/authorization at API Gateway level (not in Lambda)
- All operations are logged to CloudWatch for audit purposes
- Rate limiting should be implemented at API Gateway level

## Lambda Configuration
- **Timeout**: Set to 360 seconds (6 minutes) minimum to accommodate instance state transitions
- **Memory**: Default settings are sufficient
- **Environment Variables**: AWS credentials are provided automatically by Lambda execution role
- **Runtime**: Use `provided.al2` for custom Go runtime

## API Contract

### Request Format
```json
{
  "action": "start|stop|restart|change_type",
  "instance_id": "i-1234567890abcdef0",
  "instance_type": "t3.medium"  // Required only for change_type
}
```

### Response Format
```json
{
  "success": true,
  "message": "Instance i-1234567890abcdef0 started successfully",
  "error": ""  // Empty string on success, error message on failure
}
```

## Code Organization
- **Single file architecture**: `main.go` contains all code (appropriate for this size project)
- **Main types**: `Request`, `Response`, `EC2Manager`
- **Main functions**: `HandleRequest` (Lambda handler), `NewEC2Manager`, instance operation methods
- **Tests**: `main_test.go` with table-driven tests

## Future Enhancement Patterns
When adding new features, maintain consistency with existing patterns:
- Add new actions to the switch statement in `HandleRequest`
- Add corresponding methods to `EC2Manager`
- Include validation for new required fields
- Add tests for new functionality
- Update README documentation
- Consider if waiters are needed for state transitions

## Common Operations

### Adding a New EC2 Action
1. Add the action to `HandleRequest` switch statement
2. Implement method on `EC2Manager` struct
3. Add validation for required fields
4. Include appropriate logging
5. Use waiters if operation requires state transition
6. Add tests (both unit and validation)
7. Update README with new action documentation

### Modifying Instance Checks
- Always use `ec2.DescribeInstances` to check current state
- Check for empty results before accessing instance data
- Use `types.InstanceStateNameStopped`, `types.InstanceStateNameStopping`, etc. for state comparisons

## Dependencies Management
- Use `go mod download` and `go mod tidy` to manage dependencies
- Pin major versions in `go.mod`
- Update AWS SDK cautiously and test thoroughly
- Use `make deps` to install and tidy dependencies
