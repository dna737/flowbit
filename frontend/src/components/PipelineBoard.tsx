import { Box } from "@mui/material";
import { LayoutGroup } from "motion/react";

import { StatusColumn } from "./StatusColumn";
import type { Job, JobStatus } from "../jobs/types";
import { tokens } from "../theme";

function sortColumnJobs(jobs: Job[], tracked: Set<string>): Job[] {
  return [...jobs].sort((a, b) => {
    const aTracked = tracked.has(a.id) ? 0 : 1;
    const bTracked = tracked.has(b.id) ? 0 : 1;
    if (aTracked !== bTracked) return aTracked - bTracked;
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
  });
}

const columnLabels: Record<JobStatus, string> = {
  pending: "Pending",
  running: "Running",
  retrying: "Retrying",
  succeeded: "Succeeded",
  failed: "Failed",
};

interface PipelineBoardProps {
  jobs: Job[];
  trackedJobIds: Set<string>;
  latestTrackedJobId: string | null;
}

export function PipelineBoard({
  jobs,
  trackedJobIds,
  latestTrackedJobId,
}: PipelineBoardProps) {
  const byStatus = jobs.reduce<Record<JobStatus, Job[]>>(
    (acc, job) => {
      acc[job.status].push(job);
      return acc;
    },
    { pending: [], running: [], retrying: [], succeeded: [], failed: [] },
  );

  (Object.keys(byStatus) as JobStatus[]).forEach((status) => {
    byStatus[status] = sortColumnJobs(byStatus[status], trackedJobIds);
  });

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
      <LayoutGroup id="pipeline-board">
        {(Object.keys(columnLabels) as JobStatus[]).map((status) => (
          <StatusColumn
            key={status}
            title={columnLabels[status]}
            status={status}
            jobs={byStatus[status]}
            latestTrackedJobId={latestTrackedJobId}
          />
        ))}
      </LayoutGroup>
    </Box>
  );
}
