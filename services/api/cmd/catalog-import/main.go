// Command catalog-import builds the pinned attack catalog
// (internal/adapters/atomic/catalog.json) from Atomic Red Team at the
// commit named in curation.json (docs/adr/0006). It is a developer tool,
// run when the curation or the pin changes — never at API runtime.
//
// For every curated test it downloads the technique's YAML at the pinned
// commit, records the file's SHA-256 in the manifest, and refuses to emit
// the catalog if any curated test fails a safety rule: not Windows, not an
// automatable executor, has prerequisites (the lab has no internet to
// fetch them), references the atomics source folder, or matches the
// destructive denylist. Updating the pin means editing curation.json and
// re-running this, then updating the pinned constants in the adapter —
// a reviewed change, never automatic.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"gopkg.in/yaml.v3"
)

type curation struct {
	Source     string               `json:"source"`
	Repository string               `json:"repository"`
	Commit     string               `json:"commit"`
	CommitDate string               `json:"commit_date"`
	License    string               `json:"license"`
	Tests      []curatedTest        `json:"tests"`
	Scenarios  []emulation.Scenario `json:"scenarios"`
}

type curatedTest struct {
	GUID      string             `json:"guid"`
	Technique string             `json:"technique"`
	Tier      emulation.RiskTier `json:"tier"`
	Note      string             `json:"note"`
}

// Catalog is the emitted file's shape; internal/adapters/atomic reads it.
type Catalog struct {
	Source     string               `json:"source"`
	Repository string               `json:"repository"`
	Commit     string               `json:"commit"`
	CommitDate string               `json:"commit_date"`
	License    string               `json:"license"`
	Files      map[string]string    `json:"files"` // path at commit -> sha256
	Tests      []emulation.Test     `json:"tests"`
	Scenarios  []emulation.Scenario `json:"scenarios"`
}

type atomicFile struct {
	AttackTechnique string       `yaml:"attack_technique"`
	DisplayName     string       `yaml:"display_name"`
	AtomicTests     []atomicTest `yaml:"atomic_tests"`
}

type atomicTest struct {
	Name               string                 `yaml:"name"`
	GUID               string                 `yaml:"auto_generated_guid"`
	Description        string                 `yaml:"description"`
	SupportedPlatforms []string               `yaml:"supported_platforms"`
	InputArguments     map[string]atomicInput `yaml:"input_arguments"`
	Dependencies       []map[string]any       `yaml:"dependencies"`
	Executor           struct {
		Name              string `yaml:"name"`
		ElevationRequired bool   `yaml:"elevation_required"`
		Command           string `yaml:"command"`
		CleanupCommand    string `yaml:"cleanup_command"`
	} `yaml:"executor"`
}

type atomicInput struct {
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Default     any    `yaml:"default"`
}

var (
	commitPattern    = regexp.MustCompile(`^[0-9a-f]{40}$`)
	techniquePattern = regexp.MustCompile(`^T\d{4}(\.\d{3})?$`)
)

func main() {
	curationPath := flag.String("curation", "internal/adapters/atomic/curation.json", "curation file")
	outPath := flag.String("out", "internal/adapters/atomic/catalog.json", "catalog output")
	flag.Parse()

	if err := run(*curationPath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, "catalog-import:", err)
		os.Exit(1)
	}
}

func run(curationPath, outPath string) error {
	raw, err := os.ReadFile(curationPath)
	if err != nil {
		return err
	}
	var cur curation
	if err := json.Unmarshal(raw, &cur); err != nil {
		return fmt.Errorf("parse curation: %w", err)
	}
	if !commitPattern.MatchString(cur.Commit) {
		return fmt.Errorf("curation commit %q is not a full 40-character SHA (a branch name or short SHA is not a pin)", cur.Commit)
	}

	byTechnique := map[string][]curatedTest{}
	for _, ct := range cur.Tests {
		if !techniquePattern.MatchString(ct.Technique) {
			return fmt.Errorf("curated test %s: bad technique %q", ct.GUID, ct.Technique)
		}
		byTechnique[ct.Technique] = append(byTechnique[ct.Technique], ct)
	}

	out := Catalog{
		Source: cur.Source, Repository: cur.Repository, Commit: cur.Commit,
		CommitDate: cur.CommitDate, License: cur.License,
		Files: map[string]string{}, Scenarios: cur.Scenarios,
	}
	client := &http.Client{Timeout: 30 * time.Second}

	techniques := make([]string, 0, len(byTechnique))
	for t := range byTechnique {
		techniques = append(techniques, t)
	}
	sort.Strings(techniques)

	for _, technique := range techniques {
		path := fmt.Sprintf("atomics/%s/%s.yaml", technique, technique)
		body, err := fetch(client, cur.Repository, cur.Commit, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		out.Files[path] = hex.EncodeToString(sum[:])

		var file atomicFile
		if err := yaml.Unmarshal(body, &file); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		index := map[string]atomicTest{}
		for _, at := range file.AtomicTests {
			index[at.GUID] = at
		}
		for _, ct := range byTechnique[technique] {
			at, ok := index[ct.GUID]
			if !ok {
				return fmt.Errorf("%s: test %s not found at commit %s", path, ct.GUID, cur.Commit)
			}
			test, err := convert(file, at, ct, path)
			if err != nil {
				return fmt.Errorf("%s test %s (%q): %w", path, ct.GUID, at.Name, err)
			}
			out.Tests = append(out.Tests, test)
		}
	}

	sort.Slice(out.Tests, func(i, j int) bool {
		if out.Tests[i].TechniqueID != out.Tests[j].TechniqueID {
			return out.Tests[i].TechniqueID < out.Tests[j].TechniqueID
		}
		return out.Tests[i].GUID < out.Tests[j].GUID
	})
	known := map[string]bool{}
	for _, t := range out.Tests {
		known[t.GUID] = true
	}
	for _, s := range out.Scenarios {
		for _, g := range s.TestGUIDs {
			if !known[g] {
				return fmt.Errorf("scenario %s references %s, which is not a curated test", s.ID, g)
			}
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	fmt.Printf("wrote %s: %d tests from %d files at %s\n", outPath, len(out.Tests), len(out.Files), cur.Commit)
	fmt.Printf("catalog sha256: %s\n", hex.EncodeToString(sum[:]))
	fmt.Println("update PinnedCommit and CatalogSHA256 in internal/adapters/atomic/catalog.go to accept this build")
	return nil
}

func fetch(client *http.Client, repo, commit, path string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repo, commit, path)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: HTTP %d", path, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

func convert(file atomicFile, at atomicTest, ct curatedTest, path string) (emulation.Test, error) {
	if ct.Tier != emulation.TierBenign && ct.Tier != emulation.TierIntrusive {
		return emulation.Test{}, fmt.Errorf("tier %q is not schedulable; destructive tests are never imported", ct.Tier)
	}
	windows := false
	for _, p := range at.SupportedPlatforms {
		windows = windows || p == "windows"
	}
	if !windows {
		return emulation.Test{}, fmt.Errorf("not a Windows test")
	}
	if at.Executor.Name != "powershell" && at.Executor.Name != "command_prompt" {
		return emulation.Test{}, fmt.Errorf("executor %q is not automatable", at.Executor.Name)
	}
	if len(at.Dependencies) > 0 {
		return emulation.Test{}, fmt.Errorf("has prerequisites; installing them would need network access the lab doesn't have")
	}
	for _, text := range []string{at.Executor.Command, at.Executor.CleanupCommand} {
		if strings.Contains(text, "PathToAtomicsFolder") {
			return emulation.Test{}, fmt.Errorf("needs files from the atomics source folder, which isn't shipped to twins yet")
		}
	}

	inputs := map[string]emulation.Input{}
	for name, in := range at.InputArguments {
		inputs[name] = emulation.Input{Description: strings.TrimSpace(in.Description), Type: in.Type, Default: fmt.Sprint(in.Default)}
	}
	t := emulation.Test{
		GUID: at.GUID, TechniqueID: file.AttackTechnique, TechniqueName: file.DisplayName,
		Name: at.Name, Description: strings.TrimSpace(at.Description),
		Executor: at.Executor.Name, ElevationRequired: at.Executor.ElevationRequired,
		Command: at.Executor.Command, Cleanup: at.Executor.CleanupCommand, Inputs: inputs,
		RiskTier: ct.Tier, Origin: emulation.OriginCatalog, SourcePath: path, CurationNote: ct.Note,
	}
	if t.TechniqueID != ct.Technique {
		return emulation.Test{}, fmt.Errorf("file technique %s != curated %s", t.TechniqueID, ct.Technique)
	}
	if denied, why := emulation.Denied(t); denied {
		return emulation.Test{}, fmt.Errorf("denylisted: %s", why)
	}
	return t, nil
}
