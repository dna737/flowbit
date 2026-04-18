import { Box } from "@mui/material";

import { StatusColumn } from "./StatusColumn";
import type { Job, JobStatus } from "../jobs/types";
import { tokens } from "../theme";

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
    { pending: [], running: [], retrying: [], succeeded: [], failed: [] },
  );

  return (
    <Box
      sx={{
        flex: 1,
        minWidth: 0,
        minHeight: 0,
        p: `${tokens.spacing.lg}px`,
        display: "flex",
        gap: `${tokens.spacing.md}px`,
        overflow: "hidden",
        overflowX: "auto",
        alignItems: "stretch",
      }}
    >
      {(Object.keys(columnLabels) as JobStatus[]).map((status) => (
        <StatusColumn
          key={status}
          title={columnLabels[status]}
          status={status}
          jobs={byStatus[status]}
        />
      ))}
    </Box>
  );
}
