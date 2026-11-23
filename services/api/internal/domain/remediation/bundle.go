package remediation

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
)

type Meta struct {
	AssessmentID  string
	SnapshotID    string
	EngineVersion string
	GeneratedAt   time.Time
}

type ManualStep struct {
	ControlID   string
	Title       string
	Remediation string
}

type Bundle struct {
	Filename  string
	Script    string
	Automated []Template
	Manual    []ManualStep
}

// psQuote renders s as a PowerShell single-quoted literal. Only reviewed
// constants and IDs reach the script, but quoting is still applied to every
// string so a stray apostrophe in a title can't break (or alter) it.
func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// Render builds the bundle for an assessment's failed controls: an
// automated, reversible fix where a template exists, a manual step
// otherwise. Passing, unknown and not-applicable controls are never
// touched.
func Render(meta Meta, findings []compliance.Finding) Bundle {
	templates := Templates()
	var b Bundle
	for _, f := range findings {
		if f.Status != compliance.StatusFail {
			continue
		}
		if t, ok := templates[f.ControlID]; ok {
			b.Automated = append(b.Automated, t)
		} else {
			b.Manual = append(b.Manual, ManualStep{ControlID: f.ControlID, Title: f.Title, Remediation: f.Remediation})
		}
	}
	sort.Slice(b.Automated, func(i, j int) bool { return b.Automated[i].ControlID < b.Automated[j].ControlID })
	sort.Slice(b.Manual, func(i, j int) bool { return b.Manual[i].ControlID < b.Manual[j].ControlID })

	short := meta.AssessmentID
	if len(short) > 8 {
		short = short[:8]
	}
	b.Filename = "cyberditto-remediation-" + short + ".ps1"
	b.Script = render(meta, short, b)
	return b
}

func render(meta Meta, short string, b Bundle) string {
	var s strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&s, format, args...) }

	w("<#\n.SYNOPSIS\nCyberDitto remediation bundle.\n\n.DESCRIPTION\n")
	w("Assessment %s (snapshot %s, native engine %s), generated %s.\n\n",
		meta.AssessmentID, meta.SnapshotID, meta.EngineVersion, meta.GeneratedAt.UTC().Format(time.RFC3339))
	w("Fixes %d failed control(s) automatically and lists %d that need manual work.\n\n", len(b.Automated), len(b.Manual))
	w("Run it without parameters first: that is a dry run and changes nothing.\n")
	w("Then - ideally on a twin first - run it as an administrator with -Apply.\n")
	w("Before changing anything it saves the current values to a rollback file in\n")
	w("the current directory; run it with -Rollback <file> to put them back.\n\n")
	w("CyberDitto never runs this itself. Review it before you run it.\n#>\n")
	w(`[CmdletBinding()]
param(
    [switch]$Apply,
    [string]$Rollback
)
Set-StrictMode -Version 3
$ErrorActionPreference = 'Stop'

function Test-IsAdmin {
    $id = [Security.Principal.WindowsIdentity]::GetCurrent()
    return ([Security.Principal.WindowsPrincipal]$id).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-RegDword([string]$Path, [string]$Name) {
    $item = Get-ItemProperty -LiteralPath $Path -Name $Name -ErrorAction SilentlyContinue
    if ($null -eq $item) { return $null }
    return [int]$item.$Name
}

function Set-RegDword([string]$Path, [string]$Name, [int]$Value) {
    if (-not (Test-Path -LiteralPath $Path)) { New-Item -Path $Path -Force | Out-Null }
    New-ItemProperty -LiteralPath $Path -Name $Name -PropertyType DWord -Value $Value -Force | Out-Null
}

`)

	w("$Actions = @(\n")
	for i, t := range b.Automated {
		w("    [pscustomobject]@{\n")
		w("        Id = %s\n        Summary = %s\n", psQuote(t.ControlID), psQuote(t.Summary))
		w("        Reboot = $%t\n        Risk = %s\n", t.RebootRequired, psQuote(t.Risk))
		if len(t.Registry) > 0 {
			w("        Registry = @(\n")
			for j, r := range t.Registry {
				w("            [pscustomobject]@{ Path = %s; Name = %s; Value = %d; AbsentCompliant = $%t; Accept = @(%s); Reject = @(%s) }",
					psQuote(r.Path), psQuote(r.Name), r.Value, r.AbsentCompliant, intList(r.accepted()), intList(r.Reject))
				if j < len(t.Registry)-1 {
					w(",")
				}
				w("\n")
			}
			w("        )\n        Get = $null\n        Test = $null\n        Set = $null\n        Restore = $null\n")
		} else {
			w("        Registry = @()\n")
			w("        Get = { %s }\n", t.Script.Get)
			w("        Test = { %s }\n", t.Script.Test)
			w("        Set = {\n%s\n        }\n", indent(t.Script.Set, "            "))
			w("        Restore = { %s }\n", t.Script.Restore)
		}
		w("    }")
		if i < len(b.Automated)-1 {
			w(",")
		}
		w("\n")
	}
	w(")\n\n")

	w("$Manual = @(\n")
	for i, m := range b.Manual {
		w("    [pscustomobject]@{ Id = %s; Title = %s; Remediation = %s }", psQuote(m.ControlID), psQuote(m.Title), psQuote(m.Remediation))
		if i < len(b.Manual)-1 {
			w(",")
		}
		w("\n")
	}
	w(")\n\n")

	w(`function Test-RegCompliant($r, $cur) {
    if ($null -eq $cur) { return [bool]$r.AbsentCompliant }
    if (@($r.Reject).Count -gt 0) { return -not (@($r.Reject) -contains $cur) }
    return (@($r.Accept) -contains $cur)
}

function Get-ActionState($a) {
    if (@($a.Registry).Count -gt 0) {
        return ,@($a.Registry | ForEach-Object { [pscustomobject]@{ Path = $_.Path; Name = $_.Name; Previous = (Get-RegDword $_.Path $_.Name) } })
    }
    return (& $a.Get)
}

function Test-ActionCompliant($a, $state) {
    if (@($a.Registry).Count -gt 0) {
        foreach ($r in $a.Registry) {
            if (-not (Test-RegCompliant $r (Get-RegDword $r.Path $r.Name))) { return $false }
        }
        return $true
    }
    return [bool](& $a.Test $state)
}

function Invoke-Action($a) {
    if (@($a.Registry).Count -gt 0) {
        foreach ($r in $a.Registry) {
            # Only touch a value that is actually non-compliant: never
            # rewrite a stronger compliant setting (e.g. RunAsPPL = 2).
            if (-not (Test-RegCompliant $r (Get-RegDword $r.Path $r.Name))) { Set-RegDword $r.Path $r.Name $r.Value }
        }
        return
    }
    & $a.Set
}

function Restore-Action($a, $state) {
    if (@($a.Registry).Count -gt 0) {
        foreach ($p in @($state)) {
            if ($null -eq $p.Previous) {
                Remove-ItemProperty -LiteralPath $p.Path -Name $p.Name -ErrorAction SilentlyContinue
            } else {
                Set-RegDword $p.Path $p.Name ([int]$p.Previous)
            }
        }
        return
    }
    & $a.Restore $state
}

if ($Rollback) {
    if (-not (Test-IsAdmin)) { throw 'Run as an administrator to roll back.' }
    $saved = Get-Content -Raw -LiteralPath $Rollback | ConvertFrom-Json
    foreach ($entry in @($saved.changes)) {
        $a = $Actions | Where-Object { $_.Id -eq $entry.id }
        if ($null -eq $a) { Write-Warning "unknown control $($entry.id) in the rollback file; skipped"; continue }
        try {
            Restore-Action $a $entry.state
            Write-Host "[restored]   $($entry.id)"
        } catch {
            Write-Warning "[failed]     $($entry.id): $($_.Exception.Message)"
        }
    }
    Write-Host 'Rollback finished. Settings marked reboot-required revert fully after a reboot.'
    return
}

if ($Apply -and -not (Test-IsAdmin)) { throw 'Run as an administrator to apply.' }

$pending = @()
foreach ($a in $Actions) {
    $state = Get-ActionState $a
    if (Test-ActionCompliant $a $state) {
        Write-Host "[ok]         $($a.Id) already compliant"
        continue
    }
    $pending += [pscustomobject]@{ id = $a.Id; state = $state }
    if (-not $Apply) {
        Write-Host "[would fix]  $($a.Id): $($a.Summary)"
        if ($a.Risk) { Write-Host "             risk: $($a.Risk)" }
    }
}

if (-not $Apply) {
    Write-Host ''
    Write-Host "Dry run: $($pending.Count) change(s) would be made. Nothing was changed. Re-run with -Apply to make them."
} elseif ($pending.Count -gt 0) {
`)
	w("    $file = Join-Path (Get-Location) ('cyberditto-rollback-%s-' + (Get-Date -Format 'yyyyMMdd-HHmmss') + '.json')\n", short)
	w("    [pscustomobject]@{ assessment = %s; created = (Get-Date).ToString('o'); changes = $pending } | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $file -Encoding UTF8\n", psQuote(meta.AssessmentID))
	w(`    Write-Host "Saved the previous values to $file"
    $reboot = $false
    foreach ($p in $pending) {
        $a = $Actions | Where-Object { $_.Id -eq $p.id }
        try {
            Invoke-Action $a
            if (Test-ActionCompliant $a (Get-ActionState $a)) {
                Write-Host "[fixed]      $($a.Id)"
                if ($a.Reboot) { $reboot = $true }
            } else {
                Write-Warning "[unverified] $($a.Id): the change was made but the check still fails"
            }
        } catch {
            Write-Warning "[failed]     $($a.Id): $($_.Exception.Message)"
        }
    }
    if ($reboot) { Write-Host 'Some fixes take effect after a reboot.' }
    Write-Host "To undo: & '$($MyInvocation.MyCommand.Path)' -Rollback '$file'"
} else {
    Write-Host 'Everything this bundle fixes is already compliant.'
}

if ($Manual.Count -gt 0) {
    Write-Host ''
    Write-Host 'Needs manual attention (no automated fix in this bundle):'
    foreach ($m in $Manual) { Write-Host " - $($m.Id) $($m.Title): $($m.Remediation)" }
}
`)
	return s.String()
}

func intList(v []int) string {
	parts := make([]string, len(v))
	for i, n := range v {
		parts[i] = fmt.Sprint(n)
	}
	return strings.Join(parts, ", ")
}

func indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
