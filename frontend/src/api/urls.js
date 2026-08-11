import { apiRequest } from "./client";

export function shortenUrl({ url, customCode, expiresAt }, token) {
  return apiRequest("/shorten", {
    method: "POST",
    token,
    body: {
      url,
      custom_code: customCode,
      expires_at: expiresAt,
    },
  });
}

export function getMyURLs(token) {
  return apiRequest("/me/urls", { token });
}

export function getAnalytics(code, token) {
  return apiRequest(`/analytics/${encodeURIComponent(code)}`, { token });
}

export function updateURL(code, changes, token) {
  return apiRequest(`/links/${encodeURIComponent(code)}`, {
    method: "PATCH",
    token,
    body: changes,
  });
}

export function deleteURL(code, token) {
  return apiRequest(`/links/${encodeURIComponent(code)}`, {
    method: "DELETE",
    token,
  });
}
