// Package report renders an assessment (and its gaps) into exchange
// formats: a structured JSON report, SARIF 2.1.0 for tools that ingest
// static-analysis results, and a MITRE ATT&CK Navigator layer
// (docs/MASTER_PLAN.md §4.2 step 9, §10 Phase 5). Pure: no I/O.
package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/correlation"
)

type Format string

const (
	FormatJSON      Format = "json"
	FormatSARIF     Format = "sarif"
	FormatNavigator Format = "navigator"
)

func ParseFormat(s string) (Format, bool) {
	switch Format(s) {
	case FormatJSON, FormatSARIF, FormatNavigator:
		return Format(s), true
	}
	return "", false
}

type Input struct {
	AssetName   string
	Assessment  compliance.Assessment
	Gaps        []correlation.Gap
	GeneratedAt time.Time
}

// Filename is the download name for a format.
func Filename(in Input, f Format) string {
	short := in.Assessment.ID.String()[:8]
	switch f {
	case FormatSARIF:
		return "cyberditto-" + short + ".sarif"
	case FormatNavigator:
		return "cyberditto-" + short + "-attack-layer.json"
	}
	return "cyberditto-" + short + "-report.json"
}

// JSON is the full structured report.
func JSON(in Input) map[string]any {
	a := in.Assessment
	findings := make([]map[string]any, 0, len(a.Findings))
	for _, f := range a.Findings {
		findings = append(findings, map[string]any{
			"control_id": f.ControlID, "title": f.Title, "category": f.Category,
			"severity": f.Severity, "status": f.Status, "observed": f.Observed,
			"expected": f.Expected, "reason": f.Reason, "remediation": f.Remediation,
			"techniques": nonNil(f.Techniques),
		})
	}
	gaps := make([]map[string]any, 0, len(in.Gaps))
	for _, g := range in.Gaps {
		techs := make([]map[string]any, 0, len(g.Techniques))
		for _, t := range g.Techniques {
			techs = append(techs, map[string]any{"technique_id": t.TechniqueID, "evidence": t.State})
		}
		factors := make([]map[string]any, 0, len(g.Factors))
		for _, f := range g.Factors {
			factors = append(factors, map[string]any{"name": f.Name, "value": f.Value, "explanation": f.Explanation})
		}
		gaps = append(gaps, map[string]any{
			"control_id": g.ControlID, "title": g.Title, "severity": g.Severity,
			"priority": g.Priority, "factors": factors, "techniques": techs,
		})
	}
	return map[string]any{
		"report":       "cyberditto-assessment",
		"generated_at": in.GeneratedAt.UTC().Format(time.RFC3339),
		"asset":        in.AssetName,
		"assessment": map[string]any{
			"id": a.ID, "snapshot_id": a.SnapshotID, "twin_id": a.TwinID,
			"engine": a.Engine, "engine_version": a.EngineVersion,
			"score": a.Score, "created_at": a.CreatedAt,
			"counts": map[string]int{
				"passed": a.Counts.Passed, "failed": a.Counts.Failed,
				"unknown": a.Counts.Unknown, "not_applicable": a.Counts.NotApplicable,
			},
			"score_formula": "100 × (1 − Σ weight·fail / Σ weight) over pass/fail findings; high=5, medium=3, low=1",
		},
		"findings": findings,
		"gaps":     gaps,
	}
}

// sarifLevel maps severity to SARIF's result levels.
func sarifLevel(s compliance.Severity) string {
	switch s {
	case compliance.SeverityHigh:
		return "error"
	case compliance.SeverityMedium:
		return "warning"
	}
	return "note"
}

// SARIF renders failed findings as SARIF 2.1.0 results, one rule per
// control in the catalog. Unknown findings are included at level "none"
// with their reason, so a consumer can see what wasn't checked rather than
// assuming it passed.
func SARIF(in Input, catalog []compliance.Control) map[string]any {
	rules := make([]map[string]any, 0, len(catalog))
	index := map[string]int{}
	for i, c := range catalog {
		index[c.ID] = i
		rules = append(rules, map[string]any{
			"id":               c.ID,
			"name":             strings.ReplaceAll(c.Title, " ", ""),
			"shortDescription": map[string]string{"text": c.Title},
			"fullDescription":  map[string]string{"text": c.Rationale},
			"help":             map[string]string{"text": c.Remediation},
			"defaultConfiguration": map[string]string{
				"level": sarifLevel(c.Severity),
			},
			"properties": map[string]any{
				"category":   c.Category,
				"severity":   c.Severity,
				"techniques": nonNil(c.Techniques),
				"tags":       append([]string{"security", c.Category}, c.Techniques...),
			},
		})
	}

	location := []map[string]any{{
		"logicalLocations": []map[string]any{{
			"name":               in.AssetName,
			"fullyQualifiedName": "snapshot/" + in.Assessment.SnapshotID.String(),
			"kind":               "resource",
		}},
	}}

	results := make([]map[string]any, 0)
	for _, f := range in.Assessment.Findings {
		var level string
		switch f.Status {
		case compliance.StatusFail:
			level = sarifLevel(f.Severity)
		case compliance.StatusUnknown:
			level = "none"
		default:
			continue
		}
		msg := f.Title
		if f.Reason != "" {
			msg += ": " + f.Reason
		}
		r := map[string]any{
			"ruleId":    f.ControlID,
			"level":     level,
			"message":   map[string]string{"text": msg},
			"locations": location,
			"properties": map[string]any{
				"status":   f.Status,
				"observed": f.Observed,
				"expected": f.Expected,
			},
		}
		if i, ok := index[f.ControlID]; ok {
			r["ruleIndex"] = i
		}
		results = append(results, r)
	}

	return map[string]any{
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]any{{
			"tool": map[string]any{"driver": map[string]any{
				"name":            "CyberDitto",
				"semanticVersion": in.Assessment.EngineVersion,
				"informationUri":  "https://github.com/sshreyx1/CyberDitto",
				"rules":           rules,
			}},
			"automationDetails": map[string]string{"id": "cyberditto/" + in.Assessment.ID.String()},
			"results":           results,
		}},
	}
}

// Navigator renders the gaps as an ATT&CK Navigator layer: each technique
// a failed control makes easier is scored by its highest gap priority,
// with the controls and evidence state in the comment.
func Navigator(in Input) map[string]any {
	type agg struct {
		score    float64
		controls []string
		evidence map[string]bool
	}
	byTech := map[string]*agg{}
	maxScore := 0.0
	for _, g := range in.Gaps {
		for _, t := range g.Techniques {
			a := byTech[t.TechniqueID]
			if a == nil {
				a = &agg{evidence: map[string]bool{}}
				byTech[t.TechniqueID] = a
			}
			if g.Priority > a.score {
				a.score = g.Priority
			}
			a.controls = append(a.controls, g.ControlID)
			a.evidence[string(t.State)] = true
		}
		if g.Priority > maxScore {
			maxScore = g.Priority
		}
	}
	ids := make([]string, 0, len(byTech))
	for id := range byTech {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	techniques := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		a := byTech[id]
		ev := make([]string, 0, len(a.evidence))
		for e := range a.evidence {
			ev = append(ev, e)
		}
		sort.Strings(ev)
		techniques = append(techniques, map[string]any{
			"techniqueID": id,
			"score":       a.score,
			"enabled":     true,
			"comment":     fmt.Sprintf("Failed controls: %s. Evidence: %s.", strings.Join(a.controls, ", "), strings.Join(ev, ", ")),
		})
	}
	if maxScore == 0 {
		maxScore = 1
	}
	score := "n/a"
	if in.Assessment.Score != nil {
		score = fmt.Sprintf("%g", *in.Assessment.Score)
	}
	return map[string]any{
		"name":        "CyberDitto gaps: " + in.AssetName,
		"description": fmt.Sprintf("Techniques made easier by failed controls in assessment %s (posture score %s). Score = gap priority (severity × evidence × prevalence × exposure).", in.Assessment.ID, score),
		"domain":      "enterprise-attack",
		"versions":    map[string]string{"layer": "4.5", "navigator": "5.1.0"},
		"sorting":     3,
		"gradient":    map[string]any{"colors": []string{"#ffe766", "#ff6666"}, "minValue": 0, "maxValue": maxScore},
		"techniques":  techniques,
	}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
