import { Paper, Stack, Typography } from "@mui/material";

import { JobCard } from "./JobCard";
import type { Job, JobStatus } from "../jobs/types";

interface StatusColumnProps {
  title: string;
  status: JobStatus;
  jobs: Job[];
}

export function StatusColumn({ title, status, jobs }: StatusColumnProps) {
  return (
    <Paper
      sx={{
        p: 2,
        minHeight: 320,
        background:
          "linear-gradient(180deg, rgba(255,255,255,0.95) 0%, rgba(240,244,255,0.92) 100%)",
      }}
      variant="outlined"
    >
      <Stack spacing={2}>
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Typography variant="subtitle1">{title}</Typography>
          <Typography variant="body2" color="text.secondary">
            {jobs.length}
          </Typography>
        </Stack>
        <Stack spacing={1.5}>
          {jobs.length === 0 ? (
            <Typography variant="body2" color="text.secondary">
              No {status} jobs yet.
            </Typography>
          ) : (
            jobs.map((job) => <JobCard key={job.id} job={job} />)
          )}
        </Stack>
      </Stack>
    </Paper>
  );
}
