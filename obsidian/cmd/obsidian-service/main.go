package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hourglass/obsidian/pkg/attestation"
	"github.com/hourglass/obsidian/pkg/service"
)

type ObsidianService struct {
	attestor    *attestation.ServiceAttestor
	signingKey  *ecdsa.PrivateKey
	serviceCore *service.Core
}

func main() {
	log.Println("Starting Obsidian Attestation Service...")

	// Initialize service
	svc, err := NewObsidianService()
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Start attestation refresh loop
	go svc.attestationRefreshLoop()

	// Setup HTTP server
	router := mux.NewRouter()

	// Health and attestation endpoints
	router.HandleFunc("/health", svc.healthHandler).Methods("GET")
	router.HandleFunc("/attestation", svc.attestationHandler).Methods("GET")

	// Service endpoints
	router.HandleFunc("/api/compute", svc.computeHandler).Methods("POST")
	router.HandleFunc("/api/verify/{id}", svc.verifyHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func NewObsidianService() (*ObsidianService, error) {
	// Generate or load signing key
	signingKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key: %w", err)
	}

	// Initialize attestor
	attestor, err := attestation.NewServiceAttestor(
		"obsidian-service",
		os.Getenv("VERSION"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attestor: %w", err)
	}

	// Initialize service core
	serviceCore := service.NewCore()

	return &ObsidianService{
		attestor:    attestor,
		signingKey:  signingKey,
		serviceCore: serviceCore,
	}, nil
}

// Health check endpoint
func (s *ObsidianService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now(),
		"version": s.attestor.Version(),
	})
}

// Attestation endpoint - returns current service attestation
func (s *ObsidianService) attestationHandler(w http.ResponseWriter, r *http.Request) {
	attestation := s.attestor.GetCurrentAttestation()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attestation)
}

// Main compute endpoint - processes requests with attestation
func (s *ObsidianService) computeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req service.ComputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Add nonce from request for freshness
	req.Nonce = r.Header.Get("X-Nonce")
	if req.Nonce == "" {
		req.Nonce = generateNonce()
	}

	// Process request (non-deterministic computation)
	result, err := s.serviceCore.ProcessCompute(req)
	if err != nil {
		http.Error(w, "Computation failed", http.StatusInternalServerError)
		return
	}

	// Create signed output with attestation
	signedOutput := s.createSignedOutput(req, result)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Attestation-Id", signedOutput.Attestation.ID)
	json.NewEncoder(w).Encode(signedOutput)
}

// Verify endpoint - allows verification of previous outputs
func (s *ObsidianService) verifyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	outputID := vars["id"]

	// Retrieve and verify output
	verified, err := s.serviceCore.VerifyOutput(outputID)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"verified": verified,
		"output_id": outputID,
	})
}

func (s *ObsidianService) createSignedOutput(req service.ComputeRequest, result *service.ComputeResult) *attestation.SignedOutput {
	// Get current attestation
	currentAttestation := s.attestor.GetCurrentAttestation()

	// Create output proof
	proof := attestation.OutputProof{
		RequestID:  result.ID,
		Timestamp:  time.Now(),
		Nonce:      req.Nonce,
		InputHash:  hashData(req),
		OutputHash: hashData(result),
	}

	// Sign the proof
	proofData, _ := json.Marshal(proof)
	signature := s.sign(proofData)
	proof.Signature = hex.EncodeToString(signature)

	return &attestation.SignedOutput{
		Data:        result,
		Attestation: currentAttestation,
		OutputProof: proof,
	}
}

func (s *ObsidianService) sign(data []byte) []byte {
	hash := sha256.Sum256(data)
	r, s_, err := ecdsa.Sign(rand.Reader, s.signingKey, hash[:])
	if err != nil {
		log.Printf("Signing failed: %v", err)
		return nil
	}

	signature := append(r.Bytes(), s_.Bytes()...)
	return signature
}

func (s *ObsidianService) attestationRefreshLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.attestor.RefreshAttestation(); err != nil {
				log.Printf("Failed to refresh attestation: %v", err)
			} else {
				log.Println("Attestation refreshed successfully")
			}
		}
	}
}

func hashData(v interface{}) string {
	data, _ := json.Marshal(v)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}