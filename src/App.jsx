import { useState } from "react";
import ShortenForm from "./components/ShortenForm";
import ShortURLResult from "./components/ShortURLResult";

export default function App() {
  const [shortUrl, setShortUrl] = useState("");

  return (
    <main className="page-shell">
      <section className="card">
        <div className="intro">
          <p className="eyebrow">Quick links</p>
          <h1>URL Shortener</h1>
          <p className="subtitle">Make long links short</p>
        </div>

        <ShortenForm onSuccess={(result) => setShortUrl(result.short_url)} />

        {shortUrl && <ShortURLResult shortUrl={shortUrl} />}
      </section>
    </main>
  );
}
