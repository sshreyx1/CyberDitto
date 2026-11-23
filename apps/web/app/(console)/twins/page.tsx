"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { api, ApiError, type Twin, type TwinState } from "@/lib/api";

// States in which the fake lab provider is still stepping a twin toward
// Ready, and so is still worth an SSE subscription for (docs/MASTER_PLAN.md
// §5.7). Terminal/steady states (ready, running, stopped, failed,
// destroyed) have nothing more to stream.
const BUILD_STATES: ReadonlySet<TwinState> = new Set(["requested", "imaging", "provisioning", "applying", "verifying"]);

const STATE_COLOR: Partial<Record<TwinState, string>> = {
  ready: "text-emerald-400",
  running: "text-sky-400",
  stopped: "text-zinc-400",
  failed: "text-rose-400",
  destroyed: "text-zinc-500",
};

export default function TwinsPage() {
  const [twins, setTwins] = useState<Twin[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [logs, setLogs] = useState<Record<string, string[]>>({});

  const refresh = useCallback(async () => {
    try {
      const { twins: list } = await api.listTwins();
      setTwins(list);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not load twins.");
    }
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 2000);
    return () => clearInterval(interval);
  }, [refresh]);

  // One SSE subscription per twin still building, closed the moment the
  // set of building twin ids changes (a build finished, or a new twin
  // started one) rather than left to leak.
  const buildingIds = twins
    .filter((t) => BUILD_STATES.has(t.state))
    .map((t) => t.id)
    .join(",");

  useEffect(() => {
    if (!buildingIds) return;
    const sources = buildingIds.split(",").map((id) => {
      const source = new EventSource(api.twinEventsUrl(id), { withCredentials: true });
      source.onmessage = (event) => {
        const payload = JSON.parse(event.data) as { state: string; message: string };
        setLogs((prev) => ({
          ...prev,
          [id]: [...(prev[id] ?? []), `${payload.state}: ${payload.message}`],
        }));
      };
      return source;
    });
    return () => sources.forEach((s) => s.close());
  }, [buildingIds]);

  async function onAction(id: string, action: "start" | "stop" | "destroy") {
    setBusyId(id);
    setError(null);
    try {
      if (action === "start") await api.startTwin(id);
      if (action === "stop") await api.stopTwin(id);
      if (action === "destroy") await api.destroyTwin(id);
      await refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.detail || err.message : `Could not ${action} the twin.`);
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">Twins</h1>
        <p className="text-sm text-zinc-400">
          Built by the fake lab provider — no real hypervisor yet. To build one, open an{" "}
          <Link href="/assets" className="underline">
            asset
          </Link>{" "}
          and choose a snapshot.
        </p>
      </div>

      {error && <p className="text-sm text-rose-400">{error}</p>}

      {twins.length === 0 ? (
        <p className="text-sm text-zinc-500">No twins yet.</p>
      ) : (
        <ul className="space-y-3">
          {twins.map((t) => {
            const twinLogs = logs[t.id] ?? [];
            return (
              <li key={t.id} className="rounded-xl border border-white/10 bg-white/[0.03] p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-mono text-sm text-zinc-200">{t.id}</p>
                    <p className="font-mono text-xs text-zinc-500">snapshot {t.snapshot_id}</p>
                    <span className={`text-xs ${STATE_COLOR[t.state] ?? "text-amber-400"}`}>{t.state}</span>
                  </div>
                  <div className="flex gap-2">
                    <button
                      disabled={busyId === t.id || !(t.state === "ready" || t.state === "stopped")}
                      onClick={() => onAction(t.id, "start")}
                      className="rounded-md border border-white/10 px-3 py-1 text-xs hover:bg-white/5 disabled:opacity-30"
                    >
                      Start
                    </button>
                    <button
                      disabled={busyId === t.id || t.state !== "running"}
                      onClick={() => onAction(t.id, "stop")}
                      className="rounded-md border border-white/10 px-3 py-1 text-xs hover:bg-white/5 disabled:opacity-30"
                    >
                      Stop
                    </button>
                    <button
                      disabled={busyId === t.id || t.state === "destroyed"}
                      onClick={() => onAction(t.id, "destroy")}
                      className="rounded-md border border-white/10 px-3 py-1 text-xs text-rose-300 hover:bg-rose-500/10 disabled:opacity-30"
                    >
                      Destroy
                    </button>
                  </div>
                </div>
                {twinLogs.length > 0 && (
                  <ul className="mt-3 space-y-0.5 rounded-lg bg-black/30 p-3 font-mono text-xs text-zinc-400">
                    {twinLogs.map((line, i) => (
                      <li key={i}>{line}</li>
                    ))}
                  </ul>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
