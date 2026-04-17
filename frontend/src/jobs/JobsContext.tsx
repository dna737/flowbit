import {
  createContext,
  useContext,
  useMemo,
  useReducer,
  type Dispatch,
  type PropsWithChildren,
} from "react";

import { initialJobsState, jobsReducer } from "./reducer";
import type { JobsAction, JobsState } from "./types";

const JobsStateContext = createContext<JobsState | null>(null);
const JobsDispatchContext = createContext<Dispatch<JobsAction> | null>(null);

export function JobsProvider({ children }: PropsWithChildren) {
  const [state, dispatch] = useReducer(jobsReducer, initialJobsState);
  const memoizedState = useMemo(() => state, [state]);

  return (
    <JobsStateContext.Provider value={memoizedState}>
      <JobsDispatchContext.Provider value={dispatch}>
        {children}
      </JobsDispatchContext.Provider>
    </JobsStateContext.Provider>
  );
}

export function useJobsState() {
  const value = useContext(JobsStateContext);
  if (!value) {
    throw new Error("useJobsState must be used inside JobsProvider");
  }
  return value;
}

export function useJobsDispatch() {
  const value = useContext(JobsDispatchContext);
  if (!value) {
    throw new Error("useJobsDispatch must be used inside JobsProvider");
  }
  return value;
}
