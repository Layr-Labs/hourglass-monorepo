module github.com/Layr-Labs/hourglass-monorepo/ponos-performer/go

go 1.23.6

require (
	github.com/Layr-Labs/hourglass-monorepo/ponos v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/Layr-Labs/hourglass-monorepo/ponos => ../../ponos
