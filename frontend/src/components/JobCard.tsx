import { Box, Stack, Typography } from "@mui/material";

import type { Job } from "../jobs/types";
import { statusColor, tokens } from "../theme";

export function JobCard({ job }: { job: Job }) {
  const color = statusColor[job.status];

  return (
    <Box
      sx={{
        p: 1.25,
        borderRadius: `${tokens.radius.md}px`,
        backgroundColor: tokens.color.bgElevated,
        border: `1px solid ${tokens.color.borderSubtle}`,
        borderLeft: `2px solid ${color.main}`,
      }}
    >
      <Stack spacing={0.75}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={1}>
          <Typography
            sx={{
              minWidth: 0,
              fontFamily: tokens.font.sans,
              fontSize: 12,
              fontWeight: 600,
              color: tokens.color.textPrimary,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {job.job_type}
          </Typography>
          <Box
            sx={{
              fontFamily: tokens.font.mono,
              fontSize: 10,
              color: tokens.color.textSecondary,
              backgroundColor: tokens.color.bgSurface,
              border: `1px solid ${tokens.color.borderSubtle}`,
              borderRadius: `${tokens.radius.sm}px`,
              px: 0.75,
              py: "1px",
            }}
          >
            {job.id.slice(0, 8)}
          </Box>
        </Stack>
        <Stack direction="row" spacing={1.5} alignItems="center">
          <Typography
            sx={{
              fontFamily: tokens.font.mono,
              fontSize: 10,
              color: tokens.color.textMuted,
            }}
          >
            attempts {job.attempts}
          </Typography>
          <Typography
            sx={{
              fontFamily: tokens.font.mono,
              fontSize: 10,
              color: tokens.color.textMuted,
            }}
          >
            {formatRelativeTime(job.updated_at)}
          </Typography>
        </Stack>
        {job.last_error ? (
          <Typography
            sx={{
              fontFamily: tokens.font.mono,
              fontSize: 10,
              color: tokens.color.statusFailed,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {job.last_error}
          </Typography>
        ) : null}
      </Stack>
    </Box>
  );
}

function formatRelativeTime(value: string) {
  const updatedAt = new Date(value).getTime();
  const now = Date.now();
  const diffSeconds = Math.round((updatedAt - now) / 1000);
  const absSeconds = Math.abs(diffSeconds);
  const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });

  if (absSeconds < 60) return formatter.format(diffSeconds, "second");
  const diffMinutes = Math.round(diffSeconds / 60);
  if (Math.abs(diffMinutes) < 60) return formatter.format(diffMinutes, "minute");
  const diffHours = Math.round(diffMinutes / 60);
  if (Math.abs(diffHours) < 24) return formatter.format(diffHours, "hour");
  const diffDays = Math.round(diffHours / 24);
  return formatter.format(diffDays, "day");
}
