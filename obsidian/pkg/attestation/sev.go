package attestation

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
)

// SEVAttestor implements PlatformAttestor for AMD SEV-SNP
type SEVAttestor struct {
	devicePath string
}

// NewSEVAttestor creates a new SEV attestor
func NewSEVAttestor() (*SEVAttestor, error) {
	devicePath := "/dev/sev-guest"
	if _, err := os.Stat(devicePath); err != nil {
		return nil, fmt.Errorf("SEV device not found: %w", err)
	}

	return &SEVAttestor{
		devicePath: devicePath,
	}, nil
}

// GetPlatformAttestation generates SEV attestation report
func (s *SEVAttestor) GetPlatformAttestation(reportData []byte) (*Platform, []byte, error) {
	// In real implementation, use ioctl to get SEV report
	// For now, return mock data

	platform := &Platform{
		Type:       "SEV-SNP",
		ChipID:     "mock-chip-id",
		TCBVersion: "mock-tcb-version",
	}

	// Mock SEV report
	report := append([]byte("SEV-REPORT-"), reportData...)

	return platform, report, nil
}

// GetMeasurement returns the current VM measurement
func (s *SEVAttestor) GetMeasurement() (string, error) {
	// In real implementation, read from SEV device
	// For now, return mock measurement
	return "7b068c0c3ac29afe264134536b9be26f1e4ccd575b88d3e3be77e768414ce98d", nil
}

// MockAttestor for development
type MockAttestor struct{}

func NewMockAttestor() *MockAttestor {
	return &MockAttestor{}
}

func (m *MockAttestor) GetPlatformAttestation(reportData []byte) (*Platform, []byte, error) {
	platform := &Platform{
		Type:       "MOCK",
		ChipID:     "dev-chip-id",
		TCBVersion: "dev-tcb",
	}

	report := []byte("MOCK-REPORT")

	return platform, report, nil
}

func (m *MockAttestor) GetMeasurement() (string, error) {
	return "0000000000000000000000000000000000000000000000000000000000000000", nil
}