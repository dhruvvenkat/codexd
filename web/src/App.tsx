import { FormEvent, useEffect, useMemo, useState } from "react";
import { api, normalizeToken, Session, Status } from "./api";

const tokenKey = "codexd_token";

function routeSessionID() {
  const match = window.location.pathname.match(/^\/sessions\/([^/]+)$/);
  return match ? decodeURIComponent(match[1]) : null;
}

function navigateTo(sessionID: string | null) {
  const path = sessionID ? `/sessions/${encodeURIComponent(sessionID)}` : "/";
  window.history.pushState({}, "", path);
}

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem(tokenKey) || "");
  const [sessionID, setSessionID] = useState<string | null>(() => routeSessionID());

  useEffect(() => {
    const onPop = () => setSessionID(routeSessionID());
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  function saveToken(nextToken: string) {
    const normalized = normalizeToken(nextToken);
    if (normalized) {
      localStorage.setItem(tokenKey, normalized);
    } else {
      localStorage.removeItem(tokenKey);
    }
    setToken(normalized);
  }

  function openSession(id: string | null) {
    navigateTo(id);
    setSessionID(id);
  }

  if (!token) {
    return <TokenGate onSave={saveToken} />;
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <button className="brand" type="button" onClick={() => openSession(null)}>
          codexd
        </button>
        <button className="ghost" type="button" onClick={() => saveToken("")}>
          Change token
        </button>
      </header>
      {sessionID ? (
        <SessionDetail token={token} sessionID={sessionID} onBack={() => openSession(null)} />
      ) : (
        <SessionsPage token={token} onOpen={openSession} />
      )}
    </div>
  );
}

function TokenGate({ onSave }: { onSave: (token: string) => void }) {
  const [value, setValue] = useState("");

  function submit(event: FormEvent) {
    event.preventDefault();
    const next = value.trim();
    if (next) {
      onSave(next);
    }
  }

  return (
    <main className="token-screen">
      <form className="panel token-panel" onSubmit={submit}>
        <h1>codexd</h1>
        <label>
          Token
          <input
            autoFocus
            value={value}
            onChange={(event) => setValue(event.target.value)}
            type="password"
            autoComplete="off"
          />
        </label>
        <button type="submit">Save token</button>
      </form>
    </main>
  );
}

function SessionsPage({ token, onOpen }: { token: string; onOpen: (id: string) => void }) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [name, setName] = useState("");
  const [repoPath, setRepoPath] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function load() {
    try {
      setSessions(await api.listSessions(token));
      setError("");
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  useEffect(() => {
    void load();
    const timer = window.setInterval(load, 4000);
    return () => window.clearInterval(timer);
  }, [token]);

  async function create(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    try {
      const session = await api.createSession(token, name.trim(), repoPath.trim());
      setName("");
      setRepoPath("");
      setError("");
      onOpen(session.id);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  async function kill(id: string) {
    if (!window.confirm("Kill this session?")) {
      return;
    }
    try {
      await api.killSession(token, id);
      await load();
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  return (
    <main className="page-grid">
      <section className="panel">
        <div className="section-head">
          <h1>Sessions</h1>
          <button className="ghost" type="button" onClick={load}>
            Refresh
          </button>
        </div>
        {error && <p className="alert">{error}</p>}
        <div className="session-list">
          {sessions.length === 0 ? (
            <p className="empty">No sessions</p>
          ) : (
            sessions.map((session) => (
              <article className="session-card" key={session.id}>
                <div className="session-main">
                  <strong>{session.name}</strong>
                  <span>{session.repoPath}</span>
                </div>
                <StatusBadge status={session.status} />
                <div className="actions">
                  <button type="button" onClick={() => onOpen(session.id)}>
                    Open
                  </button>
                  <button className="danger" type="button" onClick={() => kill(session.id)}>
                    Kill
                  </button>
                </div>
              </article>
            ))
          )}
        </div>
      </section>

      <section className="panel">
        <h2>New Session</h2>
        <form className="stack" onSubmit={create}>
          <label>
            Name
            <input value={name} onChange={(event) => setName(event.target.value)} required />
          </label>
          <label>
            Repo path
            <input
              value={repoPath}
              onChange={(event) => setRepoPath(event.target.value)}
              placeholder="/home/dhruv/projects/example"
              required
            />
          </label>
          <button type="submit" disabled={busy}>
            {busy ? "Starting..." : "Start Codex"}
          </button>
        </form>
      </section>
    </main>
  );
}

function SessionDetail({
  token,
  sessionID,
  onBack
}: {
  token: string;
  sessionID: string;
  onBack: () => void;
}) {
  const [session, setSession] = useState<Session | null>(null);
  const [output, setOutput] = useState("");
  const [gitStatus, setGitStatus] = useState("");
  const [gitDiff, setGitDiff] = useState("");
  const [gitError, setGitError] = useState("");
  const [prompt, setPrompt] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const changedFiles = useMemo(
    () => gitStatus.split("\n").map((line) => line.trimEnd()).filter(Boolean),
    [gitStatus]
  );

  async function load() {
    try {
      const [nextSession, nextOutput] = await Promise.all([
        api.getSession(token, sessionID),
        api.getOutput(token, sessionID)
      ]);
      setSession(nextSession);
      setOutput(nextOutput);
      setError("");
    } catch (err) {
      setError(errorMessage(err));
      return;
    }

    try {
      const [status, diff] = await Promise.all([
        api.getGitStatus(token, sessionID),
        api.getGitDiff(token, sessionID)
      ]);
      setGitStatus(status);
      setGitDiff(diff);
      setGitError("");
    } catch (err) {
      setGitStatus("");
      setGitDiff("");
      setGitError(errorMessage(err));
    }
  }

  useEffect(() => {
    void load();
    const timer = window.setInterval(load, 3500);
    return () => window.clearInterval(timer);
  }, [token, sessionID]);

  async function send(event: FormEvent) {
    event.preventDefault();
    const text = prompt.trim();
    if (!text) {
      return;
    }
    setBusy(true);
    try {
      await api.sendInput(token, sessionID, text);
      setPrompt("");
      await load();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  async function kill() {
    if (!window.confirm("Kill this session?")) {
      return;
    }
    try {
      await api.killSession(token, sessionID);
      onBack();
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  return (
    <main className="detail-layout">
      <section className="panel detail-head">
        <button className="ghost" type="button" onClick={onBack}>
          Back
        </button>
        <div className="title-block">
          <h1>{session?.name || sessionID}</h1>
          <span>{session?.repoPath || ""}</span>
        </div>
        {session && <StatusBadge status={session.status} />}
        <button className="danger" type="button" onClick={kill}>
          Kill
        </button>
      </section>

      {error && <p className="alert">{error}</p>}

      <section className="panel">
        <div className="section-head">
          <h2>Output</h2>
          <button className="ghost" type="button" onClick={load}>
            Refresh
          </button>
        </div>
        <pre className="terminal">{output || "No output"}</pre>
        <form className="prompt-row" onSubmit={send}>
          <textarea
            value={prompt}
            onChange={(event) => setPrompt(event.target.value)}
            rows={3}
          />
          <button type="submit" disabled={busy}>
            {busy ? "Sending..." : "Send"}
          </button>
        </form>
      </section>

      <section className="split-grid">
        <div className="panel">
          <h2>Changed Files</h2>
          {gitError ? (
            <p className="alert">{gitError}</p>
          ) : changedFiles.length === 0 ? (
            <p className="empty">No changes</p>
          ) : (
            <ul className="file-list">
              {changedFiles.map((line) => (
                <li key={line}>{line}</li>
              ))}
            </ul>
          )}
        </div>
        <div className="panel">
          <h2>Diff</h2>
          <pre className="diff">{gitDiff || "No diff"}</pre>
        </div>
      </section>
    </main>
  );
}

function StatusBadge({ status }: { status: Status }) {
  return <span className={`status ${status}`}>{status.replace("_", " ")}</span>;
}

function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : "Request failed";
}
