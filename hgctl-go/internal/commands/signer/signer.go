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

	if m, ok := result.(wizardModel); ok && m.completed {
		return saveConfig(m)
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
	signerType        string
	keystorePath      string
	web3SignerURL     string
	web3SignerAddress string

	// State
	err       error
	completed bool
}

// Wizard stages
type stage int

const (
	stageSelectType stage = iota
	stageKeystorePath
	stageWeb3SignerURL
	stageWeb3SignerAddress
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
	case stageSelectType:
		m.list, cmd = m.list.Update(msg)

	case stageKeystorePath:
		m.filepicker, cmd = m.filepicker.Update(msg)
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.keystorePath = path
			m.stage = stageConfirm
		}

	case stageWeb3SignerURL, stageWeb3SignerAddress:
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
			"Enter signing address (0x...):",
			m.textInput.View(),
			helpStyle.Render("enter: continue • ctrl+c: quit"),
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
		lines = append(lines, fmt.Sprintf("  Path: %s", m.keystorePath))

	case "web3signer":
		lines = append(lines, fmt.Sprintf("  URL: %s", m.web3SignerURL))
		lines = append(lines, fmt.Sprintf("  Address: %s", m.web3SignerAddress))

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
		keystorePath := m.keystorePath
		configDir := config.GetConfigDir()
		if strings.HasPrefix(keystorePath, configDir) {
			keystorePath, _ = filepath.Rel(configDir, keystorePath)
		}

		ks := &signer.KeystoreReference{
			Name: m.contextName,
			Type: m.signerType,
			Path: keystorePath,
		}

		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
			UseRemoteSigner:    false,
			RemoteSignerConfig: nil,
			Keystore:           ks,
			PrivateKey:         false,
		}

	case "web3signer":
		rsr := &signer.RemoteSignerReference{
			Name:           m.contextName,
			ConfigPath:     "",
			CACertPath:     "",
			ClientCertPath: "",
			ClientKeyPath:  "",
		}

		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
			UseRemoteSigner:    true,
			RemoteSignerConfig: rsr,
			Keystore:           nil,
			PrivateKey:         false,
		}

	case "privatekey":
		ctx.OperatorKeys = &signer.ECDSAKeyConfig{
			UseRemoteSigner:    false,
			RemoteSignerConfig: nil,
			Keystore:           nil,
			PrivateKey:         true,
		}
	}

	return config.SaveConfig(cfg)
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
