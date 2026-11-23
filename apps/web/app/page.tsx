"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

// This page has no UI of its own — it decides where to send the visitor
// (first-run setup, login, or straight into the console) and gets out of
// the way. There is no marketing site yet (docs/MASTER_PLAN.md §10
// Phase 7), so "/" is purely a routing gate for now.
export default function Home() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const { initialized } = await api.setupStatus();
        if (cancelled) return;
        if (!initialized) {
          router.replace("/setup");
          return;
        }
        try {
          await api.me();
          if (!cancelled) router.replace("/twins");
        } catch {
          if (!cancelled) router.replace("/login");
        }
      } catch {
        if (!cancelled) setError("Could not reach the CyberDitto API. Is it running?");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [router]);

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 p-8 text-center">
      <h1 className="text-3xl font-semibold">CyberDitto</h1>
      <p className="max-w-md text-sm text-zinc-400">{error ?? "Checking setup status..."}</p>
    </main>
  );
}
