package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// PerformerRequest represents the request from the Executor
type PerformerRequest struct {
	AVSAddress string          `json:"avsAddress"`
	TaskID     string          `json:"taskId"`
	Payload    json.RawMessage `json:"payload"`
}

// PerformerResponse represents the response to the Executor
type PerformerResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// LambdaAdapter handles routing requests to AWS Lambda
type LambdaAdapter struct {
	lambdaClient *lambda.Lambda
	functionName string
	apiEndpoint  string
}

// NewLambdaAdapter creates a new Lambda adapter
func NewLambdaAdapter() (*LambdaAdapter, error) {
	// Get configuration from environment
	avsAddress := os.Getenv("AVS_ADDRESS")
	operatorSetID := os.Getenv("OPERATOR_SET_ID")
	
	if avsAddress == "" || operatorSetID == "" {
		return nil, fmt.Errorf("AVS_ADDRESS and OPERATOR_SET_ID must be set")
	}

	// Construct deterministic function name
	functionName := fmt.Sprintf("avs-%s-opset-%s-performer", 
		strings.ToLower(avsAddress), 
		operatorSetID)

	// Create AWS session using default credential chain
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Get API endpoint from environment (optional)
	apiEndpoint := os.Getenv("LAMBDA_API_ENDPOINT")

	return &LambdaAdapter{
		lambdaClient: lambda.New(sess),
		functionName: functionName,
		apiEndpoint:  apiEndpoint,
	}, nil
}

// ProcessTask handles a task request
func (a *LambdaAdapter) ProcessTask(req PerformerRequest) (*PerformerResponse, error) {
	// If API endpoint is configured, use HTTP invocation
	if a.apiEndpoint != "" {
		return a.invokeViaAPI(req)
	}

	// Otherwise, use direct Lambda invocation
	return a.invokeDirectly(req)
}

// invokeDirectly calls Lambda function directly
func (a *LambdaAdapter) invokeDirectly(req PerformerRequest) (*PerformerResponse, error) {
	// Prepare Lambda payload
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke Lambda function
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(a.functionName),
		Payload:        payload,
		InvocationType: aws.String("RequestResponse"),
	}

	result, err := a.lambdaClient.Invoke(input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda: %w", err)
	}

	// Check for Lambda errors
	if result.FunctionError != nil {
		return &PerformerResponse{
			Success: false,
			Error:   fmt.Sprintf("Lambda error: %s", *result.FunctionError),
		}, nil
	}

	// Parse Lambda response
	var lambdaResp map[string]interface{}
	if err := json.Unmarshal(result.Payload, &lambdaResp); err != nil {
		return nil, fmt.Errorf("failed to parse Lambda response: %w", err)
	}

	// Convert to performer response
	resultJSON, _ := json.Marshal(lambdaResp)
	return &PerformerResponse{
		Success: true,
		Result:  resultJSON,
	}, nil
}

// invokeViaAPI calls Lambda function via API Gateway
func (a *LambdaAdapter) invokeViaAPI(req PerformerRequest) (*PerformerResponse, error) {
	// Prepare HTTP request
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", a.apiEndpoint+"/task", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Make HTTP request
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return &PerformerResponse{
			Success: false,
			Error:   fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	// Parse response
	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert to performer response
	resultJSON, _ := json.Marshal(apiResp)
	return &PerformerResponse{
		Success: true,
		Result:  resultJSON,
	}, nil
}

// HTTP handler for performer endpoint
func (a *LambdaAdapter) handlePerform(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req PerformerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to parse request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Process task
	resp, err := a.ProcessTask(req)
	if err != nil {
		log.Printf("Failed to process task: %v", err)
		resp = &PerformerResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Health check handler
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Create Lambda adapter
	adapter, err := NewLambdaAdapter()
	if err != nil {
		log.Fatalf("Failed to create Lambda adapter: %v", err)
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/perform", adapter.handlePerform)
	mux.HandleFunc("/health", handleHealth)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Lambda adapter starting on port %s", port)
		log.Printf("Using Lambda function: %s", adapter.functionName)
		if adapter.apiEndpoint != "" {
			log.Printf("Using API endpoint: %s", adapter.apiEndpoint)
		}
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Lambda adapter stopped")
}