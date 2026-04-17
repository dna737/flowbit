import { Box, Stack } from "@mui/material";

import { StatusColumn } from "./StatusColumn";
import type { Job, JobStatus } from "../jobs/types";

const columnLabels: Record<JobStatus, string> = {
  pending: "Pending",
  running: "Running",
  retrying: "Retrying",
  succeeded: "Succeeded",
  failed: "Failed",
};

interface PipelineBoardProps {
  jobs: Job[];
}

export function PipelineBoard({ jobs }: PipelineBoardProps) {
  const byStatus = jobs.reduce<Record<JobStatus, Job[]>>(
    (acc, job) => {
      acc[job.status].push(job);
      return acc;
    },
    {
      pending: [],
      running: [],
      retrying: [],
      succeeded: [],
      failed: [],
    },
  );

  return (
    <Stack spacing={2}>
      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: {
            xs: "1fr",
            md: "repeat(2, minmax(0, 1fr))",
            xl: "repeat(5, minmax(0, 1fr))",
          },
        }}
      >
        {(Object.keys(columnLabels) as JobStatus[]).map((status) => (
          <Box key={status}>
            <StatusColumn title={columnLabels[status]} status={status} jobs={byStatus[status]} />
          </Box>
        ))}
      </Box>
    </Stack>
  );
}
