export type Status =
  | "running"
  | "needs_input"
  | "needs_approval"
  | "error"
  | "stopped"
  | "unknown";

export type Session = {
  id: string;
  name: string;
  repoPath: string;
  tmuxName: string;
  status: Status;
  createdAt: string;
  updatedAt: string;
};

type SessionsResponse = {
  sessions: Session[];
};

type OutputResponse = {
  output: string;
};

type StatusResponse = {
  status: string;
};

type DiffResponse = {
  diff: string;
};

export function normalizeToken(token: string) {
  return token.trim().replace(/^Bearer\s+/i, "");
}

async function request<T>(token: string, path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Authorization", `Bearer ${normalizeToken(token)}`);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(path, { ...init, headers });
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) {
        message = body.error;
      }
    } catch {
      // Leave the HTTP status as the error message.
    }
    throw new Error(message);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

export const api = {
  listSessions: (token: string) =>
    request<SessionsResponse>(token, "/api/sessions").then((res) => res.sessions),
  createSession: (token: string, name: string, repoPath: string) =>
    request<Session>(token, "/api/sessions", {
      method: "POST",
      body: JSON.stringify({ name, repoPath })
    }),
  getSession: (token: string, id: string) => request<Session>(token, `/api/sessions/${id}`),
  getOutput: (token: string, id: string) =>
    request<OutputResponse>(token, `/api/sessions/${id}/output`).then((res) => res.output),
  sendInput: (token: string, id: string, text: string) =>
    request<void>(token, `/api/sessions/${id}/input`, {
      method: "POST",
      body: JSON.stringify({ text })
    }),
  getGitStatus: (token: string, id: string) =>
    request<StatusResponse>(token, `/api/sessions/${id}/git/status`).then((res) => res.status),
  getGitDiff: (token: string, id: string) =>
    request<DiffResponse>(token, `/api/sessions/${id}/git/diff`).then((res) => res.diff),
  killSession: (token: string, id: string) =>
    request<void>(token, `/api/sessions/${id}`, { method: "DELETE" })
};
