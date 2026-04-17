import type { Job } from "../jobs/types";

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "/api";

interface DispatchResponse {
  error?: string;
}

export async function postDispatch(prompt: string): Promise<Job> {
  const response = await fetch(`${apiBaseUrl}/dispatch`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ prompt }),
  });

  return parseJobResponse(response);
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
