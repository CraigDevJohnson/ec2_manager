package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Request represents the incoming Lambda request
type Request struct {
	Action       string `json:"action"`                  // start, stop, restart, change_type
	InstanceID   string `json:"instance_id"`             // EC2 instance ID
	InstanceType string `json:"instance_type,omitempty"` // For change_type action
}

// Response represents the Lambda response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// EC2Manager handles EC2 operations
type EC2Manager struct {
	client *ec2.Client
}

// NewEC2Manager creates a new EC2Manager
func NewEC2Manager(ctx context.Context) (*EC2Manager, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &EC2Manager{
		client: ec2.NewFromConfig(cfg),
	}, nil
}

// StartInstance starts an EC2 instance
func (m *EC2Manager) StartInstance(ctx context.Context, instanceID string) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := m.client.StartInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	if len(result.StartingInstances) > 0 {
		log.Printf("Instance %s state changing from %s to %s",
			instanceID,
			result.StartingInstances[0].PreviousState.Name,
			result.StartingInstances[0].CurrentState.Name)
	}

	return nil
}

// StopInstance stops an EC2 instance
func (m *EC2Manager) StopInstance(ctx context.Context, instanceID string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := m.client.StopInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	if len(result.StoppingInstances) > 0 {
		log.Printf("Instance %s state changing from %s to %s",
			instanceID,
			result.StoppingInstances[0].PreviousState.Name,
			result.StoppingInstances[0].CurrentState.Name)
	}

	return nil
}

// RestartInstance restarts an EC2 instance (stop then start)
func (m *EC2Manager) RestartInstance(ctx context.Context, instanceID string) error {
	// First, stop the instance
	if err := m.StopInstance(ctx, instanceID); err != nil {
		return err
	}

	// Wait for instance to be stopped
	log.Printf("Waiting for instance %s to stop...", instanceID)
	waiter := ec2.NewInstanceStoppedWaiter(m.client)
	maxWaitTime := 5 * time.Minute
	if err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, maxWaitTime); err != nil {
		return fmt.Errorf("error waiting for instance to stop: %w", err)
	}

	// Start the instance
	log.Printf("Starting instance %s...", instanceID)
	return m.StartInstance(ctx, instanceID)
}

// ChangeInstanceType changes the instance type of an EC2 instance
func (m *EC2Manager) ChangeInstanceType(ctx context.Context, instanceID, newInstanceType string) error {
	// Check current instance state
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := m.client.DescribeInstances(ctx, describeInput)
	if err != nil {
		return fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return fmt.Errorf("instance %s not found", instanceID)
	}

	instance := result.Reservations[0].Instances[0]
	currentState := instance.State.Name

	// Instance must be stopped to change type
	if currentState != types.InstanceStateNameStopped {
		log.Printf("Instance %s is in state %s, stopping it first...", instanceID, currentState)
		if err := m.StopInstance(ctx, instanceID); err != nil {
			return err
		}

		// Wait for instance to be stopped
		waiter := ec2.NewInstanceStoppedWaiter(m.client)
		maxWaitTime := 5 * time.Minute
		if err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}, maxWaitTime); err != nil {
			return fmt.Errorf("error waiting for instance to stop: %w", err)
		}
	}

	// Modify instance type
	modifyInput := &ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(instanceID),
		InstanceType: &types.AttributeValue{
			Value: aws.String(newInstanceType),
		},
	}

	_, err = m.client.ModifyInstanceAttribute(ctx, modifyInput)
	if err != nil {
		return fmt.Errorf("failed to modify instance type: %w", err)
	}

	log.Printf("Successfully changed instance %s type to %s", instanceID, newInstanceType)
	return nil
}

// HandleRequest processes the Lambda request
func HandleRequest(ctx context.Context, request Request) (Response, error) {
	log.Printf("Received request: action=%s, instance_id=%s, instance_type=%s",
		request.Action, request.InstanceID, request.InstanceType)

	// Validate request
	if request.InstanceID == "" {
		return Response{
			Success: false,
			Message: "Validation failed",
			Error:   "instance_id is required",
		}, nil
	}

	if request.Action == "" {
		return Response{
			Success: false,
			Message: "Validation failed",
			Error:   "action is required",
		}, nil
	}

	// Create EC2 manager
	manager, err := NewEC2Manager(ctx)
	if err != nil {
		return Response{
			Success: false,
			Message: "Failed to initialize EC2 manager",
			Error:   err.Error(),
		}, nil
	}

	// Execute the requested action
	var actionErr error
	var message string

	switch request.Action {
	case "start":
		actionErr = manager.StartInstance(ctx, request.InstanceID)
		message = fmt.Sprintf("Instance %s started successfully", request.InstanceID)

	case "stop":
		actionErr = manager.StopInstance(ctx, request.InstanceID)
		message = fmt.Sprintf("Instance %s stopped successfully", request.InstanceID)

	case "restart":
		actionErr = manager.RestartInstance(ctx, request.InstanceID)
		message = fmt.Sprintf("Instance %s restarted successfully", request.InstanceID)

	case "change_type":
		if request.InstanceType == "" {
			return Response{
				Success: false,
				Message: "Validation failed",
				Error:   "instance_type is required for change_type action",
			}, nil
		}
		actionErr = manager.ChangeInstanceType(ctx, request.InstanceID, request.InstanceType)
		message = fmt.Sprintf("Instance %s type changed to %s successfully", request.InstanceID, request.InstanceType)

	default:
		return Response{
			Success: false,
			Message: "Invalid action",
			Error:   fmt.Sprintf("unknown action: %s. Valid actions are: start, stop, restart, change_type", request.Action),
		}, nil
	}

	if actionErr != nil {
		return Response{
			Success: false,
			Message: fmt.Sprintf("Failed to execute action: %s", request.Action),
			Error:   actionErr.Error(),
		}, nil
	}

	return Response{
		Success: true,
		Message: message,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
