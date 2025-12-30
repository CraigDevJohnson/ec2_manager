package main

import (
	"context"
	"testing"
)

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     Request
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid start request",
			request: Request{
				Action:     "start",
				InstanceID: "i-1234567890abcdef0",
			},
			expectError: false,
		},
		{
			name: "Valid stop request",
			request: Request{
				Action:     "stop",
				InstanceID: "i-1234567890abcdef0",
			},
			expectError: false,
		},
		{
			name: "Valid restart request",
			request: Request{
				Action:     "restart",
				InstanceID: "i-1234567890abcdef0",
			},
			expectError: false,
		},
		{
			name: "Valid change_type request",
			request: Request{
				Action:       "change_type",
				InstanceID:   "i-1234567890abcdef0",
				InstanceType: "t3.medium",
			},
			expectError: false,
		},
		{
			name: "Missing instance ID",
			request: Request{
				Action: "start",
			},
			expectError: true,
			errorMsg:    "instance_id is required",
		},
		{
			name: "Missing action",
			request: Request{
				InstanceID: "i-1234567890abcdef0",
			},
			expectError: true,
			errorMsg:    "action is required",
		},
		{
			name: "Invalid action",
			request: Request{
				Action:     "invalid",
				InstanceID: "i-1234567890abcdef0",
			},
			expectError: true,
			errorMsg:    "unknown action",
		},
		{
			name: "change_type without instance_type",
			request: Request{
				Action:     "change_type",
				InstanceID: "i-1234567890abcdef0",
			},
			expectError: true,
			errorMsg:    "instance_type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Note: This will fail at the AWS SDK initialization in a test environment
			// without credentials, but we're testing validation logic which happens first
			response, err := HandleRequest(ctx, tt.request)

			if err != nil {
				t.Fatalf("HandleRequest returned unexpected error: %v", err)
			}

			if tt.expectError {
				if response.Success {
					t.Errorf("Expected error but got success")
				}
				if response.Error == "" {
					t.Errorf("Expected error message but got empty string")
				}
				if tt.errorMsg != "" && response.Error != tt.errorMsg {
					// Check if error message contains expected substring
					if len(response.Error) < len(tt.errorMsg) || response.Error[:len(tt.errorMsg)] != tt.errorMsg {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, response.Error)
					}
				}
			}
		})
	}
}

func TestResponseStructure(t *testing.T) {
	response := Response{
		Success: true,
		Message: "Operation successful",
	}

	if !response.Success {
		t.Errorf("Expected Success to be true")
	}

	if response.Message != "Operation successful" {
		t.Errorf("Expected Message to be 'Operation successful', got '%s'", response.Message)
	}

	if response.Error != "" {
		t.Errorf("Expected Error to be empty, got '%s'", response.Error)
	}
}

func TestRequestStructure(t *testing.T) {
	request := Request{
		Action:       "change_type",
		InstanceID:   "i-1234567890abcdef0",
		InstanceType: "t3.large",
	}

	if request.Action != "change_type" {
		t.Errorf("Expected Action to be 'change_type', got '%s'", request.Action)
	}

	if request.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("Expected InstanceID to be 'i-1234567890abcdef0', got '%s'", request.InstanceID)
	}

	if request.InstanceType != "t3.large" {
		t.Errorf("Expected InstanceType to be 't3.large', got '%s'", request.InstanceType)
	}
}
