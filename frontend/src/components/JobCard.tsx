import { Chip, Paper, Stack, Typography } from "@mui/material";

import type { Job } from "../jobs/types";

export function JobCard({ job }: { job: Job }) {
  return (
    <Paper
      variant="outlined"
      sx={{
        p: 2,
        borderRadius: 3,
        backgroundColor: "background.paper",
      }}
    >
      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" spacing={1}>
          <Typography variant="subtitle2">{job.job_type}</Typography>
          <Chip label={job.id.slice(0, 8)} size="small" variant="outlined" />
        </Stack>
        <Typography variant="body2" color="text.secondary">
          Attempts: {job.attempts}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Updated {formatRelativeTime(job.updated_at)}
        </Typography>
        {job.last_error ? (
          <Typography variant="body2" color="error.main">
            {job.last_error}
          </Typography>
        ) : null}
      </Stack>
    </Paper>
  );
}

function formatRelativeTime(value: string) {
  const updatedAt = new Date(value).getTime();
  const now = Date.now();
  const diffSeconds = Math.round((updatedAt - now) / 1000);
  const absSeconds = Math.abs(diffSeconds);
  const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });

  if (absSeconds < 60) {
    return formatter.format(diffSeconds, "second");
  }

  const diffMinutes = Math.round(diffSeconds / 60);
  if (Math.abs(diffMinutes) < 60) {
    return formatter.format(diffMinutes, "minute");
  }

  const diffHours = Math.round(diffMinutes / 60);
  if (Math.abs(diffHours) < 24) {
    return formatter.format(diffHours, "hour");
  }

  const diffDays = Math.round(diffHours / 24);
  return formatter.format(diffDays, "day");
}
