package templates

import (
	_ "embed"
)

//go:embed aggregator-config.yaml
var aggregatorConfigTemplate string

//go:embed executor-config.yaml
var executorConfigTemplate string
