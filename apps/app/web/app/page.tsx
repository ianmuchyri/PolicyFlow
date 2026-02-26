"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Shield, Mail, ArrowRight, Loader2 } from "lucide-react";
import { isAuthenticated } from "@/lib/auth";
import { requestMagicLink } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "sent" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");

  useEffect(() => {
    if (isAuthenticated()) {
      router.replace("/policies");
    }
  }, [router]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!email) return;
    setStatus("loading");
    setErrorMsg("");
    try {
      await requestMagicLink(email);
      setStatus("sent");
    } catch (err: unknown) {
      setStatus("error");
      setErrorMsg(err instanceof Error ? err.message : "Something went wrong");
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-linear-to-br from-slate-50 to-blue-50 dark:from-slate-900 dark:to-slate-800 p-4">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-blue-600 mb-4 shadow-lg">
            <Shield className="w-9 h-9 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">PolicyFlow</h1>
          <p className="mt-2 text-slate-500 dark:text-slate-400">
            Policy management, simplified
          </p>
        </div>

        {/* Card */}
        <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 p-8">
          {status === "sent" ? (
            <div className="text-center py-4">
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-green-100 dark:bg-green-900/30 mb-4">
                <Mail className="w-7 h-7 text-green-600 dark:text-green-400" />
              </div>
              <h2 className="text-xl font-semibold text-slate-900 dark:text-white mb-2">
                Check your inbox
              </h2>
              <p className="text-slate-500 dark:text-slate-400 text-sm">
                We sent a login link to <strong>{email}</strong>. It expires in 24 hours.
              </p>
              <button
                onClick={() => setStatus("idle")}
                className="mt-6 text-sm text-blue-600 hover:underline dark:text-blue-400"
              >
                Use a different email
              </button>
            </div>
          ) : (
            <>
              <h2 className="text-xl font-semibold text-slate-900 dark:text-white mb-1">
                Sign in
              </h2>
              <p className="text-sm text-slate-500 dark:text-slate-400 mb-6">
                Enter your work email and we&apos;ll send you a login link.
              </p>

              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label
                    htmlFor="email"
                    className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
                  >
                    Work email
                  </label>
                  <input
                    id="email"
                    type="email"
                    autoComplete="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@company.com"
                    className="w-full px-3 py-2.5 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                  />
                </div>

                {status === "error" && (
                  <p className="text-sm text-red-600 dark:text-red-400">{errorMsg}</p>
                )}

                <button
                  type="submit"
                  disabled={status === "loading"}
                  className="w-full flex items-center justify-center gap-2 py-2.5 px-4 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg font-medium text-sm transition-colors"
                >
                  {status === "loading" ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <>
                      Send login link
                      <ArrowRight className="h-4 w-4" />
                    </>
                  )}
                </button>
              </form>
            </>
          )}
        </div>

        <p className="text-center text-xs text-slate-400 mt-6">
          No password required · Link expires in 24 hours
        </p>
      </div>
    </div>
  );
}
