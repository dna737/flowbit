export type JobStatus =
  | "pending"
  | "running"
  | "retrying"
  | "succeeded"
  | "failed";

export interface Job {
  id: string;
  job_type: string;
  parameters: Record<string, unknown>;
  status: JobStatus;
  attempts: number;
  last_error?: string | null;
  created_at: string;
  updated_at: string;
}

export interface JobsState {
  jobs: Record<string, Job>;
}

/** Synthetic ids for optimistic / client-only rows before server id exists. */
export const CLIENT_JOB_ID_PREFIX = "client:";

export function isClientJobId(id: string): boolean {
  return id.startsWith(CLIENT_JOB_ID_PREFIX);
}

/** Shown on JobCard title for pending/failed client rows. */
export const FLOWBIT_PROMPT_PARAM = "_flowbit_prompt";

export type JobsAction =
  | { type: "UPSERT"; job: Job }
  | { type: "REMOVE"; id: string }
  | { type: "SNAPSHOT"; jobs: Job[] };

export interface SnapshotMessage {
  type: "snapshot";
  jobs: Job[];
}

export const STATUS_ORDER: JobStatus[] = [
  "pending",
  "running",
  "retrying",
  "succeeded",
  "failed",
];
