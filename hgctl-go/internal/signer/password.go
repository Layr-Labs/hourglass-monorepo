package signer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// StdinPasswordProvider reads passwords from stdin
type StdinPasswordProvider struct {
	passwords map[string]string
}

func NewStdinPasswordProvider() *StdinPasswordProvider {
	return &StdinPasswordProvider{
		passwords: make(map[string]string),
	}
}

// ReadPasswordsFromStdin reads passwords in the format KEY=value from stdin
func (p *StdinPasswordProvider) ReadPasswordsFromStdin() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			p.passwords[parts[0]] = parts[1]
		}
	}
	return scanner.Err()
}

// SetPassword sets a password for a specific key
func (p *StdinPasswordProvider) SetPassword(key, password string) {
	p.passwords[key] = password
}

func (p *StdinPasswordProvider) GetPassword(keystoreName string) (string, error) {
	// Check if we have a password for this specific keystore
	if pwd, ok := p.passwords[keystoreName]; ok {
		return pwd, nil
	}

	// Check for generic BLS/ECDSA passwords
	if strings.Contains(keystoreName, "BLS") {
		if pwd, ok := p.passwords["BLS_PASSWORD"]; ok {
			return pwd, nil
		}
	}
	if strings.Contains(keystoreName, "ECDSA") {
		if pwd, ok := p.passwords["ECDSA_PASSWORD"]; ok {
			return pwd, nil
		}
	}

	return "", fmt.Errorf("no password found for keystore %s", keystoreName)
}

// InteractivePasswordProvider prompts the user for passwords
type InteractivePasswordProvider struct{}

func NewInteractivePasswordProvider() *InteractivePasswordProvider {
	return &InteractivePasswordProvider{}
}

func (p *InteractivePasswordProvider) GetPassword(keystoreName string) (string, error) {
	fmt.Printf("Enter password for keystore %s: ", keystoreName)

	// Read password without echoing
	password, err := term.ReadPassword(syscall.Stdin)
	fmt.Println() // Add newline after password input

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return string(password), nil
}

// EnvironmentPasswordProvider reads passwords from environment variables
type EnvironmentPasswordProvider struct{}

func NewEnvironmentPasswordProvider() *EnvironmentPasswordProvider {
	return &EnvironmentPasswordProvider{}
}

func (p *EnvironmentPasswordProvider) GetPassword(keystoreName string) (string, error) {
	// Try specific keystore name first
	if pwd := os.Getenv(keystoreName + "_PASSWORD"); pwd != "" {
		return pwd, nil
	}

	// Check for generic BLS/ECDSA passwords
	if strings.Contains(keystoreName, "BLS") {
		if pwd := os.Getenv("BLS_PASSWORD"); pwd != "" {
			return pwd, nil
		}
	}
	if strings.Contains(keystoreName, "ECDSA") {
		if pwd := os.Getenv("ECDSA_PASSWORD"); pwd != "" {
			return pwd, nil
		}
	}

	return "", fmt.Errorf("no password found in environment for keystore %s", keystoreName)
}

// CombinedPasswordProvider tries multiple providers in order
type CombinedPasswordProvider struct {
	providers []PasswordProvider
}

func NewCombinedPasswordProvider(providers ...PasswordProvider) *CombinedPasswordProvider {
	return &CombinedPasswordProvider{
		providers: providers,
	}
}

func (p *CombinedPasswordProvider) GetPassword(keystoreName string) (string, error) {
	for _, provider := range p.providers {
		pwd, err := provider.GetPassword(keystoreName)
		if err == nil {
			return pwd, nil
		}
	}

	// If all providers fail, return empty password (common default)
	return "", nil
}

// MapPasswordProvider provides passwords from a map
type MapPasswordProvider struct {
	passwords map[string]string
}

func NewMapPasswordProvider(passwords map[string]string) *MapPasswordProvider {
	return &MapPasswordProvider{
		passwords: passwords,
	}
}

func (p *MapPasswordProvider) GetPassword(keystoreName string) (string, error) {
	// Check if we have a password for this specific keystore
	if pwd, ok := p.passwords[keystoreName]; ok {
		return pwd, nil
	}

	// Check for generic BLS/ECDSA passwords
	if strings.Contains(keystoreName, "BLS") {
		if pwd, ok := p.passwords["BLS_PASSWORD"]; ok {
			return pwd, nil
		}
	}
	if strings.Contains(keystoreName, "ECDSA") {
		if pwd, ok := p.passwords["ECDSA_PASSWORD"]; ok {
			return pwd, nil
		}
	}

	return "", fmt.Errorf("no password found for keystore %s", keystoreName)
}
