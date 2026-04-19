import type { Job } from "../jobs/types";
import { getUserId } from "../identity/userId";

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "/api";

function withUserHeaders(headers?: HeadersInit): HeadersInit {
  return {
    ...headers,
    "X-User-Id": getUserId(),
  };
}

interface DispatchResponse {
  error?: string;
}

export async function postDispatch(prompt: string): Promise<Job> {
  const response = await fetch(`${apiBaseUrl}/dispatch`, {
    method: "POST",
    headers: withUserHeaders({
      "Content-Type": "application/json",
    }),
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
    headers: withUserHeaders(),
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
  const response = await fetch(`${apiBaseUrl}/settings/dispatch-categories`, {
    method: "PUT",
    headers: withUserHeaders({
      "Content-Type": "application/json",
    }),
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
