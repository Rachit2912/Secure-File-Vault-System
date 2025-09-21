// base api url (from env or default localhost):
export const API_BASE =
  import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

// safely parse JSON (fallback to plain text if invalid):
async function parseJSON(response: Response) {
  const text = await response.text();
  try {
    return text ? JSON.parse(text) : null;
  } catch {
    return text;
  }
}

// GET request wrapper with error handling :
export async function getReq(path: string) {
  // 1. send GET request :
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  // 2. parse response safely:
  const data = await parseJSON(res);

  // 3. handle errors (rate limit, generic errors):
  if (!res.ok) {
    const msg = (data && (data.error || data.message)) || res.statusText;
    // rate limit error notificatioin :
    if (msg.toLowerCase().includes("rate limit exceeded")) {
      notifyRateLimit(msg);
    }
    const err: any = new Error(msg);
    err.status = res.status;
    throw err;
  }

  // 4. return parsed JSON :
  return data;
}

// POST request wrapper with error handling :
export async function postReq(path: string, body?: unknown) {
  // 1. send request with JSON body :
  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });

  // 2. parse response safely:
  const data = await parseJSON(res);

  // 3. handle errors (rate limit, generic errors):
  if (!res.ok) {
    const msg = (data && (data.error || data.message)) || res.statusText;
    // rate limit error notificatioin :
    if (msg.toLowerCase().includes("rate limit exceeded")) {
      notifyRateLimit(msg);
    }
    const err: any = new Error(msg);
    err.status = res.status;
    throw err;
  }

  // 4. return parsed JSON :
  return data;
}

// dispatch a global rate-limit notification event :
export function notifyRateLimit(msg: string) {
  window.dispatchEvent(new CustomEvent("rateLimitError", { detail: msg }));
}

// File upload helper (supports FormData) :
export async function handleUpload(url: string, formData: FormData) {
  // 1. send multipart/form-data request:
  const res = await fetch(url, {
    method: "POST",
    body: formData,
    credentials: "include",
  });

  // 2. handle errors with details (quota, MIME, etc.):
  if (!res.ok) {
    let errorBody: any = {};
    try {
      errorBody = await res.json();
    } catch {
      errorBody = {};
    }

    const msg = errorBody.error || "Upload failed";

    const error: any = new Error(msg);
    error.status = res.status;
    error.details = errorBody;
    throw error;
  }

  // 3. return parsed JSON :
  return res.json();
}
