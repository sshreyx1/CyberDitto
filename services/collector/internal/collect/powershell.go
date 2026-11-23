package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// PowerShellRunner executes a fixed PowerShell script and returns its
// stdout. Abstracted so collectors built on it are testable without a real
// Windows host (docs/MASTER_PLAN.md §5.10: golden-file tests).
type PowerShellRunner interface {
	Run(ctx context.Context, script string) ([]byte, error)
}

// SystemPowerShell shells out to the real powershell.exe. script is always
// a fixed, compile-time constant passed as its own argv element (never
// built from user input or snapshot data) — this is what "no shell-string
// building" (docs/MASTER_PLAN.md §6.3) means in practice: the risk it
// guards against is untrusted data reaching a shell's parser, and nothing
// here ever puts untrusted data anywhere near this argument.
type SystemPowerShell struct{}

func (SystemPowerShell) Run(ctx context.Context, script string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("powershell: %w", err)
	}
	return out, nil
}

// runJSON runs script and decodes its stdout as JSON into v. PowerShell's
// ConvertTo-Json emits a single object (not wrapped in an array) when a
// pipeline produces exactly one result, so callers that always expect a
// list pass `| ConvertTo-Json -Compress` input that is itself guaranteed
// to be an array (e.g. by wrapping with `@(...)` in the script), rather
// than handling both shapes here.
func runJSON(ctx context.Context, runner PowerShellRunner, script string, v any) error {
	out, err := runner.Run(ctx, script)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(out, v); err != nil {
		return fmt.Errorf("parse powershell output: %w", err)
	}
	return nil
}
