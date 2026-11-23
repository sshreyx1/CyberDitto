"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, type Me } from "@/lib/api";

export default function ConsoleLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const router = useRouter();
  const [me, setMe] = useState<Me | null>(null);

  useEffect(() => {
    let cancelled = false;
    api
      .me()
      .then((result) => {
        if (!cancelled) setMe(result);
      })
      .catch(() => {
        if (!cancelled) router.replace("/login");
      });
    return () => {
      cancelled = true;
    };
  }, [router]);

  async function onLogout() {
    try {
      await api.logout();
    } finally {
      router.replace("/login");
    }
  }

  return (
    <div className="min-h-screen">
      <header className="flex items-center justify-between border-b border-white/10 px-6 py-4">
        <div className="flex items-center gap-6">
          <span className="font-semibold">CyberDitto</span>
          <nav className="flex gap-4 text-sm text-zinc-400">
            <Link href="/assets" className="hover:text-zinc-100">
              Assets
            </Link>
            <Link href="/twins" className="hover:text-zinc-100">
              Twins
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-4 text-sm text-zinc-400">
          {me && (
            <span>
              {me.email} &middot; {me.role}
            </span>
          )}
          <button onClick={onLogout} className="rounded-md border border-white/10 px-3 py-1 hover:bg-white/5">
            Log out
          </button>
        </div>
      </header>
      <main className="p-6">{children}</main>
    </div>
  );
}
