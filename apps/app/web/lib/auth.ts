const TOKEN_KEY = "pf_token";

export type Role = "SuperAdmin" | "DeptAdmin" | "Staff";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

export function isAuthenticated(): boolean {
  const token = getToken();
  if (!token) return false;
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    return payload.exp * 1000 > Date.now();
  } catch {
    return false;
  }
}

export interface TokenPayload {
  sub: string;
  email: string;
  role: Role;
}

export function getTokenPayload(): TokenPayload | null {
  const token = getToken();
  if (!token) return null;
  try {
    return JSON.parse(atob(token.split(".")[1]));
  } catch {
    return null;
  }
}

export function isSuperAdmin(): boolean {
  return getTokenPayload()?.role === "SuperAdmin";
}

export function isDeptAdmin(): boolean {
  return getTokenPayload()?.role === "DeptAdmin";
}

export function isAnyAdmin(): boolean {
  const role = getTokenPayload()?.role;
  return role === "SuperAdmin" || role === "DeptAdmin";
}
