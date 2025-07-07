package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

// ServiceAttestation represents the attestation state of a service
type ServiceAttestation struct {
	ID           string    `json:"id"`
	ServiceName  string    `json:"service_name"`
	Version      string    `json:"version"`
	BinaryHash   string    `json:"binary_hash"`
	ConfigHash   string    `json:"config_hash"`
	Timestamp    time.Time `json:"timestamp"`
	ValidUntil   time.Time `json:"valid_until"`
	Platform     Platform  `json:"platform"`
	SEVReport    []byte    `json:"sev_report,omitempty"`
	PublicKey    string    `json:"public_key"`
}

// Platform represents the TEE platform details
type Platform struct {
	Type         string `json:"type"` // "SEV-SNP", "TDX", "SGX"
	Measurement  string `json:"measurement"`
	ChipID       string `json:"chip_id,omitempty"`
	TCBVersion   string `json:"tcb_version,omitempty"`
}

// SignedOutput represents output with attestation proof
type SignedOutput struct {
	Data        interface{}         `json:"data"`
	Attestation ServiceAttestation  `json:"attestation"`
	OutputProof OutputProof         `json:"output_proof"`
}

// OutputProof binds output to attestation
type OutputProof struct {
	RequestID   string    `json:"request_id"`
	Timestamp   time.Time `json:"timestamp"`
	Nonce       string    `json:"nonce"`
	InputHash   string    `json:"input_hash"`
	OutputHash  string    `json:"output_hash"`
	Signature   string    `json:"signature"`
}

// ServiceAttestor handles attestation for a service
type ServiceAttestor struct {
	serviceName string
	version     string

	currentAttestation *ServiceAttestation
	mu                 sync.RWMutex

	platform           PlatformAttestor
}

// PlatformAttestor interface for different TEE types
type PlatformAttestor interface {
	GetPlatformAttestation(reportData []byte) (*Platform, []byte, error)
	GetMeasurement() (string, error)
}

// NewServiceAttestor creates a new attestor
func NewServiceAttestor(serviceName, version string) (*ServiceAttestor, error) {
	// Detect platform
	platform, err := detectPlatform()
	if err != nil {
		return nil, fmt.Errorf("failed to detect platform: %w", err)
	}

	attestor := &ServiceAttestor{
		serviceName: serviceName,
		version:     version,
		platform:    platform,
	}

	// Initial attestation
	if err := attestor.RefreshAttestation(); err != nil {
		return nil, fmt.Errorf("failed initial attestation: %w", err)
	}

	return attestor, nil
}

// RefreshAttestation updates the current attestation
func (a *ServiceAttestor) RefreshAttestation() error {
	// Measure current binary
	binaryHash, err := a.measureBinary()
	if err != nil {
		return fmt.Errorf("failed to measure binary: %w", err)
	}

	// Measure configuration
	configHash := a.measureConfig()

	// Create attestation ID
	attestationID := generateAttestationID()

	// Get platform measurement
	measurement, err := a.platform.GetMeasurement()
	if err != nil {
		return fmt.Errorf("failed to get platform measurement: %w", err)
	}

	// Create report data
	reportData := append([]byte(a.serviceName), []byte(binaryHash)...)

	// Get platform attestation
	platform, sevReport, err := a.platform.GetPlatformAttestation(reportData)
	if err != nil {
		return fmt.Errorf("failed to get platform attestation: %w", err)
	}
	platform.Measurement = measurement

	// Create new attestation
	newAttestation := &ServiceAttestation{
		ID:          attestationID,
		ServiceName: a.serviceName,
		Version:     a.version,
		BinaryHash:  binaryHash,
		ConfigHash:  configHash,
		Timestamp:   time.Now(),
		ValidUntil:  time.Now().Add(1 * time.Hour),
		Platform:    *platform,
		SEVReport:   sevReport,
		PublicKey:   a.getPublicKey(),
	}

	a.mu.Lock()
	a.currentAttestation = newAttestation
	a.mu.Unlock()

	return nil
}

// GetCurrentAttestation returns the current attestation
func (a *ServiceAttestor) GetCurrentAttestation() ServiceAttestation {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentAttestation == nil {
		return ServiceAttestation{}
	}

	return *a.currentAttestation
}

// Version returns the service version
func (a *ServiceAttestor) Version() string {
	return a.version
}

func (a *ServiceAttestor) measureBinary() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(execPath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (a *ServiceAttestor) measureConfig() string {
	// Hash environment variables and config files
	configData := []byte{}

	// Add relevant env vars
	for _, key := range []string{"CONFIG_PATH", "SERVICE_CONFIG"} {
		if val := os.Getenv(key); val != "" {
			configData = append(configData, []byte(key+"="+val)...)
		}
	}

	hash := sha256.Sum256(configData)
	return hex.EncodeToString(hash[:])
}

func (a *ServiceAttestor) getPublicKey() string {
	// In real implementation, derive from signing key
	return "mock-public-key"
}

func generateAttestationID() string {
	b := make([]byte, 16)
	return hex.EncodeToString(b)
}

func detectPlatform() (PlatformAttestor, error) {
	// Check for SEV
	if _, err := os.Stat("/dev/sev-guest"); err == nil {
		return NewSEVAttestor()
	}

	// Fallback to mock for development
	return NewMockAttestor(), nil
}