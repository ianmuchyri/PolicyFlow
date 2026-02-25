"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Loader2, Shield } from "lucide-react";
import { setToken } from "@/lib/auth";

// Inner component uses useSearchParams — must be wrapped in Suspense for static export.
function TokenHandler() {
  const router = useRouter();
  const searchParams = useSearchParams();

  useEffect(() => {
    const token = searchParams.get("token");
    if (token) {
      setToken(token);
      router.replace("/policies");
    } else {
      router.replace("/");
    }
  }, [router, searchParams]);

  return null;
}

// This page handles the redirect from GET /api/magic-login?token=...
// The Go server redirects to /auth-callback?token=<session-jwt>
export default function AuthCallbackPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-900">
      <div className="text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-blue-600 mb-4">
          <Shield className="w-9 h-9 text-white" />
        </div>
        <div className="flex items-center gap-2 text-slate-600 dark:text-slate-300">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>Logging you in…</span>
        </div>
      </div>
      <Suspense>
        <TokenHandler />
      </Suspense>
    </div>
  );
}
