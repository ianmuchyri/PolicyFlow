"use client";

import { Suspense, useEffect, useState, useCallback } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  FileText,
  Search,
  ChevronLeft,
  CheckCircle,
  Clock,
  AlertCircle,
  Archive,
  Loader2,
  CheckCheck,
  ChevronDown,
  ChevronUp,
  Globe,
  Building2,
} from "lucide-react";
import { Nav } from "@/components/nav";
import { isAuthenticated } from "@/lib/auth";
import {
  listPolicies,
  listDepartments,
  getPolicy,
  getPolicyVersions,
  acknowledgePolicy,
  type Policy,
  type PolicyDetail,
  type PolicyVersion,
  type Department,
} from "@/lib/api";
import MarkdownRenderer from "@/components/markdown-renderer";

const STATUS_META: Record<
  string,
  { label: string; icon: React.ElementType; classes: string }
> = {
  Draft: {
    label: "Draft",
    icon: Clock,
    classes: "bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300",
  },
  Review: {
    label: "In Review",
    icon: AlertCircle,
    classes: "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300",
  },
  Published: {
    label: "Published",
    icon: CheckCircle,
    classes: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300",
  },
  Archived: {
    label: "Archived",
    icon: Archive,
    classes: "bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400",
  },
};

function StatusBadge({ status }: { status: string }) {
  const meta = STATUS_META[status] ?? STATUS_META.Draft;
  const Icon = meta.icon;
  return (
    <span
      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${meta.classes}`}
    >
      <Icon className="h-3 w-3" />
      {meta.label}
    </span>
  );
}

function VisibilityBadge({ policy }: { policy: Policy }) {
  if (policy.visibility_type === "department") {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300">
        <Building2 className="h-3 w-3" />
        {policy.department_name ?? "Department"}
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
      <Globe className="h-3 w-3" />
      Organization-wide
    </span>
  );
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// ─── Policy List View ──────────────────────────────────────────────────────

function PolicyList({ onSelect }: { onSelect: (id: string) => void }) {
  const [policies, setPolicies] = useState<(Policy & { acknowledged: boolean })[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [deptFilter, setDeptFilter] = useState<string>("all");

  useEffect(() => {
    Promise.all([listPolicies(), listDepartments()])
      .then(([p, d]) => {
        setPolicies(p);
        setDepartments(d);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const filtered = policies.filter((p) => {
    const matchSearch =
      p.title.toLowerCase().includes(search.toLowerCase()) ||
      (p.department_name ?? p.department).toLowerCase().includes(search.toLowerCase());
    const matchStatus = statusFilter === "all" || p.status === statusFilter;
    const matchDept = deptFilter === "all" || p.department_id === deptFilter;
    return matchSearch && matchStatus && matchDept;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  if (error) {
    return <div className="text-center py-16 text-red-500">{error}</div>;
  }

  return (
    <div>
      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            type="search"
            placeholder="Search policies…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-2 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="all">All statuses</option>
          <option value="Published">Published</option>
          <option value="Draft">Draft</option>
          <option value="Review">In Review</option>
          <option value="Archived">Archived</option>
        </select>
        {departments.length > 0 && (
          <select
            value={deptFilter}
            onChange={(e) => setDeptFilter(e.target.value)}
            className="px-3 py-2 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All departments</option>
            {departments.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* List */}
      {filtered.length === 0 ? (
        <div className="text-center py-16 text-slate-400">
          <FileText className="h-10 w-10 mx-auto mb-3 opacity-40" />
          <p>No policies found</p>
        </div>
      ) : (
        <div className="divide-y divide-slate-200 dark:divide-slate-700 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
          {filtered.map((p) => (
            <button
              key={p.id}
              onClick={() => onSelect(p.id)}
              className="w-full flex items-center justify-between px-5 py-4 bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-750 transition-colors text-left group"
            >
              <div className="flex items-start gap-3 min-w-0">
                <FileText className="h-5 w-5 text-slate-400 mt-0.5 shrink-0" />
                <div className="min-w-0">
                  <p className="font-medium text-slate-900 dark:text-white group-hover:text-blue-600 transition-colors">
                    {p.title}
                  </p>
                  {(p.department_name ?? p.department) && (
                    <p className="text-xs text-slate-500 mt-0.5">
                      {p.department_name ?? p.department}
                    </p>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-3 ml-4 shrink-0">
                {p.acknowledged && p.status === "Published" && (
                  <span className="inline-flex items-center gap-1 text-xs text-green-600 dark:text-green-400">
                    <CheckCheck className="h-3.5 w-3.5" />
                    Acknowledged
                  </span>
                )}
                <StatusBadge status={p.status} />
                <ChevronLeft className="h-4 w-4 text-slate-400 rotate-180" />
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Policy Detail View ────────────────────────────────────────────────────

function PolicyDetailView({
  policyId,
  onBack,
}: {
  policyId: string;
  onBack: () => void;
}) {
  const [detail, setDetail] = useState<PolicyDetail | null>(null);
  const [versions, setVersions] = useState<PolicyVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [acking, setAcking] = useState(false);
  const [acknowledged, setAcknowledged] = useState(false);
  const [showVersions, setShowVersions] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(() => {
    Promise.all([getPolicy(policyId), getPolicyVersions(policyId)])
      .then(([d, v]) => {
        setDetail(d);
        setAcknowledged(d.acknowledged);
        setVersions(v);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [policyId]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleAcknowledge() {
    setAcking(true);
    try {
      await acknowledgePolicy(policyId);
      setAcknowledged(true);
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Error");
    } finally {
      setAcking(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  if (error || !detail) {
    return <div className="text-center py-16 text-red-500">{error || "Not found"}</div>;
  }

  const { policy, current_version } = detail;

  return (
    <div>
      {/* Back button + header */}
      <div className="mb-6">
        <button
          onClick={onBack}
          className="flex items-center gap-1 text-sm text-slate-500 hover:text-slate-900 dark:hover:text-white mb-4 transition-colors"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to policies
        </button>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">
              {policy.title}
            </h1>
            <div className="flex flex-wrap items-center gap-2 mt-2">
              <StatusBadge status={policy.status} />
              <VisibilityBadge policy={policy} />
              {current_version && (
                <span className="text-sm text-slate-500">
                  {current_version.version_string}
                </span>
              )}
            </div>
          </div>

          {/* Acknowledge button */}
          {policy.status === "Published" && current_version && (
            <div className="shrink-0">
              {acknowledged ? (
                <div className="flex items-center gap-2 text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-4 py-2 rounded-lg">
                  <CheckCheck className="h-5 w-5" />
                  <span className="font-medium text-sm">Acknowledged</span>
                </div>
              ) : (
                <button
                  onClick={handleAcknowledge}
                  disabled={acking}
                  className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg text-sm font-medium transition-colors"
                >
                  {acking ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <CheckCheck className="h-4 w-4" />
                  )}
                  Acknowledge Policy
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Content */}
      {current_version ? (
        <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 mb-4">
          <MarkdownRenderer content={current_version.content} />
        </div>
      ) : (
        <div className="text-center py-12 text-slate-400">
          <p>No content yet.</p>
        </div>
      )}

      {/* Version history */}
      {versions.length > 0 && (
        <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
          <button
            onClick={() => setShowVersions((v) => !v)}
            className="w-full flex items-center justify-between px-5 py-4 text-sm font-medium text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-750 transition-colors"
          >
            <span>Version history ({versions.length})</span>
            {showVersions ? (
              <ChevronUp className="h-4 w-4" />
            ) : (
              <ChevronDown className="h-4 w-4" />
            )}
          </button>
          {showVersions && (
            <div className="divide-y divide-slate-100 dark:divide-slate-700">
              {versions.map((v) => (
                <div key={v.id} className="px-5 py-3 flex items-center justify-between">
                  <div>
                    <span className="text-sm font-medium text-slate-900 dark:text-white">
                      {v.version_string}
                    </span>
                    {v.changelog && (
                      <p className="text-xs text-slate-500 mt-0.5">{v.changelog}</p>
                    )}
                  </div>
                  <span className="text-xs text-slate-400">{formatDate(v.created_at)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ─── Page inner (needs Suspense for useSearchParams) ──────────────────────

function PoliciesInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const selectedId = searchParams.get("id");

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/");
    }
  }, [router]);

  function handleSelect(id: string) {
    router.push(`/policies?id=${id}`);
  }

  function handleBack() {
    router.push("/policies");
  }

  return (
    <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {selectedId ? (
        <PolicyDetailView policyId={selectedId} onBack={handleBack} />
      ) : (
        <>
          <div className="mb-6">
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Policies</h1>
            <p className="text-slate-500 dark:text-slate-400 mt-1 text-sm">
              Review and acknowledge company policies
            </p>
          </div>
          <PolicyList onSelect={handleSelect} />
        </>
      )}
    </main>
  );
}

// ─── Page ──────────────────────────────────────────────────────────────────

export default function PoliciesPage() {
  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <Nav />
      <Suspense
        fallback={
          <main className="flex items-center justify-center py-24">
            <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
          </main>
        }
      >
        <PoliciesInner />
      </Suspense>
    </div>
  );
}
