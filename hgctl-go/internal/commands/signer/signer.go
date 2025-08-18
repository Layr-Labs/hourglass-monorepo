package signer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

// expandPath expands ~ to the user's home directory and returns absolute path
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	return path
}

func Command() *cli.Command {
	return &cli.Command{
		Name:   "signer",
		Usage:  "Configure signer for the current context",
		Action: signerWizard,
		Subcommands: []*cli.Command{
			{
				Name:   "remove",
				Usage:  "Remove signer configuration from current context",
				Action: removeSignerConfig,
			},
		},
	}
}

func signerWizard(c *cli.Context) error {
	// If args provided, show help
	if c.Args().Present() {
		return cli.ShowCommandHelp(c, "signer")
	}

	contextName := getContextName()

	// Load current configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, ok := cfg.Contexts[contextName]
	if !ok {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	// Phase 1: Check if OperatorKeys is configured
	if ctx.OperatorKeys == nil {
		fmt.Println(titleStyle.Render("🔑 Operator Key Configuration"))
		fmt.Println("\nOperator keys are required for operator identity.")
		fmt.Println("These must be ECDSA keys configured via private key.\n")

		if err := runOperatorKeyWizard(contextName); err != nil {
			return fmt.Errorf("failed to configure operator keys: %w", err)
		}

		// Reload config after operator key configuration
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to reload config: %w", err)
		}
		ctx = cfg.Contexts[contextName]
	}

	// Phase 2: Check if SystemSignerKeys is configured
	if ctx.SystemSignerKeys == nil {
		fmt.Println(titleStyle.Render("🔐 System Signer Configuration"))
		fmt.Println("\nSystem signer keys are used for signing operations.")
		fmt.Println("You can configure either ECDSA or BN254 keys.\n")

		if err := runSystemSignerWizard(contextName); err != nil {
			return fmt.Errorf("failed to configure system signer: %w", err)
		}
	} else {
		fmt.Println(successStyle.Render("✓ Both operator and system signer keys are already configured"))
	}

	return nil
}

// wizardType represents the type of wizard being run
type wizardType int

const (
	wizardTypeOperator wizardType = iota
	wizardTypeSystem
)

// signerWizardModel is the unified wizard model for both operator and system signer configuration
type signerWizardModel struct {
	contextName string
	wizardType  wizardType
	stage       stage
	width       int
	height      int

	// UI components
	list       list.Model
	textInput  textinput.Model
	filepicker filepicker.Model

	// Common configuration
	keyType            string // "ecdsa" or "bn254" (system signer only)
	signerType         string // "private_key", "keystore", "web3signer"
	keystoreChoiceType string // "existing" or "file"
	keystoreName       string // Name of selected keystore
	keystorePath       string // Path to keystore file

	// Web3Signer configuration
	web3SignerURL            string
	web3SignerAddress        string
	web3SignerUseTLS         bool
	web3SignerCACertPath     string
	web3SignerClientCertPath string
	web3SignerClientKeyPath  string
	web3SignerPublicKey      string
	web3SignerFromAddress    string

	// State
	err       error
	completed bool
}

// Wizard stages
type stage int

const (
	stageWelcome       stage = iota
	stageSelectKeyType       // System signer only
	stageSelectSignerType
	stageKeystoreChoice
	stageKeystoreSelect
	stageKeystorePath
	stageWeb3SignerURL
	stageWeb3SignerAddress
	stageWeb3SignerTLSChoice
	stageWeb3SignerCACert
	stageWeb3SignerClientCert
	stageWeb3SignerClientKey
	stageWeb3SignerPublicKey
	stageWeb3SignerFromAddress
	stagePrivateKeyInfo
	stageConfirm
)

// Signer type item for list
type signerItem struct {
	title       string
	description string
	signerType  string
}

func (i signerItem) Title() string       { return i.title }
func (i signerItem) Description() string { return i.description }
func (i signerItem) FilterValue() string { return i.title }

// TLS choice item for list
type tlsChoiceItem struct {
	title       string
	description string
	useTLS      bool
}

func (i tlsChoiceItem) Title() string       { return i.title }
func (i tlsChoiceItem) Description() string { return i.description }
func (i tlsChoiceItem) FilterValue() string { return i.title }

// Keystore choice item for list
type keystoreChoiceItem struct {
	title       string
	description string
	choiceType  string
}

func (i keystoreChoiceItem) Title() string       { return i.title }
func (i keystoreChoiceItem) Description() string { return i.description }
func (i keystoreChoiceItem) FilterValue() string { return i.title }

// Keystore item for existing keystore selection
type keystoreItem struct {
	name string
	path string
}

func (i keystoreItem) Title() string       { return i.name }
func (i keystoreItem) Description() string { return i.path }
func (i keystoreItem) FilterValue() string { return i.name }

// Key type item for selecting ECDSA vs BN254
type keyTypeItem struct {
	title       string
	description string
	keyType     string
}

func (i keyTypeItem) Title() string       { return i.title }
func (i keyTypeItem) Description() string { return i.description }
func (i keyTypeItem) FilterValue() string { return i.title }

func newSignerWizardModel(contextName string, wType wizardType) signerWizardModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 0

	return signerWizardModel{
		contextName: contextName,
		wizardType:  wType,
		stage:       stageWelcome,
		textInput:   ti,
		list:        list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
}

// Unified wizard methods
func (m signerWizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m signerWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Only set list size if list is initialized
		if m.list.Items() != nil {
			m.list.SetSize(msg.Width-4, msg.Height-8)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "y", "yes":
			if m.stage == stageConfirm {
				m.completed = true
				return m, tea.Quit
			}

		case "n", "no":
			if m.stage == stageConfirm {
				// Go back to appropriate stage to start over
				if m.wizardType == wizardTypeSystem && m.keyType != "" {
					m.stage = stageSelectKeyType
				} else {
					m.stage = stageSelectSignerType
				}
				return m, nil
			}
		}
	}

	// Update components based on current stage
	var cmd tea.Cmd
	switch m.stage {
	case stageWelcome:
		// Welcome stage doesn't need component updates

	case stageSelectKeyType, stageSelectSignerType, stageWeb3SignerTLSChoice, stageKeystoreChoice, stageKeystoreSelect:
		m.list, cmd = m.list.Update(msg)

	case stageKeystorePath:
		m.filepicker, cmd = m.filepicker.Update(msg)
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.keystorePath = path
			m.stage = stageConfirm
		}

	case stageWeb3SignerURL, stageWeb3SignerAddress, stageWeb3SignerCACert, stageWeb3SignerClientCert, stageWeb3SignerClientKey, stageWeb3SignerPublicKey, stageWeb3SignerFromAddress:
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m signerWizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.stage {
	case stageWelcome:
		// Move to appropriate next stage
		if m.wizardType == wizardTypeSystem {
			// System signer starts with key type selection
			items := []list.Item{
				keyTypeItem{
					title:       "ECDSA Keys",
					description: "For Ethereum-compatible signing",
					keyType:     "ecdsa",
				},
				keyTypeItem{
					title:       "BN254 Keys",
					description: "For BLS signatures",
					keyType:     "bn254",
				},
			}
			l := list.New(items, list.NewDefaultDelegate(), 0, 0)
			l.Title = "Select Key Type"
			l.SetShowStatusBar(false)
			l.SetFilteringEnabled(false)
			if m.width > 0 && m.height > 0 {
				l.SetSize(m.width-4, m.height-8)
			}
			m.list = l
			m.stage = stageSelectKeyType
		} else {
			// Operator key goes straight to signer type (ECDSA only)
			m.keyType = "ecdsa"
			m.stage = stageSelectSignerType
			items := m.getSignerTypeItems()
			m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
			m.list.Title = "Select Operator Signer Type"
		}
		return m, nil

	case stageSelectKeyType:
		selected := m.list.SelectedItem().(keyTypeItem)
		m.keyType = selected.keyType

		// Setup signer type selection based on key type
		items := m.getSignerTypeItems()
		l := list.New(items, list.NewDefaultDelegate(), 0, 0)
		l.Title = "Select Signer Type"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		if m.width > 0 && m.height > 0 {
			l.SetSize(m.width-4, m.height-8)
		}
		m.list = l
		m.stage = stageSelectSignerType
		return m, nil

	case stageSelectSignerType:
		if i, ok := m.list.SelectedItem().(signerItem); ok {
			m.signerType = i.signerType
			switch i.signerType {
			case "private_key", "privatekey":
				m.stage = stagePrivateKeyInfo
			case "keystore":
				// Check if there are existing keystores
				if m.hasExistingKeystores() {
					m.stage = stageKeystoreChoice
					items := []list.Item{
						keystoreChoiceItem{"Use Existing Keystore", "Select from existing keystores in context", "existing"},
						keystoreChoiceItem{"Add New Keystore", "Add a new keystore file", "file"},
					}
					m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
					m.list.Title = "Keystore Selection"
				} else {
					// No existing keystores, go directly to file picker
					fp := filepicker.New()
					fp.CurrentDirectory, _ = os.UserHomeDir()
					fp.AllowedTypes = []string{".json", ".keystore"}
					m.filepicker = fp
					m.stage = stageKeystorePath
					return m, fp.Init()
				}
			case "web3signer":
				m.stage = stageWeb3SignerURL
				m.textInput.Placeholder = "https://web3signer.example.com:9000"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, nil

	case stageKeystoreChoice:
		if i, ok := m.list.SelectedItem().(keystoreChoiceItem); ok {
			m.keystoreChoiceType = i.choiceType
			if i.choiceType == "existing" {
				// Load existing keystores of appropriate type
				items := m.getExistingKeystores()
				if len(items) == 0 {
					// No keystores found
					m.err = fmt.Errorf("no %s keystores found in context", strings.ToUpper(m.keyType))
					return m, tea.Quit
				}
				m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
				m.list.Title = fmt.Sprintf("Select %s Keystore", strings.ToUpper(m.keyType))
				m.stage = stageKeystoreSelect
			} else {
				// File input for new keystore
				m.stage = stageKeystorePath
				m.textInput.Placeholder = "/path/to/keystore.json"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, nil

	case stageKeystoreSelect:
		if i, ok := m.list.SelectedItem().(keystoreItem); ok {
			m.keystoreName = i.name
			m.keystorePath = i.path
			m.stage = stageConfirm
		}
		return m, nil

	case stageKeystorePath:
		path := m.textInput.Value()
		if path != "" {
			m.keystorePath = expandPath(path)
			if _, err := os.Stat(m.keystorePath); err != nil {
				m.textInput.SetValue("")
				m.textInput.Placeholder = "File not found. Enter valid path:"
				return m, nil
			}
			m.stage = stageConfirm
		}
		return m, nil

	case stageWeb3SignerURL:
		url := m.textInput.Value()
		if url != "" {
			m.web3SignerURL = url
			m.stage = stageWeb3SignerAddress
			m.textInput.Placeholder = "0x..."
			m.textInput.SetValue("")
			m.textInput.Focus()
		}
		return m, nil

	case stageWeb3SignerAddress:
		address := m.textInput.Value()
		if address != "" {
			m.web3SignerAddress = address
			m.stage = stageWeb3SignerTLSChoice
			// Initialize TLS choice list
			items := []list.Item{
				tlsChoiceItem{"No TLS", "Connect without TLS/SSL", false},
				tlsChoiceItem{"Use TLS/mTLS", "Configure TLS certificates", true},
			}
			m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
			m.list.Title = "TLS Configuration"
		}
		return m, nil

	case stageWeb3SignerTLSChoice:
		if i, ok := m.list.SelectedItem().(tlsChoiceItem); ok {
			m.web3SignerUseTLS = i.useTLS
			if i.useTLS {
				m.stage = stageWeb3SignerCACert
				m.textInput.Placeholder = "/path/to/ca-cert.pem (or press enter to skip)"
				m.textInput.SetValue("")
				m.textInput.Focus()
			} else {
				// Skip to public key/from address for web3signer
				m.stage = stageWeb3SignerPublicKey
				m.textInput.Placeholder = "Public key (hex)"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, nil

	case stageWeb3SignerCACert:
		cert := m.textInput.Value()
		if cert != "" {
			m.web3SignerCACertPath = expandPath(cert)
		}
		m.stage = stageWeb3SignerClientCert
		m.textInput.Placeholder = "/path/to/client-cert.pem (or press enter to skip)"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageWeb3SignerClientCert:
		cert := m.textInput.Value()
		if cert != "" {
			m.web3SignerClientCertPath = expandPath(cert)
		}
		m.stage = stageWeb3SignerClientKey
		m.textInput.Placeholder = "/path/to/client-key.pem (or press enter to skip)"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageWeb3SignerClientKey:
		key := m.textInput.Value()
		if key != "" {
			m.web3SignerClientKeyPath = expandPath(key)
		}
		m.stage = stageWeb3SignerPublicKey
		m.textInput.Placeholder = "Public key (hex)"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageWeb3SignerPublicKey:
		publicKey := m.textInput.Value()
		if publicKey != "" {
			m.web3SignerPublicKey = publicKey
		}
		m.stage = stageWeb3SignerFromAddress
		m.textInput.Placeholder = "0x..."
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageWeb3SignerFromAddress:
		fromAddress := m.textInput.Value()
		if fromAddress != "" {
			m.web3SignerFromAddress = fromAddress
		}
		m.stage = stageConfirm
		return m, nil

	case stagePrivateKeyInfo:
		m.stage = stageConfirm
		return m, nil

	case stageConfirm:
		m.completed = true
		return m, tea.Quit
	}

	return m, nil
}

// Helper methods for the unified wizard
func (m signerWizardModel) getSignerTypeItems() []list.Item {
	if m.wizardType == wizardTypeOperator {
		return []list.Item{
			signerItem{"Private Key", "Use OPERATOR_PRIVATE_KEY environment variable", "private_key"},
			signerItem{"Keystore", "Use local encrypted keystore file (ECDSA only)", "keystore"},
			signerItem{"Web3Signer", "Use remote signing service", "web3signer"},
		}
	}

	// System signer
	if m.keyType == "ecdsa" {
		return []list.Item{
			signerItem{"Private Key", "From environment variable", "privatekey"},
			signerItem{"Keystore", "Local encrypted key file", "keystore"},
			signerItem{"Web3Signer", "Remote signing service", "web3signer"},
		}
	}

	// BN254 only supports keystore
	return []list.Item{
		signerItem{"Keystore", "Local encrypted BN254 key file", "keystore"},
	}
}

func (m signerWizardModel) hasExistingKeystores() bool {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}
	ctx, ok := cfg.Contexts[m.contextName]
	if !ok || len(ctx.Keystores) == 0 {
		return false
	}

	// Check if there are keystores of the appropriate type
	for _, ks := range ctx.Keystores {
		if m.keyType == "ecdsa" && (ks.Type == "ecdsa" || ks.Type == "keystore") {
			return true
		}
		if m.keyType == "bn254" && ks.Type == "bn254" {
			return true
		}
	}
	return false
}

func (m signerWizardModel) getExistingKeystores() []list.Item {
	var items []list.Item
	cfg, _ := config.LoadConfig()
	ctx, _ := cfg.Contexts[m.contextName]

	for _, ks := range ctx.Keystores {
		if m.keyType == "ecdsa" && (ks.Type == "ecdsa" || ks.Type == "keystore") {
			items = append(items, keystoreItem{ks.Name, ks.Path})
		} else if m.keyType == "bn254" && ks.Type == "bn254" {
			items = append(items, keystoreItem{ks.Name, ks.Path})
		}
	}
	return items
}

func (m signerWizardModel) View() string {
	// Set header based on wizard type
	var header string
	if m.wizardType == wizardTypeOperator {
		header = titleStyle.Render("🔑 Operator Key Configuration")
	} else {
		header = titleStyle.Render("🔐 System Signer Configuration")
	}
	context := helpStyle.Render(fmt.Sprintf("Context: %s", m.contextName))

	var content string
	var help string

	switch m.stage {
	case stageWelcome:
		if m.wizardType == wizardTypeOperator {
			content = fmt.Sprintf(
				"%s\n\n%s\n\n%s\n\n%s\n\n%s",
				selectedStyle.Render("Welcome to the Signer Configuration Wizard"),
				"This wizard will help you set up signing keys for your Hourglass operations.",
				"",
				"First, we'll configure your "+selectedStyle.Render("Operator Keys")+" for operator identity.",
				helpStyle.Render("Press enter to continue"),
			)
		} else {
			content = fmt.Sprintf(
				"%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s",
				selectedStyle.Render("Welcome to the System Signer Configuration"),
				"Now we'll configure your "+selectedStyle.Render("System Signer Keys")+" for signing operations.",
				"",
				"System signer keys are used for:",
				"  • Signing transactions and messages\n  • Resource Management API Authentication\n  • AVS-specific signing requirements",
				helpStyle.Render("Press enter to continue"),
			)
		}

	case stageSelectKeyType:
		content = m.list.View()
		help = helpStyle.Render("↑/↓: navigate • enter: select • q: quit")

	case stageSelectSignerType:
		content = m.list.View()
		if m.wizardType == wizardTypeOperator {
			help = "Select how you want to configure your operator signing key"
		} else {
			help = helpStyle.Render("↑/↓: navigate • enter: select • q: quit")
		}

	case stageKeystoreChoice:
		content = m.list.View()
		help = "Choose how to configure your keystore"

	case stageKeystoreSelect:
		if len(m.list.Items()) == 0 {
			content = errorStyle.Render(fmt.Sprintf("No %s keystores found in context.\n\nPlease add a keystore first or choose a different signer type.", strings.ToUpper(m.keyType)))
			help = "Press q to exit"
		} else {
			content = m.list.View()
			help = fmt.Sprintf("Select a %s keystore", strings.ToUpper(m.keyType))
		}

	case stageKeystorePath:
		if m.wizardType == wizardTypeSystem {
			content = fmt.Sprintf(
				"%s\n\n%s\n\n%s",
				"Select your keystore file:",
				m.filepicker.View(),
				helpStyle.Render("tab: toggle • enter: select • q: quit"),
			)
		} else {
			content = fmt.Sprintf(
				"Enter path to %s keystore file:\n\n%s\n\n%s",
				strings.ToUpper(m.keyType),
				m.textInput.View(),
				helpStyle.Render(fmt.Sprintf("The keystore must be a %s type keystore", strings.ToUpper(m.keyType))),
			)
		}

	case stageWeb3SignerURL:
		content = fmt.Sprintf(
			"Enter Web3Signer URL:\n\n%s",
			m.textInput.View(),
		)

	case stageWeb3SignerAddress:
		content = fmt.Sprintf(
			"Enter the Ethereum address to use for signing:\n\n%s",
			m.textInput.View(),
		)

	case stageWeb3SignerTLSChoice:
		content = m.list.View()
		help = "Configure TLS for Web3Signer connection"

	case stageWeb3SignerCACert:
		content = fmt.Sprintf(
			"Enter path to CA certificate (optional):\n\n%s",
			m.textInput.View(),
		)

	case stageWeb3SignerClientCert:
		content = fmt.Sprintf(
			"Enter path to client certificate (optional):\n\n%s",
			m.textInput.View(),
		)

	case stageWeb3SignerClientKey:
		content = fmt.Sprintf(
			"Enter path to client key (optional):\n\n%s",
			m.textInput.View(),
		)

	case stageWeb3SignerPublicKey:
		content = fmt.Sprintf(
			"Enter the public key for this signer:\n\n%s",
			m.textInput.View(),
		)

	case stageWeb3SignerFromAddress:
		content = fmt.Sprintf(
			"Enter the from address for this signer:\n\n%s",
			m.textInput.View(),
		)

	case stagePrivateKeyInfo:
		envVar := "SYSTEM_PRIVATE_KEY"
		if m.wizardType == wizardTypeOperator {
			envVar = "OPERATOR_PRIVATE_KEY"
		}
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s\n\n%s",
			"Private key configuration.",
			"",
			"You must provide the "+selectedStyle.Render(envVar)+" environment variable.",
			"This should be an ECDSA private key (with or without 0x prefix).",
			helpStyle.Render("Press enter to continue"),
		)

	case stageConfirm:
		summary := m.buildSummary()
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			"Configuration Summary:",
			summary,
			"",
			"Save this configuration? (y to confirm, n to restart)",
		)
	}

	if help != "" {
		return fmt.Sprintf("%s\n%s\n\n%s\n\n%s", header, context, content, helpStyle.Render(help))
	}
	return fmt.Sprintf("%s\n%s\n\n%s", header, context, content)
}

func (m signerWizardModel) buildSummary() string {
	var lines []string

	if m.wizardType == wizardTypeSystem && m.keyType != "" {
		lines = append(lines, fmt.Sprintf("  Key Type: %s", selectedStyle.Render(strings.ToUpper(m.keyType))))
	}

	// Get the display type for signer
	var signerTypeDisplay string
	switch m.signerType {
	case "private_key", "privatekey":
		signerTypeDisplay = "Private Key"
	case "keystore":
		signerTypeDisplay = "Keystore"
	case "web3signer":
		signerTypeDisplay = "Web3Signer"
	}
	lines = append(lines, fmt.Sprintf("  Signer Type: %s", selectedStyle.Render(signerTypeDisplay)))

	switch m.signerType {
	case "private_key", "privatekey":
		envVar := "SYSTEM_PRIVATE_KEY"
		if m.wizardType == wizardTypeOperator {
			envVar = "OPERATOR_PRIVATE_KEY"
		}
		lines = append(lines, fmt.Sprintf("  Environment Variable: %s", envVar))

	case "keystore":
		if m.keystoreName != "" {
			lines = append(lines, fmt.Sprintf("  Keystore: %s", m.keystoreName))
			lines = append(lines, fmt.Sprintf("  Path: %s", m.keystorePath))
		} else {
			lines = append(lines, fmt.Sprintf("  Path: %s", m.keystorePath))
		}
		envVar := "SYSTEM_KEYSTORE_PASSWORD"
		if m.wizardType == wizardTypeOperator {
			envVar = "OPERATOR_KEYSTORE_PASSWORD"
		}
		lines = append(lines, fmt.Sprintf("  Password Env: %s", envVar))

	case "web3signer":
		lines = append(lines, fmt.Sprintf("  URL: %s", m.web3SignerURL))
		lines = append(lines, fmt.Sprintf("  Address: %s", m.web3SignerAddress))
		lines = append(lines, fmt.Sprintf("  Public Key: %s", m.web3SignerPublicKey))
		lines = append(lines, fmt.Sprintf("  From Address: %s", m.web3SignerFromAddress))
		lines = append(lines, fmt.Sprintf("  TLS: %v", m.web3SignerUseTLS))
		if m.web3SignerUseTLS {
			if m.web3SignerCACertPath != "" {
				lines = append(lines, fmt.Sprintf("  CA Cert: %s", m.web3SignerCACertPath))
			}
			if m.web3SignerClientCertPath != "" {
				lines = append(lines, fmt.Sprintf("  Client Cert: %s", m.web3SignerClientCertPath))
			}
			if m.web3SignerClientKeyPath != "" {
				lines = append(lines, fmt.Sprintf("  Client Key: %s", m.web3SignerClientKeyPath))
			}
		}
	}

	return strings.Join(lines, "\n")
}

func saveSignerConfig(m signerWizardModel) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, ok := cfg.Contexts[m.contextName]
	if !ok {
		return fmt.Errorf("context '%s' not found", m.contextName)
	}

	// Build the ECDSA key config that's common to both wizard types
	ecdsaConfig := buildECDSAConfig(m, ctx)

	if m.wizardType == wizardTypeOperator {
		// Save operator key configuration
		ctx.OperatorKeys = ecdsaConfig
	} else {
		// Save system signer configuration
		if ctx.SystemSignerKeys == nil {
			ctx.SystemSignerKeys = &signer.SigningKeys{}
		}

		if m.keyType == "ecdsa" {
			ctx.SystemSignerKeys.ECDSA = ecdsaConfig
		} else if m.keyType == "bn254" {
			// BN254 only supports keystore
			ctx.SystemSignerKeys.BN254 = buildBN254Keystore(m, ctx)
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		return err
	}

	// Show success message
	if m.wizardType == wizardTypeOperator {
		fmt.Println(successStyle.Render("✓ Operator key configuration saved"))
	} else {
		fmt.Println(successStyle.Render("✓ System signer configuration saved"))
	}
	fmt.Println()

	// Show environment variable reminders
	showEnvVarReminders(m)

	return nil
}

func buildECDSAConfig(m signerWizardModel, ctx *config.Context) *signer.ECDSAKeyConfig {
	switch m.signerType {
	case "private_key", "privatekey":
		return &signer.ECDSAKeyConfig{
			PrivateKey: true,
		}

	case "keystore":
		var ks *signer.KeystoreReference
		if m.keystoreName != "" {
			// Using existing keystore
			for i := range ctx.Keystores {
				if ctx.Keystores[i].Name == m.keystoreName {
					ks = &ctx.Keystores[i]
					break
				}
			}
		} else {
			// New keystore
			name := fmt.Sprintf("%s-%s", getKeystorePrefix(m.wizardType), time.Now().Format("20060102-150405"))
			newKs := signer.KeystoreReference{
				Name: name,
				Path: m.keystorePath,
				Type: "ecdsa",
			}
			ctx.Keystores = append(ctx.Keystores, newKs)
			ks = &newKs
		}

		if ks != nil {
			return &signer.ECDSAKeyConfig{
				Keystore: ks,
			}
		}

	case "web3signer":
		web3Ref := signer.RemoteSignerReference{
			Name:        fmt.Sprintf("%s-web3signer-%s", getKeystorePrefix(m.wizardType), time.Now().Format("20060102-150405")),
			Url:         m.web3SignerURL,
			PublicKey:   m.web3SignerPublicKey,
			FromAddress: m.web3SignerFromAddress,
		}

		if m.web3SignerUseTLS {
			// Process certificate paths
			configDir := config.GetConfigDir()
			if m.web3SignerCACertPath != "" {
				web3Ref.CACertPath = processPath(m.web3SignerCACertPath, configDir)
			}
			if m.web3SignerClientCertPath != "" {
				web3Ref.ClientCertPath = processPath(m.web3SignerClientCertPath, configDir)
			}
			if m.web3SignerClientKeyPath != "" {
				web3Ref.ClientKeyPath = processPath(m.web3SignerClientKeyPath, configDir)
			}
		}

		return &signer.ECDSAKeyConfig{
			RemoteSignerConfig: &web3Ref,
		}
	}

	return nil
}

func buildBN254Keystore(m signerWizardModel, ctx *config.Context) *signer.KeystoreReference {
	if m.keystoreName != "" {
		// Using existing keystore
		for i := range ctx.Keystores {
			if ctx.Keystores[i].Name == m.keystoreName {
				return &ctx.Keystores[i]
			}
		}
	} else {
		// New keystore
		keystorePath := m.keystorePath
		configDir := config.GetConfigDir()
		if strings.HasPrefix(keystorePath, configDir) {
			keystorePath, _ = filepath.Rel(configDir, keystorePath)
		}

		return &signer.KeystoreReference{
			Name: m.contextName,
			Type: "bn254",
			Path: keystorePath,
		}
	}
	return nil
}

func getKeystorePrefix(wType wizardType) string {
	if wType == wizardTypeOperator {
		return "operator"
	}
	return "system"
}

func processPath(path, configDir string) string {
	// Expand tilde
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	// Convert to relative path if within config directory
	if path != "" && strings.HasPrefix(path, configDir) {
		path, _ = filepath.Rel(configDir, path)
	}
	return path
}

func showEnvVarReminders(m signerWizardModel) {
	switch m.signerType {
	case "private_key", "privatekey":
		envVar := "SYSTEM_PRIVATE_KEY"
		if m.wizardType == wizardTypeOperator {
			envVar = "OPERATOR_PRIVATE_KEY"
		}
		fmt.Println(helpStyle.Render(fmt.Sprintf("Remember to set the %s environment variable:", envVar)))
		fmt.Println(helpStyle.Render(fmt.Sprintf("  export %s=<your-private-key>", envVar)))
		fmt.Println(helpStyle.Render("  or configure it in your secrets environment file"))

	case "keystore":
		envVar := "SYSTEM_KEYSTORE_PASSWORD"
		if m.wizardType == wizardTypeOperator {
			envVar = "OPERATOR_KEYSTORE_PASSWORD"
		}
		fmt.Println(helpStyle.Render(fmt.Sprintf("Remember to set the %s environment variable:", envVar)))
		fmt.Println(helpStyle.Render(fmt.Sprintf("  export %s=<your-keystore-password>", envVar)))
		fmt.Println(helpStyle.Render("  or configure it in your secrets environment file"))

	case "web3signer":
		fmt.Println(helpStyle.Render("Web3Signer configuration saved."))
		fmt.Println(helpStyle.Render("Ensure your Web3Signer is running and accessible at: " + m.web3SignerURL))
	}
}

// runOperatorKeyWizard runs the wizard for configuring operator private key
func runOperatorKeyWizard(contextName string) error {
	p := tea.NewProgram(newSignerWizardModel(contextName, wizardTypeOperator), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running operator key wizard: %w", err)
	}

	if m, ok := result.(signerWizardModel); ok {
		if m.err != nil {
			return m.err
		}
		if m.completed {
			return saveSignerConfig(m)
		}
	}

	return nil
}

// runSystemSignerWizard runs the wizard for system signer configuration
func runSystemSignerWizard(contextName string) error {
	p := tea.NewProgram(newSignerWizardModel(contextName, wizardTypeSystem), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running system signer wizard: %w", err)
	}

	if m, ok := result.(signerWizardModel); ok {
		if m.err != nil {
			return m.err
		}
		if m.completed {
			return saveSignerConfig(m)
		}
	}

	return nil
}

func removeSignerConfig(_ *cli.Context) error {
	contextName := getContextName()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, ok := cfg.Contexts[contextName]
	if !ok {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	hasConfig := false

	// Remove both OperatorKeys and SystemSignerKeys
	if ctx.OperatorKeys != nil {
		ctx.OperatorKeys = nil
		hasConfig = true
	}

	if ctx.SystemSignerKeys != nil {
		ctx.SystemSignerKeys = nil
		hasConfig = true
	}

	if !hasConfig {
		fmt.Println("No signer configuration found for this context")
		return nil
	}

	if err := config.SaveConfig(cfg); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("✓ Signer configuration removed"))
	return nil
}

func getContextName() string {
	cfg, err := config.LoadConfig()
	if err == nil && cfg.CurrentContext != "" {
		return cfg.CurrentContext
	}
	return "default"
}
