"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState, type FormEvent } from "react";
import { api, ApiError, type AttackCatalog, type EmulationRun, type Twin } from "@/lib/api";

const STATUS_COLOR: Record<string, string> = {
  completed: "text-emerald-400",
  running: "text-sky-400",
  aborted: "text-amber-400",
  failed: "text-rose-400",
  preflight_failed: "text-rose-400",
};

export default function EmulationPage() {
  const router = useRouter();
  const [catalog, setCatalog] = useState<AttackCatalog | null>(null);
  const [twins, setTwins] = useState<Twin[]>([]);
  const [runs, setRuns] = useState<EmulationRun[]>([]);
  const [twinId, setTwinId] = useState("");
  const [scenarioId, setScenarioId] = useState("");
  const [acknowledged, setAcknowledged] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    Promise.all([api.getAttackCatalog(), api.listTwins(), api.listRuns()])
      .then(([c, t, r]) => {
        setCatalog(c);
        setTwins(t.twins.filter((x) => x.state === "ready" || x.state === "running" || x.state === "stopped"));
        setRuns(r.runs);
        setScenarioId(c.scenarios[0]?.id ?? "");
      })
      .catch((err) => setError(err instanceof ApiError ? err.message : "Could not load emulation data."));
  }, []);

  const scenario = catalog?.scenarios.find((s) => s.id === scenarioId);
  const scenarioTests = useMemo(
    () => (scenario && catalog ? scenario.test_guids.map((g) => catalog.tests.find((t) => t.guid === g)).filter((t) => !!t) : []),
    [scenario, catalog],
  );

  async function onStart(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const run = await api.startRun({ twin_id: twinId, scenario_id: scenarioId, authorisation_acknowledged: acknowledged });
      router.push(`/emulation/runs/${run.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.detail || err.message : "Could not start the run.");
      setBusy(false);
    }
  }

  if (!catalog) return <p className="text-sm text-zinc-500">{error ?? "Loading..."}</p>;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-xl font-semibold">Emulation</h1>
        <p className="text-sm text-zinc-400">
          Curated Atomic Red Team tests, pinned to{" "}
          <span className="font-mono">
            {catalog.repository}@{catalog.commit.slice(0, 12)}
          </span>{" "}
          ({catalog.commit_date}, {catalog.license}).{" "}
          {catalog.verified ? (
            <span className="text-emerald-400">Catalog verified.</span>
          ) : (
            <span className="text-rose-400">Catalog NOT verified — every run will be refused: {catalog.verification_detail}</span>
          )}
        </p>
      </div>

      <form onSubmit={onStart} className="space-y-4 rounded-lg border border-white/10 p-4 text-sm">
        <h2 className="font-medium">Start a run</h2>
        <div className="flex flex-wrap gap-4">
          <label className="block">
            <span className="mb-1 block text-zinc-400">Twin (built twins only)</span>
            <select
              value={twinId}
              onChange={(e) => setTwinId(e.target.value)}
              required
              className="rounded-lg border border-white/10 bg-black/30 px-3 py-2"
            >
              <option value="">Choose a twin…</option>
              {twins.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.id.slice(0, 8)} · {t.state}
                </option>
              ))}
            </select>
          </label>
          <label className="block">
            <span className="mb-1 block text-zinc-400">Scenario</span>
            <select
              value={scenarioId}
              onChange={(e) => setScenarioId(e.target.value)}
              className="rounded-lg border border-white/10 bg-black/30 px-3 py-2"
            >
              {catalog.scenarios.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name} ({s.test_guids.length} tests)
                </option>
              ))}
            </select>
          </label>
        </div>
        {twins.length === 0 && (
          <p className="text-zinc-500">
            No built twins. <Link href="/assets" className="underline">Build one from an asset&apos;s snapshot</Link> first.
          </p>
        )}

        {scenario && (
          <div>
            <p className="text-zinc-400">{scenario.description}</p>
            <table className="mt-2 w-full text-left">
              <thead className="text-zinc-500">
                <tr>
                  <th className="py-1">Technique</th>
                  <th>Test</th>
                  <th>Tier</th>
                  <th>Elevated</th>
                </tr>
              </thead>
              <tbody>
                {scenarioTests.map((t) => (
                  <tr key={t.guid} className="border-t border-white/10 align-top">
                    <td className="py-1 font-mono text-xs">
                      {t.technique_id}
                      <div className="font-sans text-zinc-500">{t.technique_name}</div>
                    </td>
                    <td>
                      <details>
                        <summary className="cursor-pointer">{t.name}</summary>
                        <p className="mt-1 text-xs text-zinc-400">{t.curation_note}</p>
                        <pre className="mt-1 overflow-x-auto rounded bg-black/30 p-2 text-xs">{t.command}</pre>
                        {t.cleanup && (
                          <pre className="mt-1 overflow-x-auto rounded bg-black/30 p-2 text-xs">cleanup: {t.cleanup}</pre>
                        )}
                      </details>
                    </td>
                    <td className={t.risk_tier === "intrusive" ? "text-amber-400" : "text-zinc-300"}>{t.risk_tier}</td>
                    <td>{t.elevation_required ? "yes" : "no"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <label className="flex items-start gap-2">
          <input type="checkbox" checked={acknowledged} onChange={(e) => setAcknowledged(e.target.checked)} className="mt-1" />
          <span>{catalog.authorisation_statement}</span>
        </label>
        <p className="text-xs text-zinc-500">
          Preflight checks the twin is built, the lab is isolated, the catalog is verified, every test is from the catalog
          and not destructive, the run is within limits ({catalog.limits.max_tests_per_run} tests,{" "}
          {catalog.limits.max_run_seconds / 60} min), and you are an approver who acknowledged the statement above. A refused
          run is still recorded, with the reason.
        </p>
        {error && <p className="text-rose-400">{error}</p>}
        <button
          type="submit"
          disabled={busy || !twinId}
          className="rounded-lg bg-violet-600 px-4 py-2 font-medium text-white hover:bg-violet-500 disabled:opacity-50"
        >
          {busy ? "Starting..." : "Run preflight and start"}
        </button>
      </form>

      <div>
        <h2 className="mb-2 font-medium">Runs</h2>
        {runs.length === 0 ? (
          <p className="text-sm text-zinc-500">No runs yet.</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400">
              <tr>
                <th className="py-1">Started</th>
                <th>Scenario</th>
                <th>Twin</th>
                <th>Tests</th>
                <th>Status</th>
                <th>Provider</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((r) => (
                <tr key={r.id} className="border-t border-white/10">
                  <td className="py-2">
                    <Link href={`/emulation/runs/${r.id}`} className="underline">
                      {new Date(r.created_at).toLocaleString()}
                    </Link>
                  </td>
                  <td>{r.scenario_id || "custom"}</td>
                  <td className="font-mono text-xs">{r.twin_id.slice(0, 8)}</td>
                  <td>{r.test_guids.length}</td>
                  <td className={STATUS_COLOR[r.status] ?? ""}>{r.status.replace("_", " ")}</td>
                  <td className="text-xs text-zinc-400">{r.provider}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
