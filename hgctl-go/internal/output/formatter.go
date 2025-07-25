package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

type ReleaseWithSpec struct {
	Release     *client.Release `json:"release"`
	RuntimeSpec *runtime.Spec   `json:"runtimeSpec"`
}

type Formatter struct {
	format string
}

func NewFormatter(format string) *Formatter {
	if format == "" {
		format = "table"
	}
	return &Formatter{format: format}
}

func (f *Formatter) PrintReleaseWithSpec(data *ReleaseWithSpec) error {
	switch f.format {
	case "json":
		return f.printJSON(data)
	case "yaml":
		return f.printYAML(data)
	case "table":
		return f.printReleaseTable(data)
	default:
		return fmt.Errorf("unsupported output format: %s", f.format)
	}
}

func (f *Formatter) PrintReleases(releases []*client.Release) error {
	switch f.format {
	case "json":
		return f.printJSON(releases)
	case "yaml":
		return f.printYAML(releases)
	case "table":
		return f.printReleasesTable(releases)
	default:
		return fmt.Errorf("unsupported output format: %s", f.format)
	}
}

func (f *Formatter) printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) printYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

func (f *Formatter) PrintJSON(data interface{}) error {
	return f.printJSON(data)
}

func (f *Formatter) PrintYAML(data interface{}) error {
	return f.printYAML(data)
}

func (f *Formatter) printReleaseTable(data *ReleaseWithSpec) error {
	// Print release information
	fmt.Println("\n=== RELEASE INFORMATION ===")
	fmt.Printf("Release ID: %s\n", data.Release.ID)
	fmt.Printf("Upgrade By: %s\n", time.Unix(int64(data.Release.UpgradeByTime), 0).Format(time.RFC3339))

	// Print operator set releases
	fmt.Println("\n=== OPERATOR SET RELEASES ===")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"OPERATOR SET", "DIGEST", "REGISTRY"})

	for opSet, release := range data.Release.OperatorSetReleases {
		digest := release.Digest
		if len(digest) > 12 {
			digest = digest[:12] + "..."
		}
		table.Append([]string{opSet, digest, release.Registry})
	}
	table.Render()

	// Print runtime spec if available
	if data.RuntimeSpec != nil {
		fmt.Println("\n=== RUNTIME SPECIFICATION ===")
		fmt.Printf("API Version: %s\n", data.RuntimeSpec.APIVersion)
		fmt.Printf("Kind: %s\n", data.RuntimeSpec.Kind)
		fmt.Printf("Name: %s\n", data.RuntimeSpec.Name)
		fmt.Printf("Version: %s\n", data.RuntimeSpec.Version)

		if len(data.RuntimeSpec.Spec) > 0 {
			fmt.Println("\n=== COMPONENTS ===")
			compTable := tablewriter.NewWriter(os.Stdout)
			compTable.SetHeader([]string{"COMPONENT", "REGISTRY", "DIGEST", "ENV VARS"})

			for name, comp := range data.RuntimeSpec.Spec {
				digest := comp.Digest
				if len(digest) > 12 {
					digest = digest[:12] + "..."
				}
				envCount := fmt.Sprintf("%d", len(comp.Env))
				compTable.Append([]string{name, comp.Registry, digest, envCount})
			}
			compTable.Render()
		}
	}

	return nil
}

func (f *Formatter) printReleasesTable(releases []*client.Release) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"RELEASE ID", "UPGRADE BY", "OPERATOR SETS", "ARTIFACTS"})

	for _, release := range releases {
		upgradeBy := time.Unix(int64(release.UpgradeByTime), 0).Format("01/02/2006, 3:04:05 PM")

		var opSets string
		var artifacts string

		for opSet, rel := range release.OperatorSetReleases {
			if opSets != "" {
				opSets += ", "
			}
			opSets += fmt.Sprintf("Set %s", opSet)

			digest := rel.Digest
			if len(digest) > 12 {
				digest = digest[:12] + "..."
			}
			if artifacts != "" {
				artifacts += "\n"
			}
			artifacts += fmt.Sprintf("%s @ %s", digest, rel.Registry)
		}

		table.Append([]string{release.ID, upgradeBy, opSets, artifacts})
	}

	table.Render()
	return nil
}
