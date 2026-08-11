import { useState } from "react";

export default function ShortURLResult({ shortUrl }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(shortUrl);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1800);
    } catch {
      setCopied(false);
    }
  }

  return (
    <section className="result" aria-live="polite">
      <p className="result-label">Your shortened URL</p>
      <div className="result-row">
        <a href={shortUrl} target="_blank" rel="noreferrer">
          {shortUrl}
        </a>
        <button type="button" className="copy-button" onClick={handleCopy}>
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
    </section>
  );
}
