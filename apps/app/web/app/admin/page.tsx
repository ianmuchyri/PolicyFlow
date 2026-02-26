"use client";

import { useEffect, useState, useCallback, lazy, Suspense } from "react";
import { useRouter } from "next/navigation";
import dynamic from "next/dynamic";
import {
  Users,
  FileText,
  CheckCircle,
  PlusCircle,
  Loader2,
  RefreshCw,
  X,
  Building2,
  Pencil,
  Trash2,
  Globe,
} from "lucide-react";
import { Nav } from "@/components/nav";
import { isAuthenticated, getTokenPayload, isSuperAdmin } from "@/lib/auth";
import {
  getAdminStats,
  getMe,
  listUsers,
  listPolicies,
  listDepartments,
  createUser,
  updateUser,
  deleteUser,
  createDepartment,
  updateDepartment,
  deleteDepartment,
  createPolicy,
  updatePolicy,
  createPolicyVersion,
  type AdminStats,
  type User,
  type Policy,
  type PolicyStatus,
  type Department,
  type VisibilityType,
  type UserRole,
} from "@/lib/api";

// CodeMirror must be loaded client-side only (no SSR).
const MarkdownEditor = dynamic(() => import("@/components/codemirror-editor"), {
  ssr: false,
  loading: () => (
    <div className="h-80 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 flex items-center justify-center text-slate-400 text-sm">
      Loading editor…
    </div>
  ),
});

// ─── Helpers ───────────────────────────────────────────────────────────────

function formatDate(s: string) {
  return new Date(s).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

const ROLE_BADGE: Record<string, string> = {
  SuperAdmin: "bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-300",
  DeptAdmin: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300",
  Staff: "bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300",
};

const ROLE_LABEL: Record<string, string> = {
  SuperAdmin: "Super Admin",
  DeptAdmin: "Dept Admin",
  Staff: "Staff",
};

const STATUS_COLORS: Record<string, string> = {
  Draft: "bg-slate-100 text-slate-600",
  Review: "bg-amber-100 text-amber-700",
  Published: "bg-green-100 text-green-700",
  Archived: "bg-slate-200 text-slate-500",
};

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

// ─── Modal + Field ─────────────────────────────────────────────────────────

function Modal({
  title,
  onClose,
  children,
  wide,
}: {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
  wide?: boolean;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm overflow-y-auto">
      <div
        className={`w-full ${wide ? "max-w-2xl" : "max-w-md"} bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 p-6 my-4`}
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-slate-900 dark:text-white">{title}</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-700 dark:hover:text-white">
            <X className="h-5 w-5" />
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
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

const btnPrimary =
  "flex-1 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-2 transition-colors";

const btnCancel =
  "flex-1 py-2 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors";

// ─── Department Modals ─────────────────────────────────────────────────────

function CreateDepartmentModal({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: () => void;
}) {
  const [form, setForm] = useState({ name: "", description: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await createDepartment(form);
      onCreated();
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error creating department");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal title="Add Department" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Name">
          <input className={inputClass} required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Engineering" />
        </Field>
        <Field label="Description">
          <input className={inputClass} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="Optional description" />
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className={btnCancel}>Cancel</button>
          <button type="submit" disabled={loading} className={btnPrimary}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Create
          </button>
        </div>
      </form>
    </Modal>
  );
}

function EditDepartmentModal({
  dept,
  onClose,
  onUpdated,
}: {
  dept: Department;
  onClose: () => void;
  onUpdated: () => void;
}) {
  const [form, setForm] = useState({ name: dept.name, description: dept.description });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await updateDepartment(dept.id, form);
      onUpdated();
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal title={`Edit: ${dept.name}`} onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Name">
          <input className={inputClass} required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </Field>
        <Field label="Description">
          <input className={inputClass} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className={btnCancel}>Cancel</button>
          <button type="submit" disabled={loading} className={btnPrimary}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Save
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ─── User Modals ───────────────────────────────────────────────────────────

function CreateUserModal({
  departments,
  currentUser,
  onClose,
  onCreated,
}: {
  departments: Department[];
  currentUser: User | null;
  onClose: () => void;
  onCreated: () => void;
}) {
  const isDeptAdminCaller = currentUser?.role === "DeptAdmin";
  const [form, setForm] = useState({
    email: "",
    name: "",
    role: "Staff",
    department_id: isDeptAdminCaller ? (currentUser?.department_id ?? "") : "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await createUser({ ...form, department_id: form.department_id || null });
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
          <input className={inputClass} required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Jane Smith" />
        </Field>
        <Field label="Email">
          <input type="email" className={inputClass} required value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} placeholder="jane@company.com" />
        </Field>
        <Field label="Role">
          <select className={inputClass} value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value as UserRole })}>
            <option value="Staff">Staff</option>
            <option value="DeptAdmin">Dept Admin</option>
            {!isDeptAdminCaller && <option value="SuperAdmin">Super Admin</option>}
          </select>
        </Field>
        <Field label="Department">
          {isDeptAdminCaller ? (
            <div className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700/50 text-slate-700 dark:text-slate-300">
              <Building2 className="h-4 w-4 text-slate-400" />
              {currentUser?.department_name ?? "Your department"}
            </div>
          ) : (
            <select className={inputClass} value={form.department_id} onChange={(e) => setForm({ ...form, department_id: e.target.value })}>
              <option value="">— None —</option>
              {departments.map((d) => (
                <option key={d.id} value={d.id}>{d.name}</option>
              ))}
            </select>
          )}
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <p className="text-xs text-slate-500">A welcome email with a login link will be sent to this address.</p>
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className={btnCancel}>Cancel</button>
          <button type="submit" disabled={loading} className={btnPrimary}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Create User
          </button>
        </div>
      </form>
    </Modal>
  );
}

function EditUserModal({
  user,
  departments,
  onClose,
  onUpdated,
}: {
  user: User;
  departments: Department[];
  onClose: () => void;
  onUpdated: () => void;
}) {
  const [form, setForm] = useState({
    name: user.name,
    email: user.email,
    role: user.role,
    department_id: user.department_id ?? "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await updateUser(user.id, { ...form, department_id: form.department_id || null });
      onUpdated();
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal title={`Edit: ${user.name}`} onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Full name">
          <input className={inputClass} required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </Field>
        <Field label="Email">
          <input type="email" className={inputClass} required value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
        </Field>
        <Field label="Role">
          <select className={inputClass} value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value as UserRole })}>
            <option value="Staff">Staff</option>
            <option value="DeptAdmin">Dept Admin</option>
            <option value="SuperAdmin">Super Admin</option>
          </select>
        </Field>
        <Field label="Department">
          <select className={inputClass} value={form.department_id} onChange={(e) => setForm({ ...form, department_id: e.target.value })}>
            <option value="">— None —</option>
            {departments.map((d) => (
              <option key={d.id} value={d.id}>{d.name}</option>
            ))}
          </select>
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className={btnCancel}>Cancel</button>
          <button type="submit" disabled={loading} className={btnPrimary}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Save
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ─── Policy Modals ─────────────────────────────────────────────────────────

function CreatePolicyModal({
  departments,
  currentUser,
  onClose,
  onCreated,
}: {
  departments: Department[];
  currentUser: User | null;
  onClose: () => void;
  onCreated: () => void;
}) {
  const payload = getTokenPayload();
  const isDeptAdminUser = payload?.role === "DeptAdmin";

  const [form, setForm] = useState({
    title: "",
    visibility_type: isDeptAdminUser ? "department" : "organization" as VisibilityType,
    department_id: isDeptAdminUser ? (currentUser?.department_id ?? "") : "",
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
      const policy = await createPolicy({
        title: form.title,
        visibility_type: form.visibility_type,
        department_id: form.visibility_type === "department" ? (form.department_id || null) : null,
      });
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
    <Modal title="New Policy" onClose={onClose} wide>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Title">
          <input className={inputClass} required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="Employee Code of Conduct" />
        </Field>

        {!isDeptAdminUser && (
          <Field label="Visibility">
            <select className={inputClass} value={form.visibility_type} onChange={(e) => setForm({ ...form, visibility_type: e.target.value as VisibilityType })}>
              <option value="organization">Organization-wide</option>
              <option value="department">Department only</option>
            </select>
          </Field>
        )}

        {form.visibility_type === "department" && (
          <Field label="Department">
            {isDeptAdminUser ? (
              <div className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700/50 text-slate-700 dark:text-slate-300">
                <Building2 className="h-4 w-4 text-slate-400" />
                {currentUser?.department_name ?? "Your department"}
              </div>
            ) : (
              <select className={inputClass} value={form.department_id} onChange={(e) => setForm({ ...form, department_id: e.target.value })}>
                <option value="">— Select department —</option>
                {departments.map((d) => (
                  <option key={d.id} value={d.id}>{d.name}</option>
                ))}
              </select>
            )}
          </Field>
        )}

        <div className="grid grid-cols-2 gap-3">
          <Field label="Version">
            <input className={inputClass} value={form.version_string} onChange={(e) => setForm({ ...form, version_string: e.target.value })} />
          </Field>
          <Field label="Changelog">
            <input className={inputClass} value={form.changelog} onChange={(e) => setForm({ ...form, changelog: e.target.value })} />
          </Field>
        </div>

        <Field label="Initial content (Markdown)">
          <MarkdownEditor value={form.content} onChange={(v) => setForm({ ...form, content: v })} height="280px" />
        </Field>

        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className={btnCancel}>Cancel</button>
          <button type="submit" disabled={loading} className={btnPrimary}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Create Policy
          </button>
        </div>
      </form>
    </Modal>
  );
}

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
        <div className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-700/50 rounded-lg px-3 py-2">
          {policy.visibility_type === "organization" ? (
            <><Globe className="h-4 w-4" /> Organization-wide</>
          ) : (
            <><Building2 className="h-4 w-4" /> {policy.department_name ?? "Department"}</>
          )}
        </div>
        <Field label="Status">
          <select className={inputClass} value={status} onChange={(e) => setStatus(e.target.value as PolicyStatus)}>
            <option value="Draft">Draft</option>
            <option value="Review">Review</option>
            <option value="Published">Published</option>
            <option value="Archived">Archived</option>
          </select>
        </Field>
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className={btnCancel}>Cancel</button>
          <button type="submit" disabled={loading} className={btnPrimary}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Save
          </button>
        </div>
      </form>
    </Modal>
  );
}

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
    <Modal title={`New Version: ${policy.title}`} onClose={onClose} wide>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <Field label="Version string">
            <input className={inputClass} required value={form.version_string} onChange={(e) => setForm({ ...form, version_string: e.target.value })} placeholder="v1.1.0" />
          </Field>
          <Field label="Changelog">
            <input className={inputClass} value={form.changelog} onChange={(e) => setForm({ ...form, changelog: e.target.value })} placeholder="Updated section 3" />
          </Field>
        </div>
        <Field label="Content (Markdown)">
          <MarkdownEditor value={form.content} onChange={(v) => setForm({ ...form, content: v })} height="320px" />
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-3 pt-2">
          <button type="button" onClick={onClose} className={btnCancel}>Cancel</button>
          <button type="submit" disabled={loading} className={btnPrimary}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Publish Version
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ─── Page ──────────────────────────────────────────────────────────────────

type TabType = "overview" | "users" | "policies" | "departments";

type ModalState =
  | { type: "none" }
  | { type: "create-user" }
  | { type: "edit-user"; user: User }
  | { type: "create-policy" }
  | { type: "update-status"; policy: Policy }
  | { type: "add-version"; policy: Policy }
  | { type: "create-dept" }
  | { type: "edit-dept"; dept: Department };

export default function AdminPage() {
  const router = useRouter();
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<TabType>("overview");
  const [modal, setModal] = useState<ModalState>({ type: "none" });
  const [deleteError, setDeleteError] = useState("");

  const superAdmin = isSuperAdmin();
  const currentUserId = getTokenPayload()?.sub;

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/");
      return;
    }
    const payload = getTokenPayload();
    if (payload?.role !== "SuperAdmin" && payload?.role !== "DeptAdmin") {
      router.replace("/policies");
    }
  }, [router]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [s, me, u, p, d] = await Promise.all([
        getAdminStats(),
        getMe(),
        listUsers(),
        listPolicies(),
        listDepartments(),
      ]);
      setStats(s);
      setCurrentUser(me);
      setUsers(u);
      setPolicies(p);
      setDepartments(d);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  async function handleDeleteUser(user: User) {
    if (!confirm(`Delete user "${user.name}"? This cannot be undone.`)) return;
    setDeleteError("");
    try {
      await deleteUser(user.id);
      await loadData();
    } catch (err: unknown) {
      setDeleteError(err instanceof Error ? err.message : "Error deleting user");
    }
  }

  async function handleDeleteDept(dept: Department) {
    if (!confirm(`Delete department "${dept.name}"?`)) return;
    setDeleteError("");
    try {
      await deleteDepartment(dept.id);
      await loadData();
    } catch (err: unknown) {
      setDeleteError(err instanceof Error ? err.message : "Error: " + (err instanceof Error ? err.message : "unknown error"));
    }
  }

  const tabs: { key: TabType; label: string }[] = [
    { key: "overview", label: "Overview" },
    { key: "users", label: "Users" },
    { key: "policies", label: "Policies" },
    ...(superAdmin ? [{ key: "departments" as TabType, label: "Departments" }] : []),
  ];

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <Nav />

      <main className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Admin Dashboard</h1>
            <p className="text-slate-500 dark:text-slate-400 text-sm mt-1">Manage users and policies</p>
          </div>
          <button
            onClick={loadData}
            className="flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-900 dark:hover:text-white px-3 py-1.5 border border-slate-300 dark:border-slate-600 rounded-lg hover:bg-white dark:hover:bg-slate-700 transition-colors"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            Refresh
          </button>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 mb-6 border-b border-slate-200 dark:border-slate-700">
          {tabs.map(({ key, label }) => (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={[
                "px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
                tab === key
                  ? "border-blue-600 text-blue-600 dark:text-blue-400"
                  : "border-transparent text-slate-500 hover:text-slate-900 dark:hover:text-white",
              ].join(" ")}
            >
              {label}
            </button>
          ))}
        </div>

        {deleteError && (
          <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-400">
            {deleteError}
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-24">
            <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
          </div>
        ) : (
          <>
            {/* ── Overview Tab ──────────────────────────────────────── */}
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
                    <h2 className="text-sm font-semibold text-slate-900 dark:text-white mb-3">Acknowledgements by Policy</h2>
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

            {/* ── Departments Tab ───────────────────────────────────── */}
            {tab === "departments" && superAdmin && (
              <div>
                <div className="flex justify-end mb-4">
                  <button
                    onClick={() => setModal({ type: "create-dept" })}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"
                  >
                    <PlusCircle className="h-4 w-4" />
                    Add Department
                  </button>
                </div>
                <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-slate-200 dark:border-slate-700">
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Name</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Description</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-700">
                      {departments.map((d) => (
                        <tr key={d.id}>
                          <td className="px-5 py-3 font-medium text-slate-900 dark:text-white">
                            <div className="flex items-center gap-2">
                              <Building2 className="h-4 w-4 text-slate-400" />
                              {d.name}
                            </div>
                          </td>
                          <td className="px-5 py-3 text-slate-500 dark:text-slate-400">{d.description || "—"}</td>
                          <td className="px-5 py-3">
                            <div className="flex gap-2">
                              <button onClick={() => setModal({ type: "edit-dept", dept: d })} className="p-1.5 text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded transition-colors">
                                <Pencil className="h-4 w-4" />
                              </button>
                              <button onClick={() => handleDeleteDept(d)} className="p-1.5 text-slate-400 hover:text-red-600 dark:hover:text-red-400 rounded transition-colors">
                                <Trash2 className="h-4 w-4" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {departments.length === 0 && (
                    <div className="text-center py-12 text-slate-400">No departments yet</div>
                  )}
                </div>
              </div>
            )}

            {/* ── Users Tab ─────────────────────────────────────────── */}
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
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Department</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Joined</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-700">
                      {users.map((u) => (
                        <tr key={u.id}>
                          <td className="px-5 py-3 font-medium text-slate-900 dark:text-white">{u.name}</td>
                          <td className="px-5 py-3 text-slate-600 dark:text-slate-300">{u.email}</td>
                          <td className="px-5 py-3">
                            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${ROLE_BADGE[u.role] ?? ROLE_BADGE.Staff}`}>
                              {ROLE_LABEL[u.role] ?? u.role}
                            </span>
                          </td>
                          <td className="px-5 py-3 text-slate-500">{u.department_name ?? "—"}</td>
                          <td className="px-5 py-3 text-slate-500">{formatDate(u.created_at)}</td>
                          <td className="px-5 py-3">
                            {superAdmin && (
                              <div className="flex gap-2">
                                <button onClick={() => setModal({ type: "edit-user", user: u })} className="p-1.5 text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded transition-colors">
                                  <Pencil className="h-4 w-4" />
                                </button>
                                {u.id !== currentUserId && (
                                  <button onClick={() => handleDeleteUser(u)} className="p-1.5 text-slate-400 hover:text-red-600 dark:hover:text-red-400 rounded transition-colors">
                                    <Trash2 className="h-4 w-4" />
                                  </button>
                                )}
                              </div>
                            )}
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

            {/* ── Policies Tab ──────────────────────────────────────── */}
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
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Visibility</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
                        <th className="text-left px-5 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-700">
                      {policies.map((p) => (
                        <tr key={p.id}>
                          <td className="px-5 py-3 font-medium text-slate-900 dark:text-white">{p.title}</td>
                          <td className="px-5 py-3">
                            {p.visibility_type === "department" ? (
                              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300">
                                <Building2 className="h-3 w-3" />
                                {p.department_name ?? "Dept"}
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
                                <Globe className="h-3 w-3" />
                                Org-wide
                              </span>
                            )}
                          </td>
                          <td className="px-5 py-3">
                            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[p.status] ?? ""}`}>
                              {p.status}
                            </span>
                          </td>
                          <td className="px-5 py-3">
                            {(superAdmin || p.visibility_type === "department") && (
                              <div className="flex gap-3">
                                <button onClick={() => setModal({ type: "update-status", policy: p })} className="text-xs text-blue-600 hover:underline dark:text-blue-400">
                                  Edit status
                                </button>
                                <button onClick={() => setModal({ type: "add-version", policy: p })} className="text-xs text-slate-500 hover:underline dark:text-slate-400">
                                  + Version
                                </button>
                              </div>
                            )}
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
      {modal.type === "create-dept" && (
        <CreateDepartmentModal onClose={() => setModal({ type: "none" })} onCreated={loadData} />
      )}
      {modal.type === "edit-dept" && (
        <EditDepartmentModal dept={modal.dept} onClose={() => setModal({ type: "none" })} onUpdated={loadData} />
      )}
      {modal.type === "create-user" && (
        <CreateUserModal departments={departments} currentUser={currentUser} onClose={() => setModal({ type: "none" })} onCreated={loadData} />
      )}
      {modal.type === "edit-user" && (
        <EditUserModal user={modal.user} departments={departments} onClose={() => setModal({ type: "none" })} onUpdated={loadData} />
      )}
      {modal.type === "create-policy" && (
        <CreatePolicyModal departments={departments} currentUser={currentUser} onClose={() => setModal({ type: "none" })} onCreated={loadData} />
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
