const API_BASE_URL = "http://localhost:8080";

export async function shortenUrl({ url, customCode, expiresAt }) {
  const response = await fetch(`${API_BASE_URL}/shorten`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      url,
      custom_code: customCode,
      expires_at: expiresAt,
    }),
  });

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    throw new Error(data.error || "Unable to shorten this URL.");
  }

  return data;
}
