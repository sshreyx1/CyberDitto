package collect

import "context"

// serviceScript always wraps the pipeline in @(...) so ConvertTo-Json
// reliably produces a JSON array even when exactly one (or zero) services
// match — without it, PowerShell's ConvertTo-Json emits a bare object for
// a single-item pipeline, which would silently break JSON decoding into a
// slice only on machines with exactly one service in some edge case.
const serviceScript = `ConvertTo-Json -Compress -InputObject @(Get-Service | Select-Object Name, DisplayName, Status, StartType)`

// Service is one Windows service's config + current state
// (docs/MASTER_PLAN.md §5.10: "services (config + state)") — names and
// status only, never anything about what the service does at runtime.
type Service struct {
	Name        string `json:"Name"`
	DisplayName string `json:"DisplayName"`
	Status      int    `json:"Status"`    // Get-Service's ServiceControllerStatus enum value
	StartType   int    `json:"StartType"` // ServiceStartMode enum value
}

// ServicesCollector lists every installed Windows service's name, display
// name, run state, and start type.
type ServicesCollector struct {
	Runner PowerShellRunner
}

func (ServicesCollector) ID() string { return "services" }

func (ServicesCollector) Description() string {
	return "Every installed Windows service's name, display name, current run state, and start type (via Get-Service; no service's own configuration or data is read)."
}

func (c ServicesCollector) Collect(ctx context.Context) (any, error) {
	var services []Service
	if err := runJSON(ctx, c.Runner, serviceScript, &services); err != nil {
		return nil, err
	}
	return services, nil
}
