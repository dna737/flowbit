import { isClientJobId, type Job, type JobsAction, type JobsState } from "./types";

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
    case "REMOVE": {
      const { [action.id]: _, ...rest } = state.jobs;
      return { jobs: rest };
    }
    case "SNAPSHOT": {
      const next = action.jobs.reduce<Record<string, Job>>((acc, job) => {
        acc[job.id] = job;
        return acc;
      }, {});
      for (const [id, job] of Object.entries(state.jobs)) {
        if (isClientJobId(id) && next[id] === undefined) {
          next[id] = job;
        }
      }
      return { jobs: next };
    }
    default:
      return state;
  }
}
