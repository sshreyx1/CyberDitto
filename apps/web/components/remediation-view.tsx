"use client";

import { useEffect, useState } from "react";
import { api, ApiError, type RemediationPlan } from "@/lib/api";

export function RemediationView({ assessmentId }: { assessmentId: string }) {
  const [plan, setPlan] = useState<RemediationPlan | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api
      .getRemediationPlan(assessmentId)
      .then(setPlan)
      .catch((err) => setError(err instanceof ApiError ? err.message : "Could not load the remediation plan."));
  }, [assessmentId]);

  async function onDownload() {
    if (!plan) return;
    setBusy(true);
    setError(null);
    try {
      const blob = await api.downloadRemediationScript(assessmentId);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = plan.filename;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Download failed.");
    } finally {
      setBusy(false);
    }
  }

  if (error) return <p className="text-sm text-rose-400">{error}</p>;
  if (!plan) return <p className="text-sm text-zinc-500">Loading...</p>;
  if (plan.automated.length === 0 && plan.manual.length === 0) {
    return <p className="text-sm text-zinc-500">Nothing failed, so there is nothing to remediate.</p>;
  }

  return (
    <div className="space-y-5 text-sm">
      <div className="rounded-lg border border-white/10 bg-white/[0.03] p-3 text-zinc-300">
        The bundle is a PowerShell script for the real machine. Run it with no parameters first: that is a dry run and
        changes nothing. Running it with <span className="font-mono">-Apply</span> (as an administrator) saves every current
        value to a rollback file before changing anything, then verifies each fix;{" "}
        <span className="font-mono">-Rollback &lt;file&gt;</span> undoes it. CyberDitto never runs it for you. Try it on a
        twin first once a real lab is available.
      </div>
      <button
        onClick={onDownload}
        disabled={busy || plan.automated.length === 0}
        className="rounded-lg bg-violet-600 px-4 py-2 font-medium text-white hover:bg-violet-500 disabled:opacity-50"
      >
        {busy ? "Preparing..." : `Download ${plan.filename}`}
      </button>

      <div>
        <h3 className="mb-2 font-medium">Automated, reversible fixes ({plan.automated.length})</h3>
        <ul className="space-y-2">
          {plan.automated.map((f) => (
            <li key={f.control_id} className="rounded-lg border border-white/10 p-3">
              <span className="mr-2 font-mono text-xs text-zinc-500">{f.control_id}</span>
              {f.summary}
              {f.reboot_required && <span className="ml-2 text-xs text-amber-400">reboot required</span>}
              {f.risk && <p className="mt-1 text-xs text-zinc-400">Risk: {f.risk}</p>}
            </li>
          ))}
        </ul>
      </div>

      {plan.manual.length > 0 && (
        <div>
          <h3 className="mb-2 font-medium">Needs manual work ({plan.manual.length})</h3>
          <ul className="space-y-2">
            {plan.manual.map((m) => (
              <li key={m.control_id} className="rounded-lg border border-white/10 p-3">
                <span className="mr-2 font-mono text-xs text-zinc-500">{m.control_id}</span>
                {m.title}
                <p className="mt-1 text-xs text-zinc-400">{m.remediation}</p>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
