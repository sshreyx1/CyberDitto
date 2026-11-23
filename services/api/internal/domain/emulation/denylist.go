package emulation

import (
	"regexp"
	"strings"
)

// The destructive denylist is code, not configuration (§6.4 invariant 3,
// ADR 0006): changing it is a reviewed code change with a failing test if
// it's weakened.

// deniedTechniques match by prefix, so a parent ID also denies its
// sub-techniques (T1561 denies T1561.001).
var deniedTechniques = map[string]string{
	"T1485":     "data destruction",
	"T1486":     "data encrypted for impact (ransomware-like)",
	"T1490":     "inhibit system recovery",
	"T1491":     "defacement",
	"T1495":     "firmware corruption",
	"T1561":     "disk wipe",
	"T1529":     "system shutdown/reboot (ends the run's observation window)",
	"T1498":     "network denial of service",
	"T1499":     "endpoint denial of service",
	"T1046":     "network service scanning",
	"T1070.001": "clearing event logs destroys the evidence outcomes are judged on",
}

type commandRule struct {
	pattern *regexp.Regexp
	reason  string
}

// deniedCommands catch destructive or egress-dependent commands inside an
// otherwise-allowed technique. Matching is case-insensitive.
var deniedCommands = []commandRule{
	{regexp.MustCompile(`(?i)\bformat(\.com)?\s+[a-z]:`), "formats a volume"},
	{regexp.MustCompile(`(?i)vssadmin(\.exe)?\s+(delete|resize)\s+shadow`), "deletes volume shadow copies"},
	{regexp.MustCompile(`(?i)wmic(\.exe)?\s+shadowcopy\s+delete`), "deletes volume shadow copies"},
	{regexp.MustCompile(`(?i)\bwbadmin(\.exe)?\s+delete`), "deletes backups"},
	{regexp.MustCompile(`(?i)\bbcdedit\b`), "edits boot configuration"},
	{regexp.MustCompile(`(?i)\bcipher(\.exe)?\s+/w`), "wipes free space"},
	{regexp.MustCompile(`(?i)\bdiskpart\b`), "partitions disks"},
	{regexp.MustCompile(`(?i)\bshutdown(\.exe)?\s+/[rsp]\b`), "shuts down or restarts the machine"},
	{regexp.MustCompile(`(?i)\b(Restart|Stop)-Computer\b`), "shuts down or restarts the machine"},
	{regexp.MustCompile(`(?i)\bwevtutil(\.exe)?\s+(cl|clear-log)\b`), "clears event logs"},
	{regexp.MustCompile(`(?i)\bClear-EventLog\b`), "clears event logs"},
	// These target a drive root or the Windows/System32 directory as a
	// whole (C:\, C:\*, C:\Windows, C:\Windows\System32\*), not a single
	// named file beneath them — cleanups routinely delete their own marker
	// file under C:\Windows\Temp, which is fine.
	{regexp.MustCompile(`(?i)\b(del|erase)\s+(/[a-z]\s+)*["']?[a-z]:\\(windows(\\system32)?)?\\?(\*(\.\*)?)?["']?(\s|$)`), "deletes a drive root or the Windows directory"},
	{regexp.MustCompile(`(?i)\b(rd|rmdir)\s+(/[a-z]\s+)*["']?[a-z]:\\(windows(\\system32)?)?\\?["']?(\s|$)`), "removes a drive root or the Windows directory"},
	{regexp.MustCompile(`(?i)\bRemove-Item\b[^\n]*?\s["']?[a-z]:\\(windows(\\system32)?)?\\?\*?["']?(\s|$)`), "deletes a drive root or the Windows directory"},
	// The lab has no internet by design; a test that needs to download
	// something would fail anyway, and one that reaches out at all is
	// exactly the egress the lab is built to prevent.
	{regexp.MustCompile(`(?i)\b(Invoke-WebRequest|iwr|Invoke-RestMethod|irm|Start-BitsTransfer|bitsadmin)\b`), "downloads from the network"},
	{regexp.MustCompile(`(?i)\.(DownloadFile|DownloadString|DownloadData)\s*\(`), "downloads from the network"},
	{regexp.MustCompile(`(?i)\b(curl|wget)(\.exe)?\s`), "downloads from the network"},
	{regexp.MustCompile(`(?i)https?://`), "references a network URL"},
}

// Denied reports whether t may never be scheduled, and why. It checks the
// technique, the command, and the cleanup command.
func Denied(t Test) (bool, string) {
	for prefix, reason := range deniedTechniques {
		if t.TechniqueID == prefix || strings.HasPrefix(t.TechniqueID, prefix+".") {
			return true, "technique " + prefix + " is denied: " + reason
		}
	}
	for _, text := range []string{t.Command, t.Cleanup} {
		for _, rule := range deniedCommands {
			if rule.pattern.MatchString(text) {
				return true, "command " + rule.reason
			}
		}
	}
	for name, in := range t.Inputs {
		for _, rule := range deniedCommands {
			if rule.pattern.MatchString(in.Default) {
				return true, "input " + name + " default " + rule.reason
			}
		}
	}
	return false, ""
}
