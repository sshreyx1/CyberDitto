"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, ApiError, type Assessment } from "@/lib/api";

function formatScore(score: number | null): string {
  return score === null ? "n/a" : score.toFixed(score % 1 === 0 ? 0 : 2);
}

export default function AssessmentsPage() {
  const [items, setItems] = useState<Assessment[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .listAssessments()
      .then((r) => setItems(r.assessments))
      .catch((err) => setError(err instanceof ApiError ? err.message : "Could not load assessments."));
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">Assessments</h1>
        <p className="text-sm text-zinc-400">
          Compliance runs of the native engine against a snapshot or a twin. Run one from an asset&apos;s snapshot list or a
          twin.
        </p>
      </div>

      {error && <p className="text-sm text-rose-400">{error}</p>}

      {items === null ? (
        <p className="text-sm text-zinc-500">Loading...</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-zinc-500">No assessments yet.</p>
      ) : (
        <table className="w-full text-left text-sm">
          <thead className="text-zinc-400">
            <tr>
              <th className="py-1">When</th>
              <th>Target</th>
              <th>Score</th>
              <th>Passed</th>
              <th>Failed</th>
              <th>Unknown</th>
              <th>N/A</th>
              <th>Engine</th>
            </tr>
          </thead>
          <tbody>
            {items.map((a) => (
              <tr key={a.id} className="border-t border-white/10">
                <td className="py-2">
                  <Link href={`/assessments/${a.id}`} className="underline">
                    {new Date(a.created_at).toLocaleString()}
                  </Link>
                </td>
                <td className="text-xs text-zinc-400">{a.twin_id ? "twin" : "snapshot"}</td>
                <td className="font-medium">{formatScore(a.score)}</td>
                <td>{a.counts.passed}</td>
                <td>{a.counts.failed}</td>
                <td>{a.counts.unknown}</td>
                <td>{a.counts.not_applicable}</td>
                <td className="font-mono text-xs text-zinc-500">
                  {a.engine} {a.engine_version}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
