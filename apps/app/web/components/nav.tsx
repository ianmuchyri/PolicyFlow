"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Shield, FileText, LayoutDashboard, LogOut } from "lucide-react";
import { clearToken, getTokenPayload, isAnyAdmin } from "@/lib/auth";

export function Nav() {
  const pathname = usePathname();
  const router = useRouter();
  const payload = getTokenPayload();

  function handleLogout() {
    clearToken();
    router.push("/");
  }

  const navItems = [
    { href: "/policies", label: "Policies", icon: FileText },
    ...(isAnyAdmin()
      ? [{ href: "/admin", label: "Admin", icon: LayoutDashboard }]
      : []),
  ];

  return (
    <nav className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          {/* Logo */}
          <div className="flex items-center gap-2">
            <Shield className="h-7 w-7 text-blue-600" />
            <span className="text-lg font-semibold text-slate-900 dark:text-white">
              PolicyFlow
            </span>
          </div>

          {/* Nav links */}
          <div className="flex items-center gap-1">
            {navItems.map(({ href, label, icon: Icon }) => (
              <Link
                key={href}
                href={href}
                className={[
                  "flex items-center gap-1.5 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                  pathname === href
                    ? "bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
                    : "text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-700",
                ].join(" ")}
              >
                <Icon className="h-4 w-4" />
                {label}
              </Link>
            ))}
          </div>

          {/* User + logout */}
          <div className="flex items-center gap-3">
            {payload && (
              <span className="text-sm text-slate-500 dark:text-slate-400 hidden sm:block">
                {payload.email}
                {payload.role === "SuperAdmin" && (
                  <span className="ml-1.5 inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300">
                    Super Admin
                  </span>
                )}
                {payload.role === "DeptAdmin" && (
                  <span className="ml-1.5 inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300">
                    Dept Admin
                  </span>
                )}
                {payload.role === "Staff" && (
                  <span className="ml-1.5 inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300">
                    Staff
                  </span>
                )}
              </span>
            )}
            <button
              onClick={handleLogout}
              className="flex items-center gap-1.5 text-sm text-slate-500 hover:text-red-600 dark:text-slate-400 dark:hover:text-red-400 transition-colors px-2 py-1 rounded"
            >
              <LogOut className="h-4 w-4" />
              <span className="hidden sm:block">Logout</span>
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
}
