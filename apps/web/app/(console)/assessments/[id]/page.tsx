"use client";

import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { api, ApiError, type Assessment, type FindingStatus } from "@/lib/api";

const STATUS_ORDER: FindingStatus[] = ["fail", "unknown", "pass", "not_applicable"];
const SEVERITY_ORDER = { high: 0, medium: 1, low: 2 } as const;

const STATUS_LABEL: Record<FindingStatus, string> = {
  fail: "Failed",
  unknown: "Unknown",
  pass: "Passed",
  not_applicable: "Not applicable",
};

const STATUS_COLOR: Record<FindingStatus, string> = {
  fail: "text-rose-400",
  unknown: "text-amber-400",
  pass: "text-emerald-400",
  not_applicable: "text-zinc-500",
};

export default function AssessmentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [a, setA] = useState<Assessment | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<FindingStatus | "all">("all");

  useEffect(() => {
    api
      .getAssessment(id)
      .then(setA)
      .catch((err) => setError(err instanceof ApiError && err.status === 404 ? "Assessment not found." : "Could not load it."));
  }, [id]);

  const findings = useMemo(() => {
    const list = [...(a?.findings ?? [])].filter((f) => filter === "all" || f.status === filter);
    return list.sort(
      (x, y) =>
        STATUS_ORDER.indexOf(x.status) - STATUS_ORDER.indexOf(y.status) ||
        SEVERITY_ORDER[x.severity] - SEVERITY_ORDER[y.severity] ||
        x.control_id.localeCompare(y.control_id),
    );
  }, [a, filter]);

  if (error) return <p className="text-sm text-rose-400">{error}</p>;
  if (!a) return <p className="text-sm text-zinc-500">Loading...</p>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">
          Posture score: {a.score === null ? "n/a" : a.score}
          {a.score !== null && <span className="text-sm text-zinc-400"> / 100</span>}
        </h1>
        <p className="text-sm text-zinc-400">
          {a.counts.passed} passed · {a.counts.failed} failed · {a.counts.unknown} unknown · {a.counts.not_applicable} not
          applicable — {a.engine} engine {a.engine_version}, {new Date(a.created_at).toLocaleString()}
        </p>
        <p className="mt-1 text-xs text-zinc-500">
          Score = 100 × (1 − Σ weight of failed / Σ weight of passed+failed), weights high 5, medium 3, low 1. Unknown and
          not-applicable controls are excluded, not counted as passes.
        </p>
        <p className="font-mono text-xs text-zinc-500">
          snapshot {a.snapshot_id}
          {a.twin_id && ` · twin ${a.twin_id}`}
        </p>
      </div>

      <div className="flex gap-2 text-sm">
        {(["all", ...STATUS_ORDER] as const).map((s) => (
          <button
            key={s}
            onClick={() => setFilter(s)}
            className={`rounded-md border px-3 py-1 ${filter === s ? "border-violet-500" : "border-white/10"}`}
          >
            {s === "all" ? "All" : STATUS_LABEL[s]}
          </button>
        ))}
      </div>

      <ul className="space-y-2">
        {findings.map((f) => (
          <li key={f.control_id} className="rounded-lg border border-white/10 p-3 text-sm">
            <details>
              <summary className="cursor-pointer list-none">
                <span className={`mr-2 font-medium ${STATUS_COLOR[f.status]}`}>{STATUS_LABEL[f.status]}</span>
                <span className="mr-2 font-mono text-xs text-zinc-500">{f.control_id}</span>
                <span>{f.title}</span>
                <span className="ml-2 text-xs text-zinc-500">
                  {f.severity} · {f.category}
                </span>
              </summary>
              <div className="mt-3 space-y-2 text-zinc-300">
                {f.reason && (
                  <p>
                    <span className="text-zinc-500">Why: </span>
                    {f.reason}
                  </p>
                )}
                {f.expected && (
                  <p>
                    <span className="text-zinc-500">Expected: </span>
                    {f.expected}
                  </p>
                )}
                {f.observed !== null && f.observed !== undefined && (
                  <div>
                    <span className="text-zinc-500">Observed (from the snapshot):</span>
                    <pre className="mt-1 overflow-x-auto rounded bg-black/30 p-2 text-xs">{JSON.stringify(f.observed, null, 2)}</pre>
                  </div>
                )}
                {f.status === "fail" && (
                  <p>
                    <span className="text-zinc-500">Fix: </span>
                    {f.remediation}
                  </p>
                )}
                {f.techniques.length > 0 && (
                  <p className="text-xs text-zinc-500">ATT&amp;CK: {f.techniques.join(", ")}</p>
                )}
              </div>
            </details>
          </li>
        ))}
      </ul>
    </div>
  );
}
