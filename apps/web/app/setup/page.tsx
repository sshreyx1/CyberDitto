"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { Field } from "@/components/form-field";

export default function SetupPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await api.setup(email, password);
      router.push("/login");
    } catch (err) {
      setError(err instanceof ApiError ? err.detail || err.message : "Setup failed.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center p-8">
      <form onSubmit={onSubmit} className="w-full max-w-sm space-y-4 rounded-xl border border-white/10 bg-white/[0.03] p-8">
        <div>
          <h1 className="text-xl font-semibold">Create the owner account</h1>
          <p className="mt-1 text-sm text-zinc-400">
            This only happens once, for the first person to set up this CyberDitto instance.
          </p>
        </div>
        <Field label="Email" type="email" value={email} onChange={setEmail} required autoComplete="email" />
        <Field
          label="Password (12+ characters)"
          type="password"
          value={password}
          onChange={setPassword}
          required
          minLength={12}
          autoComplete="new-password"
        />
        {error && <p className="text-sm text-rose-400">{error}</p>}
        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-500 disabled:opacity-50"
        >
          {submitting ? "Creating..." : "Create owner account"}
        </button>
      </form>
    </main>
  );
}
