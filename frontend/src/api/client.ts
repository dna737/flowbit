import type { Job } from "../jobs/types";

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "/api";

const withCredentials = {
  credentials: "include" as const,
};

/** Best-effort ping so Postgres (e.g. Neon) wakes before heavier API calls. */
export async function wakePostgres(): Promise<void> {
  try {
    const response = await fetch(`${apiBaseUrl}/readyz`, withCredentials);
    if (!response.ok) {
      console.warn("readiness ping failed", response.status);
    }
  } catch (err) {
    console.warn("readiness ping failed", err);
  }
}

interface DispatchResponse {
  error?: string;
}

export async function postDispatch(prompt: string): Promise<Job> {
  const response = await fetch(`${apiBaseUrl}/dispatch`, {
    ...withCredentials,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ prompt }),
  });

  return parseJobResponse(response);
}

interface CategoriesBody {
  categories?: string[];
  error?: string;
}

export async function getDispatchCategories(): Promise<string[]> {
  const response = await fetch(`${apiBaseUrl}/settings/dispatch-categories`, {
    ...withCredentials,
  });
  const body = (await response.json()) as CategoriesBody;
  if (!response.ok) {
    const msg =
      typeof body.error === "string"
        ? body.error
        : `request failed with status ${response.status}`;
    throw new Error(msg);
  }
  return Array.isArray(body.categories) ? body.categories : [];
}

export async function putDispatchCategories(categories: string[]): Promise<string[]> {
  console.log("put", categories);
  const response = await fetch(`${apiBaseUrl}/settings/dispatch-categories`, {
    ...withCredentials,
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ categories }),
  });
  const body = (await response.json()) as CategoriesBody;
  if (!response.ok) {
    const msg =
      typeof body.error === "string"
        ? body.error
        : `request failed with status ${response.status}`;
    throw new Error(msg);
  }
  return Array.isArray(body.categories) ? body.categories : [];
}

export async function postJob(
  type: string,
  params: Record<string, unknown>,
): Promise<Job> {
  const response = await fetch(`${apiBaseUrl}/jobs`, {
    ...withCredentials,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ job_type: type, parameters: params }),
  });

  return parseJobResponse(response);
}

async function parseJobResponse(response: Response): Promise<Job> {
  const body = (await response.json()) as Job | DispatchResponse;
  if (!response.ok) {
    const errorMessage =
      "error" in body && typeof body.error === "string"
        ? body.error
        : `request failed with status ${response.status}`;
    throw new Error(errorMessage);
  }

  return body as Job;
}
