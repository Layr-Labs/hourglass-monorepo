package executorConfig

const (
	EnvPrefix = "WORKER_"
)

type PerformerImage struct {
	Repository string
	Tag        string
}

type AvsPerformerConfig struct {
	Image       PerformerImage
	ProcessType string
	AvsAddress  string
}

type ExecutorConfig struct {
	AvsPerformers []*AvsPerformerConfig
}
