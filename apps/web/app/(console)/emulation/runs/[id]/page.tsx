"use client";

import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { api, ApiError, type EmulationRun, type Outcome } from "@/lib/api";

const OUTCOME_COLOR: Record<Outcome, string> = {
  blocked: "text-emerald-400",
  detected: "text-sky-400",
  logged: "text-amber-400",
  missed: "text-rose-400",
  not_executed: "text-zinc-500",
  error: "text-rose-400",
  skipped: "text-zinc-500",
};

export default function RunDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [run, setRun] = useState<EmulationRun | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [aborting, setAborting] = useState(false);

  const load = useCallback(async () => {
    try {
      setRun(await api.getRun(id));
    } catch (err) {
      setError(err instanceof ApiError && err.status === 404 ? "Run not found." : "Could not load the run.");
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  // Poll while the run is in progress.
  useEffect(() => {
    if (run?.status !== "running") return;
    const t = setInterval(load, 1000);
    return () => clearInterval(t);
  }, [run?.status, load]);

  async function onAbort() {
    setAborting(true);
    try {
      setRun(await api.abortRun(id));
    } catch (err) {
      setError(err instanceof ApiError ? err.detail || err.message : "Abort failed.");
    } finally {
      setAborting(false);
    }
  }

  if (error) return <p className="text-sm text-rose-400">{error}</p>;
  if (!run) return <p className="text-sm text-zinc-500">Loading...</p>;

  const allNotExecuted = run.executions.length > 0 && run.executions.every((e) => e.outcome === "not_executed");

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold">
            Run: {run.scenario_id || "custom"} — {run.status.replace("_", " ")}
          </h1>
          <p className="text-sm text-zinc-400">{run.detail}</p>
          <p className="font-mono text-xs text-zinc-500">
            twin {run.twin_id} · provider {run.provider} · catalog {run.catalog_source}@{run.catalog_commit.slice(0, 12)} ·{" "}
            {new Date(run.created_at).toLocaleString()}
          </p>
        </div>
        {run.status === "running" && (
          <button
            onClick={onAbort}
            disabled={aborting}
            className="rounded-lg bg-rose-600 px-4 py-2 text-sm font-medium text-white hover:bg-rose-500 disabled:opacity-50"
          >
            {aborting ? "Stopping..." : "Kill switch: abort and revert"}
          </button>
        )}
      </div>

      {(run.provider === "fake" || allNotExecuted) && run.status !== "preflight_failed" && (
        <p className="rounded-lg border border-white/10 bg-white/[0.03] p-3 text-sm text-zinc-300">
          Nothing was actually executed. This instance uses the fake lab provider, which has no virtual machine, so every
          test is recorded as <span className="font-mono">not_executed</span> rather than guessed. Real outcomes (blocked,
          detected, logged, missed) need the Hyper-V provider.
        </p>
      )}

      <div>
        <h2 className="mb-2 font-medium">Preflight</h2>
        <table className="w-full text-left text-sm">
          <tbody>
            {run.preflight.map((c) => (
              <tr key={c.id} className="border-t border-white/10 align-top">
                <td className={`py-1 pr-3 font-medium ${c.passed ? "text-emerald-400" : "text-rose-400"}`}>
                  {c.passed ? "pass" : "FAIL"}
                </td>
                <td className="pr-3">{c.name}</td>
                <td className="text-xs text-zinc-400">{c.detail}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {run.authorisation_statement && (
          <p className="mt-2 text-xs text-zinc-500">
            Acknowledged by {run.authorised_by}: “{run.authorisation_statement}”
          </p>
        )}
      </div>

      <div>
        <h2 className="mb-2 font-medium">Tests ({run.executions.length}/{run.test_guids.length})</h2>
        {run.executions.length === 0 ? (
          <p className="text-sm text-zinc-500">{run.status === "preflight_failed" ? "Preflight refused the run; nothing ran." : "No results yet."}</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400">
              <tr>
                <th className="py-1">Technique</th>
                <th>Test</th>
                <th>Tier</th>
                <th>Outcome</th>
              </tr>
            </thead>
            <tbody>
              {run.executions.map((e, i) => (
                <tr key={`${e.test_guid}-${i}`} className="border-t border-white/10 align-top">
                  <td className="py-1 font-mono text-xs">{e.technique_id}</td>
                  <td>{e.test_name}</td>
                  <td>{e.risk_tier}</td>
                  <td className={OUTCOME_COLOR[e.outcome]} title={e.detail}>
                    {e.outcome.replace("_", " ")}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
