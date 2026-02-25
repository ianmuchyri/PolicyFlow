import { getToken } from "./auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken();
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }));
    throw new Error(err.message ?? `HTTP ${res.status}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

// ─── Auth ──────────────────────────────────────────────────────────────────

export function requestMagicLink(email: string) {
  return request<{ message: string }>("/api/magic-link", {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export function getMe() {
  return request<User>("/api/me");
}

// ─── Policies ─────────────────────────────────────────────────────────────

export type PolicyStatus = "Draft" | "Review" | "Published" | "Archived";

export interface Policy {
  id: string;
  title: string;
  current_version_id?: string;
  status: PolicyStatus;
  department: string;
  created_at: string;
  acknowledged?: boolean;
}

export interface PolicyVersion {
  id: string;
  policy_id: string;
  content: string;
  version_string: string;
  changelog: string;
  created_at: string;
}

export interface PolicyDetail {
  policy: Policy;
  current_version: PolicyVersion | null;
  acknowledged: boolean;
}

export function listPolicies() {
  return request<(Policy & { acknowledged: boolean })[]>("/api/policies");
}

export function getPolicy(id: string) {
  return request<PolicyDetail>(`/api/policies/${id}`);
}

export function getPolicyVersions(id: string) {
  return request<PolicyVersion[]>(`/api/policies/${id}/versions`);
}

export function acknowledgePolicy(id: string) {
  return request<{ id: string }>(`/api/policies/${id}/acknowledge`, {
    method: "POST",
  });
}

export function createPolicy(data: { title: string; department: string }) {
  return request<Policy>("/api/policies", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function updatePolicy(
  id: string,
  data: { title?: string; status?: PolicyStatus; department?: string }
) {
  return request<Policy>(`/api/policies/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export function createPolicyVersion(
  policyId: string,
  data: { content: string; version_string: string; changelog: string }
) {
  return request<PolicyVersion>(`/api/policies/${policyId}/versions`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

// ─── Users ─────────────────────────────────────────────────────────────────

export interface User {
  id: string;
  email: string;
  name: string;
  role: "Admin" | "Staff";
  created_by?: string;
  created_at: string;
}

export function listUsers() {
  return request<User[]>("/api/users");
}

export function createUser(data: { email: string; name: string; role: string }) {
  return request<User>("/api/users", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

// ─── Admin ─────────────────────────────────────────────────────────────────

export interface AdminStats {
  stats: {
    total_users: number;
    total_policies: number;
    published_count: number;
    draft_count: number;
    review_count: number;
    archived_count: number;
    total_acknowledgements: number;
  };
  ack_counts: { policy_id: string; title: string; ack_count: number }[];
}

export function getAdminStats() {
  return request<AdminStats>("/api/admin/stats");
}
