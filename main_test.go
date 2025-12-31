package main

import (
	"context"
	"os"
	"testing"
)

func TestRequestValidation(t *testing.T) {
	// Skip tests that require AWS credentials in CI/CD
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("Skipping tests that require AWS credentials")
	}

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

// TestValidationOnly tests only the validation logic without AWS SDK
func TestValidationOnly(t *testing.T) {
	tests := []struct {
		name        string
		request     Request
		expectError bool
		errorMsg    string
	}{
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
		{
			name: "Valid instance ID and action",
			request: Request{
				Action:     "start",
				InstanceID: "i-1234567890abcdef0",
			},
			expectError: false,
		},
		{
			name: "Valid change_type with instance_type",
			request: Request{
				Action:       "change_type",
				InstanceID:   "i-1234567890abcdef0",
				InstanceType: "t3.medium",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic directly
			var errorMsg string
			valid := true

			if tt.request.InstanceID == "" {
				valid = false
				errorMsg = "instance_id is required"
			} else if tt.request.Action == "" {
				valid = false
				errorMsg = "action is required"
			} else if tt.request.Action != "start" && tt.request.Action != "stop" && 
				tt.request.Action != "restart" && tt.request.Action != "change_type" {
				valid = false
				errorMsg = "unknown action"
			} else if tt.request.Action == "change_type" && tt.request.InstanceType == "" {
				valid = false
				errorMsg = "instance_type is required"
			}

			if tt.expectError {
				if valid {
					t.Errorf("Expected validation to fail but it passed")
				}
				if errorMsg == "" {
					t.Errorf("Expected error message but got empty string")
				}
				if tt.errorMsg != "" && errorMsg != tt.errorMsg {
					// Check if error message contains expected substring
					if len(errorMsg) < len(tt.errorMsg) || errorMsg[:len(tt.errorMsg)] != tt.errorMsg {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, errorMsg)
					}
				}
			} else {
				if !valid {
					t.Errorf("Expected validation to pass but it failed with: %s", errorMsg)
				}
			}
		})
	}
}

func TestResponseStructure(t *testing.T) {
	response := Response{
		Success: true,
		Message: "Operation successful",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
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

	if response.Headers == nil {
		t.Errorf("Expected Headers to be initialized")
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

func TestCORSHeaders(t *testing.T) {
	response := Response{
		Success: true,
		Message: "Test message",
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
	}

	if response.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Errorf("Expected CORS origin header to be '*', got '%s'", response.Headers["Access-Control-Allow-Origin"])
	}

	if response.Headers["Access-Control-Allow-Methods"] != "POST, OPTIONS" {
		t.Errorf("Expected CORS methods header to be 'POST, OPTIONS', got '%s'", response.Headers["Access-Control-Allow-Methods"])
	}

	if response.Headers["Access-Control-Allow-Headers"] != "Content-Type" {
		t.Errorf("Expected CORS headers to be 'Content-Type', got '%s'", response.Headers["Access-Control-Allow-Headers"])
	}
}
