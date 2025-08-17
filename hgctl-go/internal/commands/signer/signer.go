package signer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
)

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

	p := tea.NewProgram(newWizardModel(contextName), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running wizard: %w", err)
	}

	if m, ok := result.(wizardModel); ok {
		if m.err != nil {
			return m.err
		}
		if m.completed {
			return saveConfig(m)
		}
	}

	return nil
}

type wizardModel struct {
	contextName string
	stage       stage
	width       int
	height      int

	// UI components
	list       list.Model
	textInput  textinput.Model
	filepicker filepicker.Model

	// User selections
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

	// State
	err       error
	completed bool
}

// Wizard stages
type stage int

const (
	stageSelectType stage = iota
	stageKeystoreChoice
	stageKeystoreSelect
	stageKeystorePath
	stageWeb3SignerURL
	stageWeb3SignerAddress
	stageWeb3SignerTLSChoice
	stageWeb3SignerCACert
	stageWeb3SignerClientCert
	stageWeb3SignerClientKey
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

func newWizardModel(contextName string) wizardModel {
	items := []list.Item{
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

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Signer Type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return wizardModel{
		contextName: contextName,
		stage:       stageSelectType,
		list:        l,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return nil
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-8)

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
				m.stage = stageSelectType
				return m, nil
			}
		}
	}

	// Update components based on current stage
	var cmd tea.Cmd
	switch m.stage {
	case stageSelectType, stageWeb3SignerTLSChoice, stageKeystoreChoice, stageKeystoreSelect:
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

func (m wizardModel) View() string {
	header := titleStyle.Render("🔐 Signer Configuration")
	context := helpStyle.Render(fmt.Sprintf("Context: %s", m.contextName))

	var content string
	var help string

	switch m.stage {
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

	case stagePrivateKeyInfo:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			"Private Key Configuration",
			"Your private key will be read from the PRIVATE_KEY environment variable.",
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

func (m wizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.stage {
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
			m.stage = stageConfirm
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
		m.web3SignerClientKeyPath = m.textInput.Value()
		m.stage = stageConfirm

	case stageKeystoreChoice:
		selected := m.list.SelectedItem().(keystoreChoiceItem)
		m.keystoreChoiceType = selected.choiceType

		if m.keystoreChoiceType == "existing" {
			// Load existing keystores and show selection
			cfg, _ := config.LoadConfig()
			ctx, _ := cfg.Contexts[m.contextName]

			var items []list.Item
			for _, ks := range ctx.Keystores {
				// Only show ECDSA keystores for operator keys
				if ks.Type == "ecdsa" || ks.Type == "keystore" {
					items = append(items, keystoreItem{
						name: ks.Name,
						path: ks.Path,
					})
				}
			}

			if len(items) == 0 {
				// No ECDSA keystores available
				m.err = fmt.Errorf("no ECDSA keystores found in context '%s'. Please create an ECDSA keystore first or provide a keystore file", m.contextName)
				return m, tea.Quit
			}

			l := list.New(items, list.NewDefaultDelegate(), 0, 0)
			l.Title = "Select ECDSA Keystore"
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
	default:
		panic("invalid signer type provided")
	}

	return m, nil
}

func (m wizardModel) buildSummary() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("  Type: %s", selectedStyle.Render(m.signerType)))

	switch m.signerType {
	case "keystore":
		if m.keystoreName != "" {
			lines = append(lines, fmt.Sprintf("  Keystore: %s", m.keystoreName))
			lines = append(lines, fmt.Sprintf("  Path: %s", m.keystorePath))
		} else {
			lines = append(lines, fmt.Sprintf("  Path: %s", m.keystorePath))
		}
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("  Note: You must set KEYSTORE_PASSWORD environment"))
		lines = append(lines, helpStyle.Render("  variable before running commands that require signing."))

	case "web3signer":
		lines = append(lines, fmt.Sprintf("  URL: %s", m.web3SignerURL))
		lines = append(lines, fmt.Sprintf("  Address: %s", m.web3SignerAddress))
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
		lines = append(lines, "  Key source: PRIVATE_KEY environment variable")
	}

	return strings.Join(lines, "\n")
}

func saveConfig(m wizardModel) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, ok := cfg.Contexts[m.contextName]
	if !ok {
		return fmt.Errorf("context '%s' not found", m.contextName)
	}

	switch m.signerType {
	case "keystore":
		var ks *signer.KeystoreReference

		if m.keystoreName != "" {
			// User selected an existing keystore by name
			// Find the keystore in the context to validate its type
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
				return fmt.Errorf("keystore '%s' is of type '%s', but operator keys require ECDSA type",
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
				Type: "ecdsa", // Default to ECDSA for operator keys
				Path: keystorePath,
			}
		}

		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
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
		}

		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
			RemoteSignerConfig: rsr,
			Keystore:           nil,
			PrivateKey:         false,
		}

	case "privatekey":
		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
			RemoteSignerConfig: nil,
			Keystore:           nil,
			PrivateKey:         true,
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		return err
	}

	// Show success message with appropriate reminder
	fmt.Println(successStyle.Render("✓ Signer configuration saved"))

	if m.signerType == "keystore" {
		fmt.Println()
		fmt.Println(helpStyle.Render("Remember to set the KEYSTORE_PASSWORD environment variable:"))
		fmt.Println(helpStyle.Render("  specify this value in your configured secrets env file"))
		fmt.Println(helpStyle.Render("  or export KEYSTORE_PASSWORD=<your-password>"))
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

	if ctx.OperatorKeys == nil {
		fmt.Println("No signer configured for this context")
		return nil
	}

	ctx.OperatorKeys = nil

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
