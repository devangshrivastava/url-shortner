import { useState } from "react";
import { shortenUrl } from "../api/urls";

function toRFC3339(datetimeLocalValue) {
  if (!datetimeLocalValue) {
    return "";
  }

  return new Date(datetimeLocalValue).toISOString();
}

export default function ShortenForm({ onSuccess }) {
  const [url, setUrl] = useState("");
  const [customCode, setCustomCode] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event) {
    event.preventDefault();
    setError("");
    setLoading(true);

    try {
      const result = await shortenUrl({
        url,
        customCode,
        expiresAt: toRFC3339(expiresAt),
      });

      onSuccess(result);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <form className="shorten-form" onSubmit={handleSubmit}>
      <label htmlFor="long-url">Long URL</label>
      <input
        id="long-url"
        type="url"
        placeholder="https://example.com/very-long-link"
        value={url}
        onChange={(event) => setUrl(event.target.value)}
        required
      />

      <label htmlFor="custom-alias">Custom alias <span>(optional)</span></label>
      <input
        id="custom-alias"
        type="text"
        placeholder="my-link"
        value={customCode}
        onChange={(event) => setCustomCode(event.target.value)}
      />

      <label htmlFor="expiry">Expiry date and time <span>(optional)</span></label>
      <input
        id="expiry"
        type="datetime-local"
        value={expiresAt}
        onChange={(event) => setExpiresAt(event.target.value)}
      />

      {error && <p className="error-message" role="alert">{error}</p>}

      <button type="submit" disabled={loading}>
        {loading ? "Shortening..." : "Shorten URL"}
      </button>
    </form>
  );
}
