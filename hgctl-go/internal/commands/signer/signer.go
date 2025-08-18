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

// operatorKeyWizardModel handles the Phase 1 operator key configuration
type operatorKeyWizardModel struct {
	contextName string
	stage       stage

	// Signer configuration
	signerType           string // "private_key", "keystore", "web3signer"
	keystoreChoiceType   string // "existing" or "file"
	keystoreName         string // Name of selected keystore
	keystorePath         string // Path to keystore file
	web3SignerURL        string
	web3SignerAddress    string
	web3SignerTLS        bool
	web3SignerCACert     string
	web3SignerClientCert string
	web3SignerClientKey  string
	fromAddress          string
	publicKey            string

	// UI components
	list       list.Model
	textInput  textinput.Model
	filepicker filepicker.Model

	completed bool
	width     int
	height    int
}

// systemSignerWizardModel is the full wizard for system signer configuration
type systemSignerWizardModel struct {
	contextName string
	stage       stage
	width       int
	height      int

	// UI components
	list       list.Model
	textInput  textinput.Model
	filepicker filepicker.Model

	// User selections
	keyType                  string // "ecdsa" or "bn254"
	signerType               string
	keystoreChoiceType       string // "existing" or "file"
	keystoreName             string
	keystorePath             string
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
	// Operator key stages
	stageOperatorKeyWelcome stage = iota
	stageOperatorSelectType
	stageOperatorKeystoreChoice
	stageOperatorKeystoreSelect
	stageOperatorKeystorePath
	stageOperatorWeb3SignerURL
	stageOperatorWeb3SignerAddress
	stageOperatorWeb3SignerTLSChoice
	stageOperatorWeb3SignerCACert
	stageOperatorWeb3SignerClientCert
	stageOperatorWeb3SignerClientKey
	stageOperatorPrivateKeyInfo
	stageOperatorWeb3SignerFromAddress
	stageOperatorWeb3SignerPublicKey
	stageOperatorConfirm

	// System signer stages
	stageSystemSignerWelcome
	stageSelectKeyType
	stageSelectType
	stageKeystoreChoice
	stageKeystoreSelect
	stageKeystorePath
	stageWeb3SignerURL
	stageWeb3SignerAddress
	stageWeb3SignerTLSChoice
	stageWeb3SignerCACert
	stageWeb3SignerClientCert
	stageWeb3SignerClientKey
	stageWeb3SignerFromAddress
	stageWeb3SignerPublicKey
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

func newOperatorKeyWizardModel(contextName string) operatorKeyWizardModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 0

	return operatorKeyWizardModel{
		contextName: contextName,
		stage:       stageOperatorKeyWelcome,
		textInput:   ti,
		list:        list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
}

func newSystemSignerWizardModel(contextName string) systemSignerWizardModel {
	return systemSignerWizardModel{
		contextName: contextName,
		stage:       stageSystemSignerWelcome,
	}
}

// Operator key wizard methods
func (m operatorKeyWizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m operatorKeyWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.stage == stageOperatorConfirm {
				m.completed = true
				return m, tea.Quit
			}

		case "n", "no":
			if m.stage == stageOperatorConfirm {
				// Go back to signer type selection to start over
				m.stage = stageOperatorSelectType
				return m, nil
			}
		}
	}

	// Update components based on current stage
	var cmd tea.Cmd
	switch m.stage {
	case stageOperatorKeyWelcome:
		// Welcome stage doesn't need component updates

	case stageOperatorSelectType, stageOperatorWeb3SignerTLSChoice, stageOperatorKeystoreChoice, stageOperatorKeystoreSelect:
		m.list, cmd = m.list.Update(msg)

	case stageOperatorKeystorePath:
		m.filepicker, cmd = m.filepicker.Update(msg)
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.keystorePath = path
			m.stage = stageOperatorConfirm
		}

	case stageOperatorWeb3SignerURL, stageOperatorWeb3SignerAddress, stageOperatorWeb3SignerCACert, stageOperatorWeb3SignerClientCert, stageOperatorWeb3SignerClientKey:
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m operatorKeyWizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.stage {
	case stageOperatorKeyWelcome:
		// Move to signer type selection
		m.stage = stageOperatorSelectType
		// Initialize list for signer type selection
		items := []list.Item{
			signerItem{"Private Key", "Use OPERATOR_PRIVATE_KEY environment variable", "private_key"},
			signerItem{"Keystore", "Use local encrypted keystore file (ECDSA only)", "keystore"},
			signerItem{"Web3Signer", "Use remote signing service", "web3signer"},
		}
		m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
		m.list.Title = "Select Operator Signer Type"
		return m, nil

	case stageOperatorSelectType:
		if i, ok := m.list.SelectedItem().(signerItem); ok {
			m.signerType = i.signerType
			switch i.signerType {
			case "private_key":
				m.stage = stageOperatorPrivateKeyInfo
			case "keystore":
				m.stage = stageOperatorKeystoreChoice
				// Initialize list for keystore choice
				items := []list.Item{
					keystoreChoiceItem{"Use Existing Keystore", "Select from existing ECDSA keystores in context", "existing"},
					keystoreChoiceItem{"Add New Keystore", "Add a new ECDSA keystore file", "file"},
				}
				m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
				m.list.Title = "Keystore Selection"
			case "web3signer":
				m.stage = stageOperatorWeb3SignerURL
				m.textInput.Placeholder = "https://web3signer.example.com:9000"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, nil

	case stageOperatorKeystoreChoice:
		if i, ok := m.list.SelectedItem().(keystoreChoiceItem); ok {
			m.keystoreChoiceType = i.choiceType
			if i.choiceType == "existing" {
				// Load existing ECDSA keystores
				ctx, _ := config.GetCurrentContext()
				var items []list.Item
				for _, ks := range ctx.Keystores {
					if ks.Type == "ecdsa" {
						items = append(items, keystoreItem{ks.Name, ks.Path})
					}
				}
				if len(items) == 0 {
					// No ECDSA keystores found
					return m, tea.Quit
				}
				m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
				m.list.Title = "Select ECDSA Keystore"
				m.stage = stageOperatorKeystoreSelect
			} else {
				// File input for new keystore
				m.stage = stageOperatorKeystorePath
				m.textInput.Placeholder = "/path/to/keystore.json"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, nil

	case stageOperatorKeystoreSelect:
		if i, ok := m.list.SelectedItem().(keystoreItem); ok {
			m.keystoreName = i.name
			m.keystorePath = i.path
			m.stage = stageOperatorConfirm
		}
		return m, nil

	case stageOperatorKeystorePath:
		path := m.textInput.Value()
		if path != "" {
			m.keystorePath = expandPath(path)
			// Validate it's an ECDSA keystore
			if _, err := os.Stat(m.keystorePath); err != nil {
				m.textInput.SetValue("")
				m.textInput.Placeholder = "File not found. Enter valid path:"
				return m, nil
			}
			m.stage = stageOperatorConfirm
		}
		return m, nil

	case stageOperatorWeb3SignerURL:
		url := m.textInput.Value()
		if url != "" {
			m.web3SignerURL = url
			m.stage = stageOperatorWeb3SignerAddress
			m.textInput.Placeholder = "0x..."
			m.textInput.SetValue("")
			m.textInput.Focus()
		}
		return m, nil

	case stageOperatorWeb3SignerAddress:
		address := m.textInput.Value()
		if address != "" {
			m.web3SignerAddress = address
			m.stage = stageOperatorWeb3SignerTLSChoice
			// Initialize TLS choice list
			items := []list.Item{
				tlsChoiceItem{"No TLS", "Connect without TLS/SSL", false},
				tlsChoiceItem{"Use TLS/mTLS", "Configure TLS certificates", true},
			}
			m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
			m.list.Title = "TLS Configuration"
		}
		return m, nil

	case stageOperatorWeb3SignerTLSChoice:
		if i, ok := m.list.SelectedItem().(tlsChoiceItem); ok {
			m.web3SignerTLS = i.useTLS
			if i.useTLS {
				m.stage = stageOperatorWeb3SignerCACert
				m.textInput.Placeholder = "/path/to/ca-cert.pem (or press enter to skip)"
				m.textInput.SetValue("")
				m.textInput.Focus()
			} else {
				m.stage = stageOperatorConfirm
			}
		}
		return m, nil

	case stageOperatorWeb3SignerCACert:
		cert := m.textInput.Value()
		if cert != "" {
			m.web3SignerCACert = expandPath(cert)
		}
		m.stage = stageOperatorWeb3SignerClientCert
		m.textInput.Placeholder = "/path/to/client-cert.pem (or press enter to skip)"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageOperatorWeb3SignerClientCert:
		cert := m.textInput.Value()
		if cert != "" {
			m.web3SignerClientCert = expandPath(cert)
		}
		m.stage = stageOperatorWeb3SignerClientKey
		m.textInput.Placeholder = "/path/to/client-key.pem (or press enter to skip)"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageOperatorWeb3SignerClientKey:
		key := m.textInput.Value()
		if key != "" {
			m.web3SignerClientKey = expandPath(key)
		}
		m.stage = stageOperatorWeb3SignerPublicKey
		return m, nil

	case stageOperatorWeb3SignerPublicKey:
		publicKey := m.textInput.Value()
		if publicKey != "" {
			m.publicKey = expandPath(publicKey)
		}
		m.stage = stageOperatorWeb3SignerFromAddress
		m.textInput.Placeholder = "Remote Signer Public Key"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageOperatorWeb3SignerFromAddress:
		fromAddress := m.textInput.Value()
		if fromAddress != "" {
			m.fromAddress = expandPath(fromAddress)
		}
		m.stage = stageOperatorConfirm
		m.textInput.Placeholder = "Remote Signer From Address"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case stageOperatorPrivateKeyInfo:
		m.stage = stageOperatorConfirm
		return m, nil

	case stageOperatorConfirm:
		m.completed = true
		return m, tea.Quit
	}

	return m, nil
}

func (m operatorKeyWizardModel) View() string {
	header := titleStyle.Render("🔑 Operator Key Configuration")
	context := helpStyle.Render(fmt.Sprintf("Context: %s", m.contextName))

	var content string
	var help string

	switch m.stage {
	case stageOperatorKeyWelcome:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s\n\n%s",
			selectedStyle.Render("Welcome to the Signer Configuration Wizard"),
			"This wizard will help you set up signing keys for your Hourglass operations.",
			"",
			"First, we'll configure your "+selectedStyle.Render("Operator Keys")+" for operator identity.",
			helpStyle.Render("Press enter to continue"),
		)

	case stageOperatorSelectType:
		content = m.list.View()
		help = "Select how you want to configure your operator signing key"

	case stageOperatorKeystoreChoice:
		content = m.list.View()
		help = "Choose how to configure your keystore"

	case stageOperatorKeystoreSelect:
		if len(m.list.Items()) == 0 {
			content = errorStyle.Render("No ECDSA keystores found in context.\n\nPlease add a keystore first or choose a different signer type.")
			help = "Press q to exit"
		} else {
			content = m.list.View()
			help = "Select an ECDSA keystore"
		}

	case stageOperatorKeystorePath:
		content = fmt.Sprintf(
			"Enter path to ECDSA keystore file:\n\n%s\n\n%s",
			m.textInput.View(),
			helpStyle.Render("The keystore must be an ECDSA type keystore"),
		)

	case stageOperatorWeb3SignerURL:
		content = fmt.Sprintf(
			"Enter Web3Signer URL:\n\n%s",
			m.textInput.View(),
		)

	case stageOperatorWeb3SignerAddress:
		content = fmt.Sprintf(
			"Enter the Ethereum address to use for signing:\n\n%s",
			m.textInput.View(),
		)

	case stageOperatorWeb3SignerTLSChoice:
		content = m.list.View()
		help = "Configure TLS for Web3Signer connection"

	case stageOperatorWeb3SignerCACert:
		content = fmt.Sprintf(
			"Enter path to CA certificate (optional):\n\n%s",
			m.textInput.View(),
		)

	case stageOperatorWeb3SignerClientCert:
		content = fmt.Sprintf(
			"Enter path to client certificate (optional):\n\n%s",
			m.textInput.View(),
		)

	case stageOperatorWeb3SignerClientKey:
		content = fmt.Sprintf(
			"Enter path to client key (optional):\n\n%s",
			m.textInput.View(),
		)

	case stageOperatorWeb3SignerPublicKey:
		content = fmt.Sprintf(
			"Enter the public key for this signer:\n\n%s",
			m.textInput.View(),
		)

	case stageOperatorWeb3SignerFromAddress:
		content = fmt.Sprintf(
			"Enter the from address for this signer:\n\n%s",
			m.textInput.View(),
		)

	case stageOperatorPrivateKeyInfo:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s\n\n%s",
			"Operator keys will use private key mode.",
			"",
			"You must provide the "+selectedStyle.Render("OPERATOR_PRIVATE_KEY")+" environment variable.",
			"This should be an ECDSA private key (with or without 0x prefix).",
			helpStyle.Render("Press enter to continue"),
		)

	case stageOperatorConfirm:
		var summary string
		switch m.signerType {
		case "private_key":
			summary = fmt.Sprintf(
				"  Type: %s\n  Environment Variable: OPERATOR_PRIVATE_KEY",
				selectedStyle.Render("Private Key (ECDSA)"),
			)
		case "keystore":
			if m.keystoreName != "" {
				summary = fmt.Sprintf(
					"  Type: %s\n  Keystore: %s\n  Path: %s\n  Password Env: OPERATOR_KEYSTORE_PASSWORD",
					selectedStyle.Render("Keystore (ECDSA)"),
					m.keystoreName,
					m.keystorePath,
				)
			} else {
				summary = fmt.Sprintf(
					"  Type: %s\n  Path: %s\n  Password Env: OPERATOR_KEYSTORE_PASSWORD",
					selectedStyle.Render("Keystore (ECDSA)"),
					m.keystorePath,
				)
			}
		case "web3signer":
			summary = fmt.Sprintf(
				"  Type: %s\n  URL: %s\n  Address: %s\n  Public Key: %s\n  From Address: %s\n  TLS: %v",
				selectedStyle.Render("Web3Signer"),
				m.web3SignerURL,
				m.web3SignerAddress,
				m.publicKey,
				m.fromAddress,
				m.web3SignerTLS,
			)
			if m.web3SignerTLS {
				if m.web3SignerCACert != "" {
					summary += fmt.Sprintf("\n  CA Cert: %s", m.web3SignerCACert)
				}
				if m.web3SignerClientCert != "" {
					summary += fmt.Sprintf("\n  Client Cert: %s", m.web3SignerClientCert)
				}
				if m.web3SignerClientKey != "" {
					summary += fmt.Sprintf("\n  Client Key: %s", m.web3SignerClientKey)
				}
			}
		}

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

// System signer wizard methods
func (m systemSignerWizardModel) Init() tea.Cmd {
	return nil
}

func (m systemSignerWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				// Go back to key type selection to start over
				m.stage = stageSelectKeyType
				return m, nil
			}
		}
	}

	// Update components based on current stage
	var cmd tea.Cmd
	switch m.stage {
	case stageSystemSignerWelcome:
		// Welcome stage doesn't need component updates

	case stageSelectKeyType, stageSelectType, stageWeb3SignerTLSChoice, stageKeystoreChoice, stageKeystoreSelect:
		m.list, cmd = m.list.Update(msg)

	case stageKeystorePath:
		m.filepicker, cmd = m.filepicker.Update(msg)
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.keystorePath = path
			m.stage = stageConfirm
		}

	case stageWeb3SignerURL, stageWeb3SignerAddress, stageWeb3SignerCACert, stageWeb3SignerClientCert, stageWeb3SignerClientKey:
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m systemSignerWizardModel) View() string {
	header := titleStyle.Render("🔐 System Signer Configuration")
	context := helpStyle.Render(fmt.Sprintf("Context: %s", m.contextName))

	var content string
	var help string

	switch m.stage {
	case stageSystemSignerWelcome:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s",
			selectedStyle.Render("Welcome to the System Signer Configuration"),
			"Now we'll configure your "+selectedStyle.Render("System Signer Keys")+" for signing operations.",
			"",
			"System signer keys are used for:",
			"  • Signing transactions and messages\n  • Management API Authentication\n  • AVS-specific signing requirements",
			helpStyle.Render("Press enter to continue"),
		)

	case stageSelectKeyType:
		content = m.list.View()
		help = helpStyle.Render("↑/↓: navigate • enter: select • q: quit")

	case stageSelectType:
		content = m.list.View()
		help = helpStyle.Render("↑/↓: navigate • enter: select • q: quit")

	case stageKeystoreChoice:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"How would you like to configure the keystore?",
			m.list.View(),
			helpStyle.Render("↑/↓: navigate • enter: select • q: quit"),
		)

	case stageKeystoreSelect:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Select a keystore:",
			m.list.View(),
			helpStyle.Render("↑/↓: navigate • enter: select • q: quit"),
		)

	case stageKeystorePath:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Select your keystore file:",
			m.filepicker.View(),
			helpStyle.Render("tab: toggle • enter: select • q: quit"),
		)

	case stageWeb3SignerURL:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Enter Web3Signer URL:",
			m.textInput.View(),
			helpStyle.Render("enter: continue • ctrl+c: quit"),
		)

	case stageWeb3SignerAddress:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Enter remote signing hex address (0x...):",
			m.textInput.View(),
			helpStyle.Render("enter: continue • ctrl+c: quit"),
		)

	case stageWeb3SignerTLSChoice:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Configure TLS certificates for Web3Signer?",
			m.list.View(),
			helpStyle.Render("↑/↓: navigate • enter: select • q: quit"),
		)

	case stageWeb3SignerCACert:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			"Enter CA certificate path (optional):",
			"Leave empty to skip. Used to verify the Web3Signer server's certificate.",
			m.textInput.View(),
			helpStyle.Render("enter: continue (empty to skip) • ctrl+c: quit"),
		)

	case stageWeb3SignerClientCert:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			"Enter client certificate path (optional):",
			"Leave empty to skip. Required for mutual TLS authentication.",
			m.textInput.View(),
			helpStyle.Render("enter: continue (empty to skip) • ctrl+c: quit"),
		)

	case stageWeb3SignerClientKey:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			"Enter client key path (optional):",
			"Leave empty to skip. Private key for the client certificate.",
			m.textInput.View(),
			helpStyle.Render("enter: continue (empty to skip) • ctrl+c: quit"),
		)

	case stageWeb3SignerPublicKey:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Enter the public key for this signer:",
			m.textInput.View(),
			helpStyle.Render("enter: continue • ctrl+c: quit"),
		)

	case stageWeb3SignerFromAddress:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Enter the from address for this signer:",
			m.textInput.View(),
			helpStyle.Render("enter: continue • ctrl+c: quit"),
		)

	case stagePrivateKeyInfo:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			"Private Key Configuration",
			"Your private key will be read from the SYSTEM_PRIVATE_KEY environment variable.",
			"Make sure to set it before running commands that require signing.",
			helpStyle.Render("Press enter to continue"),
		)

	case stageConfirm:
		summary := m.buildSummary()
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"Configuration Summary:",
			summary,
			"Save this configuration? (y/n)",
		)
	}

	if help != "" && m.stage == stageSelectType {
		return fmt.Sprintf("%s\n%s\n\n%s\n\n%s", header, context, content, help)
	}

	return fmt.Sprintf("%s\n%s\n\n%s", header, context, content)
}

func (m systemSignerWizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.stage {
	case stageSystemSignerWelcome:
		// Move to key type selection
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
		return m, nil

	case stageSelectKeyType:
		selected := m.list.SelectedItem().(keyTypeItem)
		m.keyType = selected.keyType

		// Setup signer type selection based on key type
		var items []list.Item
		if m.keyType == "ecdsa" {
			// ECDSA supports all signer types
			items = []list.Item{
				signerItem{
					title:       "Keystore",
					description: "Local encrypted key file",
					signerType:  "keystore",
				},
				signerItem{
					title:       "Web3Signer",
					description: "Remote signing service",
					signerType:  "web3signer",
				},
				signerItem{
					title:       "Private Key",
					description: "From environment variable",
					signerType:  "privatekey",
				},
			}
		} else {
			// BN254 only supports keystore
			items = []list.Item{
				signerItem{
					title:       "Keystore",
					description: "Local encrypted BN254 key file",
					signerType:  "keystore",
				},
			}
		}

		l := list.New(items, list.NewDefaultDelegate(), 0, 0)
		l.Title = "Select Signer Type"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		if m.width > 0 && m.height > 0 {
			l.SetSize(m.width-4, m.height-8)
		}
		m.list = l
		m.stage = stageSelectType
		return m, nil

	case stageSelectType:
		selected := m.list.SelectedItem().(signerItem)
		m.signerType = selected.signerType

		switch m.signerType {
		case "keystore":
			// Check if there are existing keystores
			cfg, err := config.LoadConfig()
			if err == nil {
				ctx, ok := cfg.Contexts[m.contextName]
				if ok && len(ctx.Keystores) > 0 {
					// Show choice between existing and new keystore
					items := []list.Item{
						keystoreChoiceItem{
							title:       "Use existing keystore",
							description: "Select from available keystores",
							choiceType:  "existing",
						},
						keystoreChoiceItem{
							title:       "Provide keystore file",
							description: "Browse for a keystore file",
							choiceType:  "file",
						},
					}
					l := list.New(items, list.NewDefaultDelegate(), 0, 0)
					l.Title = "Keystore Configuration"
					l.SetShowStatusBar(false)
					l.SetFilteringEnabled(false)
					if m.width > 0 && m.height > 0 {
						l.SetSize(m.width-4, m.height-8)
					}
					m.list = l
					m.stage = stageKeystoreChoice
					return m, nil
				}
			}
			// No existing keystores, go directly to file picker
			fp := filepicker.New()
			fp.CurrentDirectory, _ = os.UserHomeDir()
			fp.AllowedTypes = []string{".json", ".keystore"}
			m.filepicker = fp
			m.stage = stageKeystorePath
			return m, fp.Init()

		case "web3signer":
			m.textInput = textinput.New()
			m.textInput.Placeholder = "https://localhost:9000"
			m.textInput.Focus()
			m.textInput.CharLimit = 200
			m.stage = stageWeb3SignerURL
			return m, textinput.Blink

		case "privatekey":
			m.stage = stagePrivateKeyInfo
		}

	case stageWeb3SignerURL:
		m.web3SignerURL = m.textInput.Value()
		m.textInput = textinput.New()
		m.textInput.Placeholder = "0x..."
		m.textInput.Focus()
		m.textInput.CharLimit = 42
		m.stage = stageWeb3SignerAddress
		return m, textinput.Blink

	case stageWeb3SignerAddress:
		m.web3SignerAddress = m.textInput.Value()
		// Ask about TLS configuration
		items := []list.Item{
			tlsChoiceItem{
				title:       "Configure TLS certificates",
				description: "Set up CA cert and client certificates",
				useTLS:      true,
			},
			tlsChoiceItem{
				title:       "Skip TLS configuration",
				description: "Continue without certificates",
				useTLS:      false,
			},
		}
		l := list.New(items, list.NewDefaultDelegate(), 0, 0)
		l.Title = "TLS Configuration"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		// Set the size based on current window dimensions
		if m.width > 0 && m.height > 0 {
			l.SetSize(m.width-4, m.height-8)
		}
		m.list = l
		m.stage = stageWeb3SignerTLSChoice

	case stageWeb3SignerTLSChoice:
		selected := m.list.SelectedItem().(tlsChoiceItem)
		m.web3SignerUseTLS = selected.useTLS
		if m.web3SignerUseTLS {
			// Start TLS certificate configuration
			m.textInput = textinput.New()
			m.textInput.Placeholder = "/path/to/ca-cert.pem (leave empty to skip)"
			m.textInput.Focus()
			m.textInput.CharLimit = 500
			m.stage = stageWeb3SignerCACert
			return m, textinput.Blink
		} else {
			// Move to public key input
			m.stage = stageWeb3SignerPublicKey
			m.textInput = textinput.New()
			m.textInput.Placeholder = "Public key (hex)"
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, textinput.Blink
		}

	case stageWeb3SignerCACert:
		m.web3SignerCACertPath = m.textInput.Value()
		m.textInput = textinput.New()
		m.textInput.Placeholder = "/path/to/client-cert.pem (leave empty to skip)"
		m.textInput.Focus()
		m.textInput.CharLimit = 500
		m.stage = stageWeb3SignerClientCert
		return m, textinput.Blink

	case stageWeb3SignerClientCert:
		m.web3SignerClientCertPath = m.textInput.Value()
		m.textInput = textinput.New()
		m.textInput.Placeholder = "/path/to/client-key.pem (leave empty to skip)"
		m.textInput.Focus()
		m.textInput.CharLimit = 500
		m.stage = stageWeb3SignerClientKey
		return m, textinput.Blink

	case stageWeb3SignerClientKey:
		key := m.textInput.Value()
		if key != "" {
			m.web3SignerClientKeyPath = expandPath(key)
		}
		// Move to public key input
		m.stage = stageWeb3SignerPublicKey
		m.textInput = textinput.New()
		m.textInput.Placeholder = "Public key (hex)"
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink

	case stageWeb3SignerPublicKey:
		publicKey := m.textInput.Value()
		if publicKey != "" {
			m.web3SignerPublicKey = publicKey
		}
		// Move to from address input
		m.stage = stageWeb3SignerFromAddress
		m.textInput = textinput.New()
		m.textInput.Placeholder = "0x..."
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink

	case stageWeb3SignerFromAddress:
		fromAddress := m.textInput.Value()
		if fromAddress != "" {
			m.web3SignerFromAddress = fromAddress
		}
		m.stage = stageConfirm
		return m, nil

	case stageKeystoreChoice:
		selected := m.list.SelectedItem().(keystoreChoiceItem)
		m.keystoreChoiceType = selected.choiceType

		if m.keystoreChoiceType == "existing" {
			// Load existing keystores and show selection
			cfg, _ := config.LoadConfig()
			ctx, _ := cfg.Contexts[m.contextName]

			var items []list.Item
			for _, ks := range ctx.Keystores {
				// Filter keystores based on selected key type for system signer
				if m.keyType == "ecdsa" {
					if ks.Type == "ecdsa" || ks.Type == "keystore" {
						items = append(items, keystoreItem{
							name: ks.Name,
							path: ks.Path,
						})
					}
				} else if m.keyType == "bn254" {
					if ks.Type == "bn254" {
						items = append(items, keystoreItem{
							name: ks.Name,
							path: ks.Path,
						})
					}
				}
			}

			if len(items) == 0 {
				// No keystores of the selected type available
				keyTypeDisplay := strings.ToUpper(m.keyType)
				m.err = fmt.Errorf("no %s keystores found in context '%s'. Please create a %s keystore first or provide a keystore file",
					keyTypeDisplay, m.contextName, keyTypeDisplay)
				return m, tea.Quit
			}

			l := list.New(items, list.NewDefaultDelegate(), 0, 0)
			l.Title = fmt.Sprintf("Select %s Keystore", strings.ToUpper(m.keyType))
			l.SetShowStatusBar(false)
			l.SetFilteringEnabled(true)
			if m.width > 0 && m.height > 0 {
				l.SetSize(m.width-4, m.height-8)
			}
			m.list = l
			m.stage = stageKeystoreSelect
		} else {
			// Go to file picker
			fp := filepicker.New()
			fp.CurrentDirectory, _ = os.UserHomeDir()
			fp.AllowedTypes = []string{".json", ".keystore"}
			m.filepicker = fp
			m.stage = stageKeystorePath
			return m, fp.Init()
		}

	case stageKeystoreSelect:
		selected := m.list.SelectedItem().(keystoreItem)
		m.keystoreName = selected.name
		m.keystorePath = selected.path
		m.stage = stageConfirm

	case stagePrivateKeyInfo:
		m.stage = stageConfirm

	case stageConfirm:
		// User confirmed the configuration
		m.completed = true
		return m, tea.Quit

	default:
		// This should not happen unless there's a bug in stage management
		return m, nil
	}

	return m, nil
}

func (m systemSignerWizardModel) buildSummary() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("  Key Type: %s", selectedStyle.Render(strings.ToUpper(m.keyType))))
	lines = append(lines, fmt.Sprintf("  Signer Type: %s", selectedStyle.Render(m.signerType)))

	switch m.signerType {
	case "keystore":
		if m.keystoreName != "" {
			lines = append(lines, fmt.Sprintf("  Keystore: %s", m.keystoreName))
			lines = append(lines, fmt.Sprintf("  Path: %s", m.keystorePath))
		} else {
			lines = append(lines, fmt.Sprintf("  Path: %s", m.keystorePath))
		}
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("  Note: You must set SYSTEM_KEYSTORE_PASSWORD environment"))
		lines = append(lines, helpStyle.Render("  variable before running commands that require signing."))

	case "web3signer":
		lines = append(lines, fmt.Sprintf("  URL: %s", m.web3SignerURL))
		lines = append(lines, fmt.Sprintf("  Address: %s", m.web3SignerAddress))
		lines = append(lines, fmt.Sprintf("  Public Key: %s", m.web3SignerPublicKey))
		lines = append(lines, fmt.Sprintf("  From Address: %s", m.web3SignerFromAddress))
		if m.web3SignerCACertPath != "" {
			lines = append(lines, fmt.Sprintf("  CA Cert: %s", m.web3SignerCACertPath))
		}
		if m.web3SignerClientCertPath != "" {
			lines = append(lines, fmt.Sprintf("  Client Cert: %s", m.web3SignerClientCertPath))
		}
		if m.web3SignerClientKeyPath != "" {
			lines = append(lines, fmt.Sprintf("  Client Key: %s", m.web3SignerClientKeyPath))
		}

	case "privatekey":
		lines = append(lines, "  Key source: SYSTEM_PRIVATE_KEY environment variable")
	}

	return strings.Join(lines, "\n")
}

func saveOperatorKey(m operatorKeyWizardModel) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, ok := cfg.Contexts[m.contextName]
	if !ok {
		return fmt.Errorf("context '%s' not found", m.contextName)
	}

	// Configure operator key based on selected type
	switch m.signerType {
	case "private_key":
		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
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
			name := fmt.Sprintf("operator-%s", time.Now().Format("20060102-150405"))
			newKs := signer.KeystoreReference{
				Name: name,
				Path: m.keystorePath,
				Type: "ecdsa",
			}
			ctx.Keystores = append(ctx.Keystores, newKs)
			ks = &newKs
		}

		if ks != nil {
			ctx.OperatorKeys = &signer.ECDSAKeyConfig{
				Keystore: ks,
			}
		}

	case "web3signer":
		web3Ref := signer.RemoteSignerReference{
			Name:        fmt.Sprintf("operator-web3signer-%s", time.Now().Format("20060102-150405")),
			Url:         m.web3SignerURL,
			PublicKey:   m.publicKey,
			FromAddress: m.fromAddress,
		}

		if m.web3SignerTLS {
			if m.web3SignerCACert != "" {
				web3Ref.CACertPath = m.web3SignerCACert
			}
			if m.web3SignerClientCert != "" {
				web3Ref.ClientCertPath = m.web3SignerClientCert
			}
			if m.web3SignerClientKey != "" {
				web3Ref.ClientKeyPath = m.web3SignerClientKey
			}
		}

		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
			RemoteSignerConfig: &web3Ref,
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		return err
	}

	// Show success message with appropriate reminder
	fmt.Println(successStyle.Render("✓ Operator key configuration saved"))
	fmt.Println()

	switch m.signerType {
	case "private_key":
		fmt.Println(helpStyle.Render("Remember to set the OPERATOR_PRIVATE_KEY environment variable:"))
		fmt.Println(helpStyle.Render("  export OPERATOR_PRIVATE_KEY=<your-private-key>"))
	case "keystore":
		fmt.Println(helpStyle.Render("Remember to set the OPERATOR_KEYSTORE_PASSWORD environment variable:"))
		fmt.Println(helpStyle.Render("  export OPERATOR_KEYSTORE_PASSWORD=<your-keystore-password>"))
	case "web3signer":
		fmt.Println(helpStyle.Render("Web3Signer configuration saved."))
		fmt.Println(helpStyle.Render("Ensure your Web3Signer is running and accessible at: " + m.web3SignerURL))
	}
	fmt.Println(helpStyle.Render("  or configure it in your secrets environment file"))

	return nil
}

func saveSystemSigner(m systemSignerWizardModel) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, ok := cfg.Contexts[m.contextName]
	if !ok {
		return fmt.Errorf("context '%s' not found", m.contextName)
	}

	// Initialize SystemSignerKeys if nil
	if ctx.SystemSignerKeys == nil {
		ctx.SystemSignerKeys = &signer.SigningKeys{}
	}

	if m.keyType == "ecdsa" {
		// Handle ECDSA configuration
		switch m.signerType {
		case "keystore":
			var ks *signer.KeystoreReference

			if m.keystoreName != "" {
				// User selected an existing keystore by name
				var foundKeystore *signer.KeystoreReference
				for i := range ctx.Keystores {
					if ctx.Keystores[i].Name == m.keystoreName {
						foundKeystore = &ctx.Keystores[i]
						break
					}
				}

				if foundKeystore == nil {
					return fmt.Errorf("keystore '%s' not found in context", m.keystoreName)
				}

				// Validate that it's an ECDSA keystore
				if foundKeystore.Type != "ecdsa" && foundKeystore.Type != "keystore" {
					return fmt.Errorf("keystore '%s' is of type '%s', but ECDSA key type was selected",
						m.keystoreName, foundKeystore.Type)
				}

				ks = foundKeystore
			} else {
				// User provided a file path
				keystorePath := m.keystorePath
				configDir := config.GetConfigDir()
				if strings.HasPrefix(keystorePath, configDir) {
					keystorePath, _ = filepath.Rel(configDir, keystorePath)
				}

				ks = &signer.KeystoreReference{
					Name: m.contextName,
					Type: "ecdsa",
					Path: keystorePath,
				}
			}

			ctx.SystemSignerKeys.ECDSA = &signer.ECDSAKeyConfig{
				RemoteSignerConfig: nil,
				Keystore:           ks,
				PrivateKey:         false,
			}

		case "web3signer":
			// Handle path expansion and relative paths for certificates
			caCertPath := m.web3SignerCACertPath
			clientCertPath := m.web3SignerClientCertPath
			clientKeyPath := m.web3SignerClientKeyPath

			// Expand tilde in paths if present
			if strings.HasPrefix(caCertPath, "~/") {
				if home, err := os.UserHomeDir(); err == nil {
					caCertPath = filepath.Join(home, caCertPath[2:])
				}
			}
			if strings.HasPrefix(clientCertPath, "~/") {
				if home, err := os.UserHomeDir(); err == nil {
					clientCertPath = filepath.Join(home, clientCertPath[2:])
				}
			}
			if strings.HasPrefix(clientKeyPath, "~/") {
				if home, err := os.UserHomeDir(); err == nil {
					clientKeyPath = filepath.Join(home, clientKeyPath[2:])
				}
			}

			// Convert to relative paths if within config directory
			configDir := config.GetConfigDir()
			if caCertPath != "" && strings.HasPrefix(caCertPath, configDir) {
				caCertPath, _ = filepath.Rel(configDir, caCertPath)
			}
			if clientCertPath != "" && strings.HasPrefix(clientCertPath, configDir) {
				clientCertPath, _ = filepath.Rel(configDir, clientCertPath)
			}
			if clientKeyPath != "" && strings.HasPrefix(clientKeyPath, configDir) {
				clientKeyPath, _ = filepath.Rel(configDir, clientKeyPath)
			}

			rsr := &signer.RemoteSignerReference{
				Name:           m.contextName,
				Url:            m.web3SignerURL,
				ConfigPath:     "",
				CACertPath:     caCertPath,
				ClientCertPath: clientCertPath,
				ClientKeyPath:  clientKeyPath,
				PublicKey:      m.web3SignerPublicKey,
				FromAddress:    m.web3SignerFromAddress,
			}

			ctx.SystemSignerKeys.ECDSA = &signer.ECDSAKeyConfig{
				RemoteSignerConfig: rsr,
				Keystore:           nil,
				PrivateKey:         false,
			}

		case "privatekey":
			ctx.SystemSignerKeys.ECDSA = &signer.ECDSAKeyConfig{
				RemoteSignerConfig: nil,
				Keystore:           nil,
				PrivateKey:         true,
			}
		}
	} else if m.keyType == "bn254" {
		// BN254 only supports keystore
		var ks *signer.KeystoreReference

		if m.keystoreName != "" {
			// User selected an existing keystore by name
			var foundKeystore *signer.KeystoreReference
			for i := range ctx.Keystores {
				if ctx.Keystores[i].Name == m.keystoreName {
					foundKeystore = &ctx.Keystores[i]
					break
				}
			}

			if foundKeystore == nil {
				return fmt.Errorf("keystore '%s' not found in context", m.keystoreName)
			}

			// Validate that it's a BN254 keystore
			if foundKeystore.Type != "bn254" {
				return fmt.Errorf("keystore '%s' is of type '%s', but BN254 key type was selected",
					m.keystoreName, foundKeystore.Type)
			}

			ks = foundKeystore
		} else {
			// User provided a file path
			keystorePath := m.keystorePath
			configDir := config.GetConfigDir()
			if strings.HasPrefix(keystorePath, configDir) {
				keystorePath, _ = filepath.Rel(configDir, keystorePath)
			}

			ks = &signer.KeystoreReference{
				Name: m.contextName,
				Type: "bn254",
				Path: keystorePath,
			}
		}

		ctx.SystemSignerKeys.BN254 = ks
	}

	if err := config.SaveConfig(cfg); err != nil {
		return err
	}

	// Show success message with appropriate reminder
	fmt.Println(successStyle.Render("✓ System signer configuration saved"))

	if m.signerType == "keystore" {
		fmt.Println()
		fmt.Println(helpStyle.Render("Remember to set the SYSTEM_KEYSTORE_PASSWORD environment variable:"))
		fmt.Println(helpStyle.Render("  specify this value in your configured secrets env file"))
		fmt.Println(helpStyle.Render("  or export SYSTEM_KEYSTORE_PASSWORD=<your-password>"))
	} else if m.signerType == "privatekey" {
		fmt.Println()
		fmt.Println(helpStyle.Render("Remember to set the SYSTEM_PRIVATE_KEY environment variable:"))
		fmt.Println(helpStyle.Render("  specify this value in your configured secrets env file"))
		fmt.Println(helpStyle.Render("  or export SYSTEM_PRIVATE_KEY=<your-private-key>"))
	}

	return nil
}

// runOperatorKeyWizard runs a simple wizard for configuring operator private key
func runOperatorKeyWizard(contextName string) error {
	p := tea.NewProgram(newOperatorKeyWizardModel(contextName), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running operator key wizard: %w", err)
	}

	if m, ok := result.(operatorKeyWizardModel); ok && m.completed {
		return saveOperatorKey(m)
	}

	return nil
}

// runSystemSignerWizard runs the full wizard for system signer configuration
func runSystemSignerWizard(contextName string) error {
	p := tea.NewProgram(newSystemSignerWizardModel(contextName), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running system signer wizard: %w", err)
	}

	if m, ok := result.(systemSignerWizardModel); ok {
		if m.err != nil {
			return m.err
		}
		if m.completed {
			return saveSystemSigner(m)
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
