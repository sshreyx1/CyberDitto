"use client";

import Link from "next/link";
import { useCallback, useEffect, useState, type FormEvent } from "react";
import { api, ApiError, type Asset } from "@/lib/api";

export default function AssetsPage() {
  const [assets, setAssets] = useState<Asset[] | null>(null);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const { assets: list } = await api.listAssets();
      setAssets(list);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not load assets.");
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  async function onCreate(e: FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError(null);
    try {
      await api.createAsset(name);
      setName("");
      await refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.detail || err.message : "Could not create the asset.");
    } finally {
      setCreating(false);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">Assets</h1>
        <p className="text-sm text-zinc-400">
          Machines you track. Upload a collector snapshot to an asset, then build a twin from it.
        </p>
      </div>

      <form onSubmit={onCreate} className="flex gap-2">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Asset name, e.g. finance-laptop-01"
          required
          maxLength={200}
          aria-label="Asset name"
          className="w-80 rounded-lg border border-white/10 bg-black/30 px-3 py-2 text-sm"
        />
        <button
          type="submit"
          disabled={creating}
          className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-500 disabled:opacity-50"
        >
          {creating ? "Adding..." : "Add asset"}
        </button>
      </form>

      {error && <p className="text-sm text-rose-400">{error}</p>}

      {assets === null ? (
        <p className="text-sm text-zinc-500">Loading...</p>
      ) : assets.length === 0 ? (
        <p className="text-sm text-zinc-500">No assets yet.</p>
      ) : (
        <ul className="space-y-2">
          {assets.map((a) => (
            <li key={a.id} className="rounded-lg border border-white/10 p-3">
              <Link href={`/assets/${a.id}`} className="font-medium hover:underline">
                {a.name}
              </Link>
              <p className="font-mono text-xs text-zinc-500">{a.id}</p>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
