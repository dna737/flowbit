import { describe, expect, it } from "vitest";

import { initialJobsState, jobsReducer } from "./reducer";
import type { Job } from "./types";

const baseJob: Job = {
  id: "550e8400-e29b-41d4-a716-446655440000",
  job_type: "echo",
  parameters: { message: "hello" },
  status: "pending",
  attempts: 0,
  last_error: null,
  created_at: "2026-04-16T10:00:00Z",
  updated_at: "2026-04-16T10:00:00Z",
};

describe("jobsReducer", () => {
  it("replaces state with snapshot payload", () => {
    const next = jobsReducer(
      {
        jobs: {
          old: { ...baseJob, id: "old" },
        },
      },
      {
        type: "SNAPSHOT",
        jobs: [
          baseJob,
          { ...baseJob, id: "newer", status: "running", attempts: 1 },
        ],
      },
    );

    expect(Object.keys(next.jobs)).toEqual([
      "550e8400-e29b-41d4-a716-446655440000",
      "newer",
    ]);
  });

  it("upserts a single job", () => {
    const next = jobsReducer(initialJobsState, {
      type: "UPSERT",
      job: baseJob,
    });

    expect(next.jobs[baseJob.id]).toEqual(baseJob);

    const updated = jobsReducer(next, {
      type: "UPSERT",
      job: { ...baseJob, status: "succeeded", attempts: 1 },
    });

    expect(updated.jobs[baseJob.id].status).toBe("succeeded");
    expect(updated.jobs[baseJob.id].attempts).toBe(1);
  });
});
