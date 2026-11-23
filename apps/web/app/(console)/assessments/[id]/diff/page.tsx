"use client";

import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { api, ApiError, type AssessmentDiff } from "@/lib/api";

const DIRECTION_COLOR = {
  regressed: "text-rose-400",
  improved: "text-emerald-400",
  changed: "text-amber-400",
} as const;

function DiffView() {
  const { id } = useParams<{ id: string }>();
  const against = useSearchParams().get("against") ?? "";
  const [diff, setDiff] = useState<AssessmentDiff | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .diffAssessments(id, against)
      .then(setDiff)
      .catch((err) => setError(err instanceof ApiError ? err.detail || err.message : "Could not load the comparison."));
  }, [id, against]);

  if (error) return <p className="text-sm text-rose-400">{error}</p>;
  if (!diff) return <p className="text-sm text-zinc-500">Loading...</p>;

  const { before, after, changes } = diff;
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">
          Score {before.score ?? "n/a"} → {after.score ?? "n/a"}
        </h1>
        <p className="text-sm text-zinc-400">
          <Link href={`/assessments/${before.id}`} className="underline">
            {new Date(before.created_at).toLocaleString()}
          </Link>{" "}
          →{" "}
          <Link href={`/assessments/${after.id}`} className="underline">
            {new Date(after.created_at).toLocaleString()}
          </Link>
        </p>
      </div>

      {changes.length === 0 ? (
        <p className="text-sm text-zinc-500">No control changed status.</p>
      ) : (
        <table className="w-full text-left text-sm">
          <thead className="text-zinc-400">
            <tr>
              <th className="py-1">Control</th>
              <th>Severity</th>
              <th>Before</th>
              <th>After</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {changes.map((c) => (
              <tr key={c.control_id} className="border-t border-white/10">
                <td className="py-2">
                  <span className="mr-2 font-mono text-xs text-zinc-500">{c.control_id}</span>
                  {c.title}
                </td>
                <td>{c.severity}</td>
                <td>{c.before || "—"}</td>
                <td>{c.after || "—"}</td>
                <td className={DIRECTION_COLOR[c.direction]}>{c.direction}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default function AssessmentDiffPage() {
  return (
    <Suspense fallback={<p className="text-sm text-zinc-500">Loading...</p>}>
      <DiffView />
    </Suspense>
  );
}
