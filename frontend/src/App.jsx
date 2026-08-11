import { useEffect, useState } from "react";
import { getMyURLs } from "./api/urls";
import LoginForm from "./components/LoginForm";
import ShortenForm from "./components/ShortenForm";
import ShortURLResult from "./components/ShortURLResult";
import SignupForm from "./components/SignupForm";
import URLList from "./components/URLList";

const TOKEN_KEY = "token";

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY));
  const [authView, setAuthView] = useState("login");
  const [authMessage, setAuthMessage] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [urls, setUrls] = useState([]);
  const [loadingURLs, setLoadingURLs] = useState(false);
  const [urlsError, setURLsError] = useState("");

  useEffect(() => {
    if (!token) {
      setUrls([]);
      return;
    }

    loadURLs(token);
  }, [token]);

  async function loadURLs(activeToken) {
    setLoadingURLs(true);
    setURLsError("");

    try {
      const result = await getMyURLs(activeToken);
      setUrls(result);
    } catch (requestError) {
      if (requestError.status === 401) {
        handleUnauthorized();
        return;
      }

      setURLsError(requestError.message || "Unable to load your links.");
    } finally {
      setLoadingURLs(false);
    }
  }

  function handleLogin(newToken) {
    localStorage.setItem(TOKEN_KEY, newToken);
    setAuthMessage("");
    setToken(newToken);
  }

  function handleLogout() {
    localStorage.removeItem(TOKEN_KEY);
    setShortUrl("");
    setAuthView("login");
    setToken(null);
  }

  function handleUnauthorized() {
    handleLogout();
    setAuthMessage("Your session has expired. Please log in again.");
  }

  function handleSignupSuccess() {
    setAuthView("login");
    setAuthMessage("Account created. You can log in now.");
  }

  async function handleShortenSuccess(result) {
    setShortUrl(result.short_url);
    await loadURLs(token);
  }

  if (!token) {
    return (
      <main className="page-shell">
        <section className="card auth-card">
          <div className="intro">
            <h1>URL Shortener</h1>
            <p className="subtitle">Sign in to create and manage your links.</p>
          </div>

          <div className="auth-tabs" aria-label="Authentication options">
            <button
              type="button"
              className={authView === "login" ? "tab active" : "tab"}
              onClick={() => setAuthView("login")}
            >
              Login
            </button>
            <button
              type="button"
              className={authView === "signup" ? "tab active" : "tab"}
              onClick={() => setAuthView("signup")}
            >
              Signup
            </button>
          </div>

          {authMessage && <p className="success-message">{authMessage}</p>}

          {authView === "login" ? (
            <LoginForm onLogin={handleLogin} />
          ) : (
            <SignupForm onSignupSuccess={handleSignupSuccess} />
          )}
        </section>
      </main>
    );
  }

  return (
    <main className="page-shell">
      <section className="card dashboard-card">
        <header className="dashboard-header">
          <div>
            <h1>URL Shortener</h1>
            <p className="subtitle">Create and manage your links.</p>
          </div>
          <button type="button" className="logout-button" onClick={handleLogout}>
            Logout
          </button>
        </header>

        <section>
          <h2>Create Link</h2>
          <ShortenForm
            token={token}
            onSuccess={handleShortenSuccess}
            onUnauthorized={handleUnauthorized}
          />
          {shortUrl && <ShortURLResult shortUrl={shortUrl} />}
        </section>

        <section className="links-section">
          <h2>My Links</h2>
          <URLList urls={urls} loading={loadingURLs} error={urlsError} />
        </section>
      </section>
    </main>
  );
}
