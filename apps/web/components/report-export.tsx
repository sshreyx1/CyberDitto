"use client";

import { useState } from "react";
import { api, ApiError, type ReportFormat } from "@/lib/api";

const FORMATS: { format: ReportFormat; label: string; hint: string }[] = [
  { format: "json", label: "JSON report", hint: "Findings, gaps and the score, machine-readable" },
  { format: "sarif", label: "SARIF 2.1.0", hint: "Failed and unknown controls as results, for SARIF viewers" },
  { format: "navigator", label: "ATT&CK layer", hint: "Open in MITRE ATT&CK Navigator; techniques scored by gap priority" },
];

export function ReportExport({ assessmentId }: { assessmentId: string }) {
  const [busy, setBusy] = useState<ReportFormat | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function download(format: ReportFormat) {
    setBusy(format);
    setError(null);
    try {
      const { blob, filename } = await api.downloadReport(assessmentId, format);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Export failed.");
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-2 text-sm">
      <span className="text-zinc-500">Export:</span>
      {FORMATS.map((f) => (
        <button
          key={f.format}
          title={f.hint}
          onClick={() => download(f.format)}
          disabled={busy !== null}
          className="rounded-md border border-white/10 px-3 py-1 hover:border-violet-500 disabled:opacity-50"
        >
          {busy === f.format ? "Preparing..." : f.label}
        </button>
      ))}
      {error && <span className="text-rose-400">{error}</span>}
    </div>
  );
}
