// Hand-written typed client for the control-plane API, kept in sync with
// api/openapi.yaml by hand for now — codegen (openapi-typescript) is not
// wired up yet (docs/PROGRESS.md, Taskfile.yml's `gen` task).
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  status: number;
  detail?: string;
  // The parsed error body, for endpoints whose error response carries data
  // (e.g. a refused emulation run returns the run with its preflight).
  body?: unknown;

  constructor(status: number, title: string, detail?: string, body?: unknown) {
    super(title);
    this.status = status;
    this.detail = detail;
    this.body = body;
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
    let body: unknown;
    try {
      body = await res.json();
      const problem = body as Problem;
      title = problem.title ?? title;
      detail = problem.detail;
    } catch {
      // Not a JSON body (e.g. a plain 401/403 with no body).
    }
    throw new ApiError(res.status, title, detail, body);
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

export interface Asset {
  id: string;
  workspace_id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface SnapshotSummary {
  id: string;
  asset_id: string;
  version: number;
  sha256: string;
  created_at: string;
}

export type FindingStatus = "pass" | "fail" | "unknown" | "not_applicable";
export type Severity = "high" | "medium" | "low";

export interface Finding {
  control_id: string;
  title: string;
  category: string;
  severity: Severity;
  status: FindingStatus;
  observed: unknown;
  expected: string;
  reason: string;
  remediation: string;
  techniques: string[];
}

export interface Assessment {
  id: string;
  snapshot_id: string;
  twin_id: string | null;
  engine: string;
  engine_version: string;
  score: number | null;
  counts: { passed: number; failed: number; unknown: number; not_applicable: number };
  findings?: Finding[];
  created_at: string;
}

export interface AssessmentChange {
  control_id: string;
  title: string;
  severity: Severity;
  before: FindingStatus | "";
  after: FindingStatus | "";
  direction: "regressed" | "improved" | "changed";
}

export interface AssessmentDiff {
  before: Assessment;
  after: Assessment;
  changes: AssessmentChange[];
}

export type RiskTier = "benign" | "intrusive";

export interface CatalogTest {
  guid: string;
  technique_id: string;
  technique_name: string;
  name: string;
  description: string;
  executor: string;
  elevation_required: boolean;
  command: string;
  cleanup?: string;
  risk_tier: RiskTier;
  curation_note?: string;
}

export interface Scenario {
  id: string;
  name: string;
  description: string;
  test_guids: string[];
}

export interface AttackCatalog {
  source: string;
  repository: string;
  commit: string;
  commit_date: string;
  license: string;
  verified: boolean;
  verification_detail: string;
  authorisation_statement: string;
  limits: { max_tests_per_run: number; max_run_seconds: number; max_test_seconds: number };
  tests: CatalogTest[];
  scenarios: Scenario[];
}

export interface PreflightCheck {
  id: string;
  name: string;
  passed: boolean;
  detail: string;
}

export type RunStatus = "preflight_failed" | "running" | "completed" | "aborted" | "failed";
export type Outcome = "blocked" | "detected" | "logged" | "missed" | "not_executed" | "error" | "skipped";

export interface Execution {
  test_guid: string;
  technique_id: string;
  test_name: string;
  risk_tier: RiskTier;
  outcome: Outcome;
  detail: string;
  started_at: string;
  finished_at: string;
}

export interface EmulationRun {
  id: string;
  twin_id: string;
  scenario_id: string;
  test_guids: string[];
  status: RunStatus;
  detail: string;
  preflight: PreflightCheck[];
  catalog_source: string;
  catalog_commit: string;
  provider: string;
  authorised_by: string;
  authorisation_statement: string;
  executions: Execution[];
  created_at: string;
  finished_at: string | null;
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

  listAssets: () => request<{ assets: Asset[] }>("/api/v1/assets"),

  createAsset: (name: string) =>
    request<Asset>("/api/v1/assets", { method: "POST", body: JSON.stringify({ name }) }),

  getAsset: (id: string) => request<Asset>(`/api/v1/assets/${id}`),

  listSnapshots: (assetId: string) =>
    request<{ snapshots: SnapshotSummary[] }>(`/api/v1/assets/${assetId}/snapshots`),

  // The collector's output file is sent as-is (a Blob keeps the exact
  // bytes), so the server's SHA-256 matches the one the collector printed.
  uploadSnapshot: (assetId: string, file: Blob) =>
    request<SnapshotSummary>(`/api/v1/assets/${assetId}/snapshots`, { method: "POST", body: file }),

  runAssessment: (target: { snapshot_id: string } | { twin_id: string }) =>
    request<Assessment>("/api/v1/assessments", { method: "POST", body: JSON.stringify(target) }),

  listAssessments: () => request<{ assessments: Assessment[] }>("/api/v1/assessments"),

  getAssessment: (id: string) => request<Assessment>(`/api/v1/assessments/${id}`),

  listAssetAssessments: (assetId: string) =>
    request<{ assessments: Assessment[] }>(`/api/v1/assets/${assetId}/assessments`),

  diffAssessments: (afterId: string, beforeId: string) =>
    request<AssessmentDiff>(`/api/v1/assessments/${afterId}/diff?against=${encodeURIComponent(beforeId)}`),

  getAttackCatalog: () => request<AttackCatalog>("/api/v1/attack/catalog"),

  // A refused run comes back as 422 with the run body (its preflight
  // results), so this resolves either way instead of throwing on 422.
  startRun: async (body: {
    twin_id: string;
    scenario_id?: string;
    test_guids?: string[];
    authorisation_acknowledged: boolean;
  }): Promise<EmulationRun> => {
    try {
      return await request<EmulationRun>("/api/v1/runs", { method: "POST", body: JSON.stringify(body) });
    } catch (err) {
      if (err instanceof ApiError && err.status === 422 && err.body && typeof err.body === "object" && "preflight" in err.body) {
        return err.body as EmulationRun;
      }
      throw err;
    }
  },

  listRuns: () => request<{ runs: EmulationRun[] }>("/api/v1/runs"),

  getRun: (id: string) => request<EmulationRun>(`/api/v1/runs/${id}`),

  abortRun: (id: string) => request<EmulationRun>(`/api/v1/runs/${id}/abort`, { method: "POST" }),

  listTwins: () => request<{ twins: Twin[] }>("/api/v1/twins"),

  createTwin: (snapshotId: string) =>
    request<Twin>("/api/v1/twins", { method: "POST", body: JSON.stringify({ snapshot_id: snapshotId }) }),

  getTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}`),

  startTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}/start`, { method: "POST" }),

  stopTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}/stop`, { method: "POST" }),

  destroyTwin: (id: string) => request<Twin>(`/api/v1/twins/${id}/destroy`, { method: "POST" }),

  // Not a plain request(): this is a stream URL for the caller to hand to
  // an EventSource, not a JSON response to decode.
  twinEventsUrl: (id: string) => `${API_BASE}/api/v1/twins/${id}/events`,
};
