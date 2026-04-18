import { Box } from "@mui/material";
import { useCallback, useEffect, useMemo, useState } from "react";

import { CommandBar } from "./components/CommandBar";
import { MetricsStrip } from "./components/MetricsStrip";
import { PipelineBoard } from "./components/PipelineBoard";
import { RightRail } from "./components/RightRail";
import { useJobSocket } from "./hooks/useJobSocket";
import {
  loadWatchlistFromSession,
  prependWatchlistEntry,
  saveWatchlistToSession,
  type DispatchWatchlistEntry,
} from "./jobs/dispatchWatchlist";
import { JobsProvider, useJobsDispatch, useJobsState } from "./jobs/JobsContext";
import type { Job } from "./jobs/types";

export default function App() {
  return (
    <JobsProvider>
      <Dashboard />
    </JobsProvider>
  );
}

function Dashboard() {
  const { jobs } = useJobsState();
  const dispatch = useJobsDispatch();
  const connectionStatus = useJobSocket(dispatch);

  const [watchlist, setWatchlist] = useState<DispatchWatchlistEntry[]>(() =>
    loadWatchlistFromSession(),
  );

  useEffect(() => {
    saveWatchlistToSession(watchlist);
  }, [watchlist]);

  const trackedJobIds = useMemo(
    () => new Set(watchlist.map((entry) => entry.jobId)),
    [watchlist],
  );

  const latestTrackedJobId = useMemo(() => watchlist[0]?.jobId ?? null, [watchlist]);

  const handleDispatched = useCallback((job: Job, prompt: string) => {
    setWatchlist((current) => prependWatchlistEntry(current, job.id, prompt));
  }, []);

  const sortedJobs = Object.values(jobs).sort(
    (left, right) =>
      new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime(),
  );

  return (
    <Box
      sx={{
        height: "100%",
        overflow: "hidden",
        display: "flex",
        flexDirection: "column",
        backgroundColor: "background.default",
      }}
    >
      <MetricsStrip jobs={sortedJobs} />
      <Box sx={{ flex: 1, display: "flex", minHeight: 0 }}>
        <PipelineBoard
          jobs={sortedJobs}
          trackedJobIds={trackedJobIds}
          latestTrackedJobId={latestTrackedJobId}
        />
        <RightRail
          jobs={sortedJobs}
          trackedJobIds={trackedJobIds}
          latestTrackedJobId={latestTrackedJobId}
        />
      </Box>
      <CommandBar
        onJobCreated={(job) => dispatch({ type: "UPSERT", job })}
        onDispatched={handleDispatched}
        connectionStatus={connectionStatus}
      />
    </Box>
  );
}
