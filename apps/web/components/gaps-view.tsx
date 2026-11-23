"use client";

import { useEffect, useState } from "react";
import { api, ApiError, type GapsResult } from "@/lib/api";

const EVIDENCE_COLOR = {
  succeeded: "text-rose-400",
  blocked: "text-emerald-400",
  untested: "text-zinc-500",
} as const;

export function GapsView({ assessmentId }: { assessmentId: string }) {
  const [data, setData] = useState<GapsResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .getGaps(assessmentId)
      .then(setData)
      .catch((err) => setError(err instanceof ApiError ? err.message : "Could not load gaps."));
  }, [assessmentId]);

  if (error) return <p className="text-sm text-rose-400">{error}</p>;
  if (!data) return <p className="text-sm text-zinc-500">Loading...</p>;

  const twins = data.evidence_base.twins?.length ?? 0;
  return (
    <div className="space-y-4 text-sm">
      <p className="text-zinc-400">
        {data.gaps.length} failed controls, ranked by {data.formula}. Evidence base: {twins} twin
        {twins === 1 ? "" : "s"} of this snapshot, {data.evidence_base.runs} finished run
        {data.evidence_base.runs === 1 ? "" : "s"}, {data.evidence_base.observations} test result
        {data.evidence_base.observations === 1 ? "" : "s"}. Only real executions count as evidence; a technique nobody has
        executed is &quot;untested&quot;, not assumed to work.
      </p>
      {data.gaps.length === 0 ? (
        <p className="text-zinc-500">No failed controls, so no gaps.</p>
      ) : (
        <ol className="space-y-2">
          {data.gaps.map((g, i) => (
            <li key={g.control_id} className="rounded-lg border border-white/10 p-3">
              <details>
                <summary className="cursor-pointer list-none">
                  <span className="mr-2 text-zinc-500">#{i + 1}</span>
                  <span className="mr-2 font-medium">{g.priority}</span>
                  <span className="mr-2 font-mono text-xs text-zinc-500">{g.control_id}</span>
                  {g.title}
                  <span className="ml-2 text-xs text-zinc-500">{g.severity}</span>
                </summary>
                <div className="mt-3 space-y-2 text-zinc-300">
                  <p>
                    <span className="text-zinc-500">Why it failed: </span>
                    {g.reason}
                  </p>
                  <table className="text-xs">
                    <tbody>
                      {g.factors.map((f) => (
                        <tr key={f.name}>
                          <td className="pr-3 text-zinc-500">{f.name}</td>
                          <td className="pr-3 font-mono">× {f.value}</td>
                          <td className="text-zinc-400">{f.explanation}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  <div className="text-xs">
                    <span className="text-zinc-500">Techniques this makes easier: </span>
                    {g.techniques.map((t) => (
                      <span key={t.technique_id} className="mr-3">
                        <span className="font-mono">{t.technique_id}</span>{" "}
                        <span className={EVIDENCE_COLOR[t.evidence]}>({t.evidence})</span>
                      </span>
                    ))}
                  </div>
                  <p>
                    <span className="text-zinc-500">Fix: </span>
                    {g.remediation}
                  </p>
                </div>
              </details>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}
