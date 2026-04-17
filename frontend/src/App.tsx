import { Box, Container, Stack, Typography } from "@mui/material";

import { ConnectionBadge } from "./components/ConnectionBadge";
import { DispatchPanel } from "./components/DispatchPanel";
import { PipelineBoard } from "./components/PipelineBoard";
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
        minHeight: "100vh",
        py: 6,
        background:
          "radial-gradient(circle at top left, rgba(18,86,216,0.16), transparent 32%), linear-gradient(180deg, #f5f7fb 0%, #eef3fb 100%)",
      }}
    >
      <Container maxWidth="xl">
        <Stack spacing={3}>
          <Stack
            direction={{ xs: "column", md: "row" }}
            justifyContent="space-between"
            alignItems={{ xs: "flex-start", md: "center" }}
            spacing={2}
          >
            <Box>
              <Typography variant="h4">Flowbit Live Visualizer</Typography>
              <Typography variant="body1" color="text.secondary">
                Watch jobs move across the pipeline in real time.
              </Typography>
            </Box>
            <ConnectionBadge status={connectionStatus} />
          </Stack>
          <DispatchPanel onJobCreated={(job) => dispatch({ type: "UPSERT", job })} />
          <PipelineBoard jobs={sortedJobs} />
        </Stack>
      </Container>
    </Box>
  );
}
