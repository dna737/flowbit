import type { Job, JobsAction, JobsState } from "./types";

export const initialJobsState: JobsState = {
  jobs: {},
};

export function jobsReducer(state: JobsState, action: JobsAction): JobsState {
  switch (action.type) {
    case "UPSERT":
      return {
        jobs: {
          ...state.jobs,
          [action.job.id]: action.job,
        },
      };
    case "SNAPSHOT":
      return {
        jobs: action.jobs.reduce<Record<string, Job>>((acc, job) => {
          acc[job.id] = job;
          return acc;
        }, {}),
      };
    default:
      return state;
  }
}
