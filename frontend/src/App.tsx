import { Box } from "@mui/material";

import { CommandBar } from "./components/CommandBar";
import { MetricsStrip } from "./components/MetricsStrip";
import { PipelineBoard } from "./components/PipelineBoard";
import { RightRail } from "./components/RightRail";
import { useJobSocket } from "./hooks/useJobSocket";
import { JobsProvider, useJobsDispatch, useJobsState } from "./jobs/JobsContext";

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
        <PipelineBoard jobs={sortedJobs} />
        <RightRail jobs={sortedJobs} />
      </Box>
      <CommandBar
        onJobCreated={(job) => dispatch({ type: "UPSERT", job })}
        connectionStatus={connectionStatus}
      />
    </Box>
  );
}
