import { API_BASE_URL } from "../api/client";

function formatDate(value) {
  if (!value) {
    return "—";
  }

  return new Date(value).toLocaleString();
}

export default function URLList({ urls, loading, error }) {
  if (loading) {
    return <p className="muted-message">Loading your links...</p>;
  }

  if (error) {
    return <p className="error-message" role="alert">{error}</p>;
  }

  if (urls.length === 0) {
    return <p className="muted-message">You have not created any links yet.</p>;
  }

  return (
    <div className="url-list">
      {urls.map((url) => {
        const shortURL = `${API_BASE_URL}/${url.code}`;

        return (
          <article className="url-item" key={url.code}>
            <a href={shortURL} target="_blank" rel="noreferrer" className="short-link">
              /{url.code}
            </a>
            <a href={url.long_url} target="_blank" rel="noreferrer" className="long-link">
              {url.long_url}
            </a>
            <dl className="url-details">
              <div>
                <dt>Expires</dt>
                <dd>{url.expires_at ? formatDate(url.expires_at) : "Never"}</dd>
              </div>
              <div>
                <dt>Created</dt>
                <dd>{formatDate(url.created_at)}</dd>
              </div>
              {url.total_clicks !== undefined && (
                <div>
                  <dt>Clicks</dt>
                  <dd>{url.total_clicks}</dd>
                </div>
              )}
            </dl>
          </article>
        );
      })}
    </div>
  );
}
