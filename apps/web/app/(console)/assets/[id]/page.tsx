"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState, type ChangeEvent } from "react";
import { api, ApiError, type Asset, type Assessment, type SnapshotSummary } from "@/lib/api";

export default function AssetDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const [asset, setAsset] = useState<Asset | null>(null);
  const [snapshots, setSnapshots] = useState<SnapshotSummary[] | null>(null);
  const [history, setHistory] = useState<Assessment[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const [a, s, h] = await Promise.all([api.getAsset(id), api.listSnapshots(id), api.listAssetAssessments(id)]);
      setAsset(a);
      setSnapshots(s.snapshots);
      setHistory(h.assessments);
    } catch (err) {
      setError(err instanceof ApiError && err.status === 404 ? "Asset not found." : "Could not load the asset.");
    }
  }, [id]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  async function onUpload(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;
    setBusy(true);
    setError(null);
    try {
      await api.uploadSnapshot(id, file);
      await refresh();
    } catch (err) {
      setError(err instanceof ApiError ? `${err.message}${err.detail ? `: ${err.detail}` : ""}` : "Upload failed.");
    } finally {
      setBusy(false);
    }
  }

  async function onBuildTwin(snapshotId: string) {
    setBusy(true);
    setError(null);
    try {
      await api.createTwin(snapshotId);
      router.push("/twins");
    } catch (err) {
      setError(err instanceof ApiError ? err.detail || err.message : "Could not build a twin.");
      setBusy(false);
    }
  }

  async function onAssess(snapshotId: string) {
    setBusy(true);
    setError(null);
    try {
      const a = await api.runAssessment({ snapshot_id: snapshotId });
      router.push(`/assessments/${a.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.detail || err.message : "Could not run the assessment.");
      setBusy(false);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">{asset?.name ?? "Asset"}</h1>
        <p className="font-mono text-xs text-zinc-500">{id}</p>
      </div>

      <div className="space-y-2 rounded-lg border border-white/10 p-4 text-sm">
        <p className="font-medium">Upload a snapshot</p>
        <p className="text-zinc-400">
          Run the collector on the machine (<code className="font-mono">collector --consent --output snapshot.json</code>)
          and upload the file it writes.
        </p>
        <input type="file" accept="application/json,.json" onChange={onUpload} disabled={busy} aria-label="Snapshot file" />
      </div>

      {error && <p className="text-sm text-rose-400">{error}</p>}

      <div>
        <h2 className="mb-2 font-medium">Snapshots</h2>
        {snapshots === null ? (
          <p className="text-sm text-zinc-500">Loading...</p>
        ) : snapshots.length === 0 ? (
          <p className="text-sm text-zinc-500">No snapshots yet.</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400">
              <tr>
                <th className="py-1">Version</th>
                <th>Uploaded</th>
                <th>SHA-256</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {snapshots.map((s) => (
                <tr key={s.id} className="border-t border-white/10">
                  <td className="py-2">v{s.version}</td>
                  <td>{new Date(s.created_at).toLocaleString()}</td>
                  <td className="font-mono text-xs" title={s.sha256}>
                    {s.sha256.slice(0, 16)}…
                  </td>
                  <td className="space-x-2 text-right">
                    <button
                      onClick={() => onAssess(s.id)}
                      disabled={busy}
                      className="rounded-md border border-white/10 px-3 py-1 text-xs hover:bg-white/5 disabled:opacity-30"
                    >
                      Assess
                    </button>
                    <button
                      onClick={() => onBuildTwin(s.id)}
                      disabled={busy}
                      className="rounded-md border border-white/10 px-3 py-1 text-xs hover:bg-white/5 disabled:opacity-30"
                    >
                      Build twin
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div>
        <h2 className="mb-2 font-medium">Assessment history</h2>
        {history.length === 0 ? (
          <p className="text-sm text-zinc-500">No assessments yet.</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400">
              <tr>
                <th className="py-1">When</th>
                <th>Snapshot</th>
                <th>Target</th>
                <th>Score</th>
                <th>Failed</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {history.map((a, i) => {
                const previous = history[i + 1];
                const version = snapshots?.find((s) => s.id === a.snapshot_id)?.version;
                return (
                  <tr key={a.id} className="border-t border-white/10">
                    <td className="py-2">
                      <Link href={`/assessments/${a.id}`} className="underline">
                        {new Date(a.created_at).toLocaleString()}
                      </Link>
                    </td>
                    <td>{version ? `v${version}` : "—"}</td>
                    <td className="text-xs text-zinc-400">{a.twin_id ? "twin" : "snapshot"}</td>
                    <td className="font-medium">{a.score ?? "n/a"}</td>
                    <td>{a.counts.failed}</td>
                    <td className="text-right">
                      {previous && previous.engine_version === a.engine_version && (
                        <Link href={`/assessments/${a.id}/diff?against=${previous.id}`} className="text-xs underline">
                          compare with previous
                        </Link>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
