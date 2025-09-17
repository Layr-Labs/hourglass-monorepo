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
		},
		Action: telemetryWizard,
	}
}

func enableTelemetry(c *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	anonymous := c.Bool("anonymous")
	enabled := true
	cfg.TelemetryEnabled = &enabled
	cfg.TelemetryAnonymous = &anonymous

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if metrics, err := telemetry.MetricsFromContext(c.Context); err == nil {
		metrics.AddMetricWithDimensions("TelemetryConfigChanged", 1, map[string]string{
			"action":    "enabled",
			"anonymous": fmt.Sprintf("%v", anonymous),
		})
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

func disableTelemetry(c *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	disabled := false
	cfg.TelemetryEnabled = &disabled

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if metrics, err := telemetry.MetricsFromContext(c.Context); err == nil {
		metrics.AddMetricWithDimensions("TelemetryConfigChanged", 1, map[string]string{
			"action": "disabled",
		})
	}

	fmt.Println(successStyle.Render("✓ Telemetry disabled"))
	fmt.Println(helpStyle.Render("  No data will be collected"))

	return nil
}

func showStatus(c *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	outputFormat := c.String("output")

	status := "Disabled"
	mode := "N/A"

	if cfg.TelemetryEnabled != nil && *cfg.TelemetryEnabled {
		status = "Enabled"

		if cfg.TelemetryAnonymous != nil && *cfg.TelemetryAnonymous {
			mode = "Anonymous"
		} else {
			mode = "Full"
		}
	}

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

	formatter := output.NewFormatter(outputFormat)
	return formatter.Print(data)
}

func telemetryWizard(c *cli.Context) error {
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

	if m, ok := finalModel.(telemetryWizardModel); ok {
		if m.completed {
			return applyTelemetryConfig(c, m.choice)
		}
		if m.cancelled {
			fmt.Println(helpStyle.Render("Configuration cancelled"))
			return nil
		}
	}

	return nil
}

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

func applyTelemetryConfig(c *cli.Context, choice telemetryChoice) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch choice {
	case choiceEnableFull:
		enabled := true
		anonymous := false
		cfg.TelemetryEnabled = &enabled
		cfg.TelemetryAnonymous = &anonymous

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if metrics, err := telemetry.MetricsFromContext(c.Context); err == nil {
			metrics.AddMetricWithDimensions("TelemetryConfigChanged", 1, map[string]string{
				"action":    "enabled",
				"anonymous": "false",
				"source":    "wizard",
			})
		}

		fmt.Println()
		fmt.Println(successStyle.Render("✓ Telemetry enabled"))
		fmt.Println(infoStyle.Render("  • Usage data will be collected"))
		fmt.Println(infoStyle.Render("  • Operator address will be included for better insights"))
		fmt.Println(infoStyle.Render("  • All data is privacy-preserving"))
		fmt.Println()
		fmt.Println(helpStyle.Render("To disable telemetry later, run: hgctl telemetry disable"))

	case choiceEnableAnonymous:
		enabled := true
		anonymous := true
		cfg.TelemetryEnabled = &enabled
		cfg.TelemetryAnonymous = &anonymous

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if metrics, err := telemetry.MetricsFromContext(c.Context); err == nil {
			metrics.AddMetricWithDimensions("TelemetryConfigChanged", 1, map[string]string{
				"action":    "enabled",
				"anonymous": "true",
				"source":    "wizard",
			})
		}

		fmt.Println()
		fmt.Println(successStyle.Render("✓ Telemetry enabled (anonymous mode)"))
		fmt.Println(infoStyle.Render("  • Usage data will be collected"))
		fmt.Println(infoStyle.Render("  • Operator address will NOT be tracked"))
		fmt.Println(infoStyle.Render("  • All data is anonymized"))
		fmt.Println()
		fmt.Println(helpStyle.Render("To change settings later, run: hgctl telemetry configure"))

	case choiceDisable:
		disabled := false
		prevEnabled := cfg.TelemetryEnabled
		if prevEnabled != nil && *prevEnabled {
			if metrics, err := telemetry.MetricsFromContext(c.Context); err == nil {
				metrics.AddMetricWithDimensions("TelemetryConfigChanged", 1, map[string]string{
					"action": "disabled",
					"source": "wizard",
				})
				client, ok := telemetry.ClientFromContext(c.Context)
				if ok {
					for _, metric := range metrics.Metrics {
						_ = client.AddMetric(c.Context, metric)
					}
				}
				_ = client.Close()
			}

		}

		cfg.TelemetryEnabled = &disabled

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println()
		fmt.Println(successStyle.Render("✓ Telemetry disabled"))
		fmt.Println(helpStyle.Render("  No data will be collected"))
		fmt.Println()
		fmt.Println(helpStyle.Render("To enable telemetry later, run: hgctl telemetry enable"))
	}

	return nil
}

func getEnvVar(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
