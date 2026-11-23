// Command collector is CyberDitto's read-only host collector
// (docs/MASTER_PLAN.md §5.10, replacing v1's scripts/config_collector.ps1).
// It makes no network calls of its own — the API's snapshot upload
// endpoint is a separate, explicit step (docs/PROGRESS.md: not built yet)
// — and never collects file contents, documents, browser data,
// credentials, tokens, event-log contents, clipboard data, or screenshots.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sshreyx1/cyberditto/services/collector/internal/collect"
	"github.com/sshreyx1/cyberditto/services/collector/internal/pseudonymise"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("collector", flag.ContinueOnError)
	fs.SetOutput(stderr)
	consent := fs.Bool("consent", false, "required to actually collect; without it, only --dry-run is allowed")
	dryRun := fs.Bool("dry-run", false, "print the manifest of what would be collected and exit without collecting anything")
	output := fs.String("output", "", "write the snapshot JSON to this path instead of stdout")
	pseudonymiseFlag := fs.Bool("pseudonymise", false, "replace the hostname and non-built-in account names with stable local placeholders before output")
	mapPath := fs.String("pseudonymise-map", "collector-pseudonymise-map.json", "where the local, reversible placeholder map is stored (0600; never upload this file)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	collectors := collect.All()
	manifest := collect.Manifest(collectors)

	if *dryRun {
		printManifest(stdout, manifest)
		return 0
	}

	if !*consent {
		fmt.Fprintln(stderr, "collector: refusing to run without --consent. Here is exactly what would be collected:")
		printManifest(stderr, manifest)
		return 1
	}

	snap := collect.Run(context.Background(), collectors, time.Now)

	if *pseudonymiseFlag {
		mapper, err := pseudonymise.LoadMapper(*mapPath)
		if err != nil {
			fmt.Fprintf(stderr, "collector: %v\n", err)
			return 1
		}
		if err := pseudonymiseSnapshot(&snap, mapper); err != nil {
			fmt.Fprintf(stderr, "collector: %v\n", err)
			return 1
		}
		if err := mapper.Save(*mapPath); err != nil {
			fmt.Fprintf(stderr, "collector: %v\n", err)
			return 1
		}
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "collector: encode snapshot: %v\n", err)
		return 1
	}
	digest := sha256.Sum256(data)

	if *output == "" {
		if _, err := stdout.Write(append(data, '\n')); err != nil {
			fmt.Fprintf(stderr, "collector: write snapshot: %v\n", err)
			return 1
		}
	} else {
		if err := os.WriteFile(*output, data, 0o644); err != nil {
			fmt.Fprintf(stderr, "collector: write snapshot: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "wrote %s\n", *output)
	}
	fmt.Fprintf(stderr, "sha256: %s\n", hex.EncodeToString(digest[:]))

	return 0
}

func printManifest(w io.Writer, manifest []collect.CategoryInfo) {
	fmt.Fprintln(w, "This collector will read the following, and nothing else:")
	for _, c := range manifest {
		fmt.Fprintf(w, "  - %-18s %s\n", c.ID, c.Description)
	}
	fmt.Fprintln(w, "It never reads file contents, documents, browser data, credentials, tokens, event-log contents, the clipboard, or screenshots, and makes no network calls.")
}
