"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import {
  Users,
  FileText,
  CheckCircle,
  PlusCircle,
  Loader2,
  RefreshCw,
  X,
} from "lucide-react";
import { Nav } from "@/components/nav";
import { isAuthenticated, getTokenPayload } from "@/lib/auth";
import {
  getAdminStats,
  listUsers,
  listPolicies,
  createUser,
  createPolicy,
  updatePolicy,
  createPolicyVersion,
  type AdminStats,
  type User,
  type Policy,
  type PolicyStatus,
} from "@/lib/api";

// ─── Stat Card ─────────────────────────────────────────────────────────────

function StatCard({
  label,
  value,
  icon: Icon,
  color,
}: {
  label: string;
  value: number;
  icon: React.ElementType;
  color: string;
}) {
  return (
    <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-5 flex items-center gap-4">
      <div className={`p-3 rounded-lg ${color}`}>
        <Icon className="h-5 w-5" />
      </div>
      <div>
        <p className="text-2xl font-bold text-slate-900 dark:text-white">{value}</p>
        <p className="text-xs text-slate-500 dark:text-slate-400">{label}</p>
      </div>
    </div>
  );
}

// ─── Modal ─────────────────────────────────────────────────────────────────

function Modal({
  title,
  onClose,
  children,
}: {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
      <div className="w-full max-w-md bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-slate-900 dark:text-white">{title}</h3>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-700 dark:hover:text-white"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
        {label}
      </label>
      {children}
    </div>
  );
}

const inputClass =
  "w-full px-3 py-2 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500";

// ─── Create User Modal ─────────────────────────────────────────────────────

function CreateUserModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const [form, setForm] = useState({ email: "", name: "", role: "Staff" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await createUser(form);
      onCreated();
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error creating user");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal title="Add User" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Full name">
          <input
            className={inputClass}
            required
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            placeholder="Jane Smith"
          />
        </Field>
        <Field label="Email">
          <input
            type="email"
            className={inputClass}
            required
            value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
            placeholder="jane@company.com"
          />
        </Field>
        <Field label="Role">
          <select
            className={inputClass}
            value={form.role}
            onChange={(e) => setForm({ ...form, role: e.target.value })}
          >
            <option value="Staff">Staff</option>
            <option value="Admin">Admin</option>
          </select>
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <p className="text-xs text-slate-500">
          A welcome email with a login link will be sent to this address.
        </p>
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className="flex-1 py-2 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-2 transition-colors"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Create User
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ─── Create Policy Modal ───────────────────────────────────────────────────

function CreatePolicyModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const [form, setForm] = useState({
    title: "",
    department: "",
    content: "",
    version_string: "v1.0.0",
    changelog: "Initial version",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const policy = await createPolicy({ title: form.title, department: form.department });
      if (form.content) {
        await createPolicyVersion(policy.id, {
          content: form.content,
          version_string: form.version_string,
          changelog: form.changelog,
        });
      }
      onCreated();
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error creating policy");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal title="New Policy" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Title">
          <input className={inputClass} required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="Employee Code of Conduct" />
        </Field>
        <Field label="Department">
          <input className={inputClass} value={form.department} onChange={(e) => setForm({ ...form, department: e.target.value })} placeholder="Human Resources" />
        </Field>
        <Field label="Initial content">
          <textarea className={inputClass} rows={5} value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} placeholder="Policy content (Markdown supported)" />
        </Field>
        <div className="grid grid-cols-2 gap-3">
          <Field label="Version">
            <input className={inputClass} value={form.version_string} onChange={(e) => setForm({ ...form, version_string: e.target.value })} />
          </Field>
          <Field label="Changelog">
            <input className={inputClass} value={form.changelog} onChange={(e) => setForm({ ...form, changelog: e.target.value })} />
          </Field>
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className="flex-1 py-2 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
          <button type="submit" disabled={loading} className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-2 transition-colors">
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Create Policy
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ─── Update Status Modal ───────────────────────────────────────────────────

function UpdateStatusModal({
  policy,
  onClose,
  onUpdated,
}: {
  policy: Policy;
  onClose: () => void;
  onUpdated: () => void;
}) {
  const [status, setStatus] = useState<PolicyStatus>(policy.status);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      await updatePolicy(policy.id, { status });
      onUpdated();
      onClose();
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal title={`Update: ${policy.title}`} onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Status">
          <select className={inputClass} value={status} onChange={(e) => setStatus(e.target.value as PolicyStatus)}>
            <option value="Draft">Draft</option>
            <option value="Review">Review</option>
            <option value="Published">Published</option>
            <option value="Archived">Archived</option>
          </select>
        </Field>
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className="flex-1 py-2 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
          <button type="submit" disabled={loading} className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-2 transition-colors">
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Save
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ─── Add Version Modal ─────────────────────────────────────────────────────

function AddVersionModal({
  policy,
  onClose,
  onCreated,
}: {
  policy: Policy;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [form, setForm] = useState({ content: "", version_string: "", changelog: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await createPolicyVersion(policy.id, form);
      onCreated();
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal title={`New Version: ${policy.title}`} onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <Field label="Version string">
            <input className={inputClass} required value={form.version_string} onChange={(e) => setForm({ ...form, version_string: e.target.value })} placeholder="v1.1.0" />
          </Field>
          <Field label="Changelog">
            <input className={inputClass} value={form.changelog} onChange={(e) => setForm({ ...form, changelog: e.target.value })} placeholder="Updated section 3" />
          </Field>
        </div>
        <Field label="Content">
          <textarea className={inputClass} rows={6} required value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} placeholder="Policy content…" />
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className="flex-1 py-2 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
          <button type="submit" disabled={loading} className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-2 transition-colors">
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Publish Version
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ─── Page ──────────────────────────────────────────────────────────────────

type ModalState =
  | { type: "none" }
  | { type: "create-user" }
  | { type: "create-policy" }
  | { type: "update-status"; policy: Policy }
  | { type: "add-version"; policy: Policy };

export default function AdminPage() {
  const router = useRouter();
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<"overview" | "users" | "policies">("overview");
  const [modal, setModal] = useState<ModalState>({ type: "none" });

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/");
      return;
    }
    const payload = getTokenPayload();
    if (payload?.role !== "Admin") {
      router.replace("/policies");
    }
  }, [router]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [s, u, p] = await Promise.all([getAdminStats(), listUsers(), listPolicies()]);
      setStats(s);
      setUsers(u);
      setPolicies(p);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const STATUS_COLORS: Record<string, string> = {
    Draft: "bg-slate-100 text-slate-600",
    Review: "bg-amber-100 text-amber-700",
    Published: "bg-green-100 text-green-700",
    Archived: "bg-slate-200 text-slate-500",
  };

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <Nav />

      <main className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Admin Dashboard</h1>
            <p className="text-slate-500 dark:text-slate-400 text-sm mt-1">Manage users and policies</p>
          </div>
          <button onClick={loadData} className="flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-900 dark:hover:text-white px-3 py-1.5 border border-slate-300 dark:border-slate-600 rounded-lg hover:bg-white dark:hover:bg-slate-700 transition-colors">
            <RefreshCw className="h-3.5 w-3.5" />
            Refresh
          </button>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 mb-6 border-b border-slate-200 dark:border-slate-700">
          {(["overview", "users", "policies"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={[
                "px-4 py-2 text-sm font-medium capitalize border-b-2 -mb-px transition-colors",
                tab === t
                  ? "border-blue-600 text-blue-600 dark:text-blue-400"
                  : "border-transparent text-slate-500 hover:text-slate-900 dark:hover:text-white",
              ].join(" ")}
            >
              {t}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-24">
            <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
          </div>
        ) : (
          <>
            {/* Overview Tab */}
            {tab === "overview" && stats && (
              <div className="space-y-6">
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                  <StatCard label="Total Users" value={stats.stats.total_users} icon={Users} color="bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400" />
                  <StatCard label="Published Policies" value={stats.stats.published_count} icon={CheckCircle} color="bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400" />
                  <StatCard label="Total Policies" value={stats.stats.total_policies} icon={FileText} color="bg-violet-100 text-violet-600 dark:bg-violet-900/30 dark:text-violet-400" />
                  <StatCard label="Acknowledgements" value={stats.stats.total_acknowledgements} icon={CheckCircle} color="bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400" />
                </div>

                {stats.ack_counts && stats.ack_counts.length > 0 && (
                  <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-5">
                    <h2 className="text-sm font-semibold text-slate-900 dark:text-white mb-3">
                      Acknowledgements by Policy
                    </h2>
                    <div className="space-y-2">
                      {stats.ack_counts.map((a) => (
                        <div key={a.policy_id} className="flex items-center justify-between text-sm">
                          <span className="text-slate-700 dark:text-slate-300">{a.title}</span>
                          <span className="font-medium text-slate-900 dark:text-white">
                            {a.ack_count} <span className="text-slate-400 font-normal">/ {stats.stats.total_users}</span>
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Users Tab */}
            {tab === "users" && (
              <div>
                <div className="flex justify-end mb-4">
                  <button
                    onClick={() => setModal({ type: "create-user" })}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"
                  >
                    <PlusCircle className="h-4 w-4" />
                    Add User
                  </button>
                </div>
                <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-slate-200 dark:border-slate-700">
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Name</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Email</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Role</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Joined</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-700">
                      {users.map((u) => (
                        <tr key={u.id}>
                          <td className="px-5 py-3 font-medium text-slate-900 dark:text-white">{u.name}</td>
                          <td className="px-5 py-3 text-slate-600 dark:text-slate-300">{u.email}</td>
                          <td className="px-5 py-3">
                            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${u.role === "Admin" ? "bg-violet-100 text-violet-700" : "bg-slate-100 text-slate-600"}`}>
                              {u.role}
                            </span>
                          </td>
                          <td className="px-5 py-3 text-slate-500">
                            {new Date(u.created_at).toLocaleDateString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {users.length === 0 && (
                    <div className="text-center py-12 text-slate-400">No users yet</div>
                  )}
                </div>
              </div>
            )}

            {/* Policies Tab */}
            {tab === "policies" && (
              <div>
                <div className="flex justify-end mb-4">
                  <button
                    onClick={() => setModal({ type: "create-policy" })}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"
                  >
                    <PlusCircle className="h-4 w-4" />
                    New Policy
                  </button>
                </div>
                <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-slate-200 dark:border-slate-700">
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Title</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Department</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-700">
                      {policies.map((p) => (
                        <tr key={p.id}>
                          <td className="px-5 py-3 font-medium text-slate-900 dark:text-white">{p.title}</td>
                          <td className="px-5 py-3 text-slate-500">{p.department || "—"}</td>
                          <td className="px-5 py-3">
                            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[p.status] ?? ""}`}>
                              {p.status}
                            </span>
                          </td>
                          <td className="px-5 py-3">
                            <div className="flex gap-2">
                              <button
                                onClick={() => setModal({ type: "update-status", policy: p })}
                                className="text-xs text-blue-600 hover:underline dark:text-blue-400"
                              >
                                Edit status
                              </button>
                              <button
                                onClick={() => setModal({ type: "add-version", policy: p })}
                                className="text-xs text-slate-500 hover:underline dark:text-slate-400"
                              >
                                + Version
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {policies.length === 0 && (
                    <div className="text-center py-12 text-slate-400">No policies yet</div>
                  )}
                </div>
              </div>
            )}
          </>
        )}
      </main>

      {/* Modals */}
      {modal.type === "create-user" && (
        <CreateUserModal onClose={() => setModal({ type: "none" })} onCreated={loadData} />
      )}
      {modal.type === "create-policy" && (
        <CreatePolicyModal onClose={() => setModal({ type: "none" })} onCreated={loadData} />
      )}
      {modal.type === "update-status" && (
        <UpdateStatusModal policy={modal.policy} onClose={() => setModal({ type: "none" })} onUpdated={loadData} />
      )}
      {modal.type === "add-version" && (
        <AddVersionModal policy={modal.policy} onClose={() => setModal({ type: "none" })} onCreated={loadData} />
      )}
    </div>
  );
}
