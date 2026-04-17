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

export type JobsAction =
  | { type: "UPSERT"; job: Job }
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
