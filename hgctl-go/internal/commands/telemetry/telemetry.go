package telemetry

import (
	"fmt"
	"os"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF"))
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "telemetry",
		Usage: "Configure telemetry settings",
		Description: `Configure telemetry settings for hgctl.

Telemetry helps improve hgctl by collecting anonymous usage data.
All data is privacy-preserving and you maintain full control.`,
		Subcommands: []*cli.Command{
			{
				Name:   "enable",
				Usage:  "Enable telemetry with full data collection",
				Action: enableTelemetry,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "anonymous",
						Usage: "Enable telemetry without operator address tracking",
					},
				},
			},
			{
				Name:   "disable",
				Usage:  "Disable telemetry completely",
				Action: disableTelemetry,
			},
			{
				Name:   "status",
				Usage:  "Show current telemetry configuration",
				Action: showStatus,
			},
			{
				Name:    "configure",
				Aliases: []string{"config"},
				Usage:   "Configure telemetry interactively",
				Action:  telemetryWizard,
			},
		},
		Action: telemetryWizard, // Default to wizard if no subcommand
	}
}

// enableTelemetry enables telemetry with optional anonymous mode
func enableTelemetry(c *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	anonymous := c.Bool("anonymous")
	enabled := true
	cfg.TelemetryEnabled = &enabled
	cfg.TelemetryAnonymous = anonymous

	// Track telemetry enablement
	telemetry.TrackEvent("telemetry_enabled", map[string]interface{}{
		"anonymous": anonymous,
	})

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if anonymous {
		fmt.Println(successStyle.Render("✓ Telemetry enabled (anonymous mode)"))
		fmt.Println(infoStyle.Render("  • Usage data will be collected"))
		fmt.Println(infoStyle.Render("  • Operator address will NOT be tracked"))
		fmt.Println(infoStyle.Render("  • All data is anonymized"))
	} else {
		fmt.Println(successStyle.Render("✓ Telemetry enabled"))
		fmt.Println(infoStyle.Render("  • Usage data will be collected"))
		fmt.Println(infoStyle.Render("  • Operator address will be included for better insights"))
		fmt.Println(infoStyle.Render("  • All data is privacy-preserving"))
	}

	return nil
}

// disableTelemetry disables telemetry completely
func disableTelemetry(c *cli.Context) error {
	// Track telemetry disablement before disabling
	telemetry.TrackEvent("telemetry_disabled", map[string]interface{}{})

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	disabled := false
	cfg.TelemetryEnabled = &disabled

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(successStyle.Render("✓ Telemetry disabled"))
	fmt.Println(helpStyle.Render("  No data will be collected"))

	return nil
}

// showStatus displays the current telemetry configuration
func showStatus(c *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	outputFormat := c.String("output")

	// Determine telemetry status
	status := "Disabled"
	mode := "N/A"

	if cfg.TelemetryEnabled != nil && *cfg.TelemetryEnabled {
		status = "Enabled"

		if cfg.TelemetryAnonymous {
			mode = "Anonymous"
		} else {
			mode = "Full"
		}
	}

	// Check for API key
	hasAPIKey := cfg.PostHogAPIKey != ""
	apiKeyStatus := "Not configured"
	if hasAPIKey {
		apiKeyStatus = "Configured"
	}

	data := map[string]interface{}{
		"status":   status,
		"mode":     mode,
		"api_key":  apiKeyStatus,
		"endpoint": "https://us.i.posthog.com",
	}

	// Add environment variable overrides if present
	envOverrides := make(map[string]string)
	if envEnabled := getEnvVar("HGCTL_TELEMETRY_ENABLED"); envEnabled != "" {
		envOverrides["HGCTL_TELEMETRY_ENABLED"] = envEnabled
	}
	if envKey := getEnvVar("HGCTL_POSTHOG_KEY"); envKey != "" {
		envOverrides["HGCTL_POSTHOG_KEY"] = "***configured***"
	}
	if envEndpoint := getEnvVar("HGCTL_POSTHOG_ENDPOINT"); envEndpoint != "" {
		envOverrides["HGCTL_POSTHOG_ENDPOINT"] = envEndpoint
	}

	if len(envOverrides) > 0 {
		data["env_overrides"] = envOverrides
	}

	// Use formatter to output the data
	formatter := output.NewFormatter(outputFormat)
	return formatter.Print(data)
}

// telemetryWizard runs an interactive wizard for configuring telemetry
func telemetryWizard(c *cli.Context) error {
	// If args provided, show help
	if c.Args().Present() {
		return cli.ShowCommandHelp(c, "telemetry")
	}

	fmt.Println(titleStyle.Render("📊 Telemetry Configuration"))
	fmt.Println()
	fmt.Println("Telemetry helps us improve hgctl by understanding how it's used.")
	fmt.Println("All data collected is privacy-preserving and you have full control.")
	fmt.Println()

	model := newTelemetryWizardModel()
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("failed to run wizard: %w", err)
	}

	// Check if user completed the wizard
	if m, ok := finalModel.(telemetryWizardModel); ok {
		if m.completed {
			return applyTelemetryConfig(m.choice)
		}
		if m.cancelled {
			fmt.Println(helpStyle.Render("Configuration cancelled"))
			return nil
		}
	}

	return nil
}

// telemetryWizardModel is the model for the telemetry configuration wizard
type telemetryWizardModel struct {
	list      list.Model
	choice    telemetryChoice
	width     int
	height    int
	completed bool
	cancelled bool
}

type telemetryChoice string

const (
	choiceEnableFull      telemetryChoice = "enable_full"
	choiceEnableAnonymous telemetryChoice = "enable_anonymous"
	choiceDisable         telemetryChoice = "disable"
)

// telemetryItem represents a telemetry configuration option
type telemetryItem struct {
	title       string
	description string
	choice      telemetryChoice
}

func (i telemetryItem) Title() string       { return i.title }
func (i telemetryItem) Description() string { return i.description }
func (i telemetryItem) FilterValue() string { return i.title }

func newTelemetryWizardModel() telemetryWizardModel {
	items := []list.Item{
		telemetryItem{
			title:       "Enable Telemetry",
			description: "Collect usage data including operator address for better insights",
			choice:      choiceEnableFull,
		},
		telemetryItem{
			title:       "Enable Anonymous Telemetry",
			description: "Collect usage data WITHOUT operator address (fully anonymous)",
			choice:      choiceEnableAnonymous,
		},
		telemetryItem{
			title:       "Disable Telemetry",
			description: "Do not collect any data",
			choice:      choiceDisable,
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Choose telemetry configuration:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return telemetryWizardModel{
		list: l,
	}
}

func (m telemetryWizardModel) Init() tea.Cmd {
	return nil
}

func (m telemetryWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter:
			if item, ok := m.list.SelectedItem().(telemetryItem); ok {
				m.choice = item.choice
				m.completed = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m telemetryWizardModel) View() string {
	if m.completed || m.cancelled {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Use ↑/↓ to navigate, Enter to select, Esc to cancel"))
	return b.String()
}

// applyTelemetryConfig applies the selected telemetry configuration
func applyTelemetryConfig(choice telemetryChoice) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch choice {
	case choiceEnableFull:
		enabled := true
		cfg.TelemetryEnabled = &enabled
		cfg.TelemetryAnonymous = false

		// Track telemetry enablement from wizard
		telemetry.TrackEvent("telemetry_enabled", map[string]interface{}{
			"anonymous": false,
			"source":    "wizard",
		})

		fmt.Println()
		fmt.Println(successStyle.Render("✓ Telemetry enabled"))
		fmt.Println(infoStyle.Render("  • Usage data will be collected"))
		fmt.Println(infoStyle.Render("  • Operator address will be included for better insights"))
		fmt.Println(infoStyle.Render("  • All data is privacy-preserving"))
		fmt.Println()
		fmt.Println(helpStyle.Render("To disable telemetry later, run: hgctl telemetry disable"))

	case choiceEnableAnonymous:
		enabled := true
		cfg.TelemetryEnabled = &enabled
		cfg.TelemetryAnonymous = true

		// Track anonymous telemetry enablement from wizard
		telemetry.TrackEvent("telemetry_enabled", map[string]interface{}{
			"anonymous": true,
			"source":    "wizard",
		})

		fmt.Println()
		fmt.Println(successStyle.Render("✓ Telemetry enabled (anonymous mode)"))
		fmt.Println(infoStyle.Render("  • Usage data will be collected"))
		fmt.Println(infoStyle.Render("  • Operator address will NOT be tracked"))
		fmt.Println(infoStyle.Render("  • All data is anonymized"))
		fmt.Println()
		fmt.Println(helpStyle.Render("To change settings later, run: hgctl telemetry configure"))

	case choiceDisable:
		// Track telemetry disablement from wizard before disabling
		telemetry.TrackEvent("telemetry_disabled", map[string]interface{}{
			"source": "wizard",
		})

		disabled := false
		cfg.TelemetryEnabled = &disabled

		fmt.Println()
		fmt.Println(successStyle.Render("✓ Telemetry disabled"))
		fmt.Println(helpStyle.Render("  No data will be collected"))
		fmt.Println()
		fmt.Println(helpStyle.Render("To enable telemetry later, run: hgctl telemetry enable"))
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// getEnvVar safely gets an environment variable
func getEnvVar(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
