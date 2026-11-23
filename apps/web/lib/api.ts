// Hand-written typed client for the control-plane API, kept in sync with
// api/openapi.yaml by hand for now — codegen (openapi-typescript) is not
// wired up yet (docs/PROGRESS.md, Taskfile.yml's `gen` task).
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  status: number;
  detail?: string;

  constructor(status: number, title: string, detail?: string) {
    super(title);
    this.status = status;
    this.detail = detail;
  }
}

interface Problem {
  title?: string;
  detail?: string;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    // The session lives in an httpOnly cookie, never in localStorage
    // (docs/MASTER_PLAN.md §6.2) — this is what makes it come along.
    credentials: "include",
    headers: {
      ...(init?.body ? { "Content-Type": "application/json" } : {}),
      ...init?.headers,
    },
  });

  if (!res.ok) {
    let title = res.statusText || `request failed (${res.status})`;
    let detail: string | undefined;
    try {
      const problem = (await res.json()) as Problem;
      title = problem.title ?? title;
      detail = problem.detail;
    } catch {
      // Not a problem+json body (e.g. a plain 401/403 with no body).
    }
    throw new ApiError(res.status, title, detail);
  }

  if (res.status === 204 || res.headers.get("Content-Length") === "0") {
    return undefined as T;
  }
  return (await res.json()) as T;
}

export type Role = "owner" | "admin" | "analyst" | "viewer";

export interface Me {
  user_id: string;
  email: string;
  workspace_id: string;
  role: Role;
  approver: boolean;
}

export interface LoginResult {
  user_id: string;
  email: string;
  role: Role;
}

export type TwinState =
  | "requested"
  | "imaging"
  | "provisioning"
  | "applying"
  | "verifying"
  | "ready"
  | "running"
  | "stopped"
  | "failed"
  | "destroyed";

export interface Twin {
  id: string;
  workspace_id: string;
  snapshot_id: string;
  provider: string;
  state: TwinState;
  fidelity_score: number | null;
  created_at: string;
  updated_at: string;
}

export const api = {
  setupStatus: () => request<{ initialized: boolean }>("/api/v1/auth/setup-status"),

  setup: (email: string, password: string) =>
    request<void>("/api/v1/auth/setup", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  login: (email: string, password: string) =>
    request<LoginResult>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  logout: () => request<void>("/api/v1/auth/logout", { method: "POST" }),

  me: () => request<Me>("/api/v1/me"),

  listTwins: () => request<{ twins: Twin[] }>("/api/v1/twins"),

  createTwin: () => request<Twin>("/api/v1/twins", { method: "POST", body: JSON.stringify({}) }),

  getTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}`),

  startTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}/start`, { method: "POST" }),

  stopTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}/stop`, { method: "POST" }),

  destroyTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}/destroy`, { method: "POST" }),

  // Not a plain request(): this is a stream URL for the caller to hand to
  // an EventSource, not a JSON response to decode.
  twinEventsUrl: (id: string) => `${API_BASE}/api/v1/twins/${id}/events`,
};
