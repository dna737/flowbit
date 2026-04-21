import { describe, expect, it } from "vitest";

import { initialJobsState, jobsReducer } from "./reducer";
import { FLOWBIT_PROMPT_PARAM, type Job } from "./types";

const baseJob: Job = {
  id: "550e8400-e29b-41d4-a716-446655440000",
  job_type: "general",
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

  it("removes a job by id", () => {
    const next = jobsReducer(
      {
        jobs: {
          a: { ...baseJob, id: "a" },
          b: { ...baseJob, id: "b" },
        },
      },
      { type: "REMOVE", id: "a" },
    );

    expect(next.jobs.a).toBeUndefined();
    expect(next.jobs.b).toBeDefined();
  });

  it("preserves client-prefixed jobs on snapshot when not in payload", () => {
    const clientId = "client:550e8400-e29b-41d4-a716-446655440000";
    const clientJob: Job = {
      ...baseJob,
      id: clientId,
      status: "pending",
      parameters: { [FLOWBIT_PROMPT_PARAM]: "hello" },
    };

    const next = jobsReducer(
      {
        jobs: {
          [clientId]: clientJob,
        },
      },
      {
        type: "SNAPSHOT",
        jobs: [baseJob],
      },
    );

    expect(next.jobs[baseJob.id]).toEqual(baseJob);
    expect(next.jobs[clientId]).toEqual(clientJob);
  });

  it("snapshot overwrites when server sends same id as client row", () => {
    const sharedId = "client:shared";
    const clientJob: Job = {
      ...baseJob,
      id: sharedId,
      status: "pending",
    };
    const serverJob: Job = { ...baseJob, id: sharedId, status: "running", attempts: 1 };

    const next = jobsReducer(
      { jobs: { [sharedId]: clientJob } },
      { type: "SNAPSHOT", jobs: [serverJob] },
    );

    expect(next.jobs[sharedId]).toEqual(serverJob);
  });
});
