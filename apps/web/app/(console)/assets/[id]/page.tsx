"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState, type ChangeEvent } from "react";
import { api, ApiError, type Asset, type SnapshotSummary } from "@/lib/api";

export default function AssetDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const [asset, setAsset] = useState<Asset | null>(null);
  const [snapshots, setSnapshots] = useState<SnapshotSummary[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const [a, s] = await Promise.all([api.getAsset(id), api.listSnapshots(id)]);
      setAsset(a);
      setSnapshots(s.snapshots);
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
                  <td className="text-right">
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
    </div>
  );
}
