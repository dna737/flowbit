import { Box, Stack, Typography } from "@mui/material";
import { motion, useReducedMotion } from "motion/react";

import { FLOWBIT_PROMPT_PARAM, type Job } from "../jobs/types";
import { statusColor, tokens } from "../theme";
import { JobCardStatusGlyph } from "./JobCardStatusGlyph";

interface JobCardProps {
  job: Job;
  /** Most recent job dispatched from this UI in the session (watchlist head). */
  isLatestDispatch?: boolean;
}

export function JobCard({ job, isLatestDispatch }: JobCardProps) {
  const color = statusColor[job.status];
  const prefersReducedMotion = useReducedMotion();

  return (
    <motion.div
      layout={!prefersReducedMotion}
      layoutId={prefersReducedMotion ? undefined : `job-card-${job.id}`}
      transition={
        prefersReducedMotion
          ? { duration: 0 }
          : { type: "spring", stiffness: 400, damping: 32 }
      }
      style={{ width: "100%", minWidth: 0 }}
    >
      <Box
        sx={{
          p: 1.25,
          borderRadius: `${tokens.radius.md}px`,
          backgroundColor: isLatestDispatch ? tokens.color.accentBlueMuted : tokens.color.bgElevated,
          border: isLatestDispatch
            ? `1px solid ${tokens.color.accentBlue}`
            : `1px solid ${tokens.color.borderSubtle}`,
          borderLeft: `3px solid ${color.main}`,
          boxShadow: isLatestDispatch ? `0 0 0 1px ${tokens.color.accentBlue}33` : undefined,
        }}
      >
      <Stack spacing={0.75}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={1}>
          {job.status === "retrying" ||
          job.status === "succeeded" ||
          job.status === "failed" ? (
            <Box
              key={`${job.id}-${job.status}`}
              sx={{
                flexShrink: 0,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                minWidth: 18,
              }}
            >
              <JobCardStatusGlyph status={job.status} color={color.main} />
            </Box>
          ) : null}
          <Stack
            direction="row"
            alignItems="center"
            spacing={0.75}
            sx={{ flex: 1, minWidth: 0 }}
          >
            {isLatestDispatch ? (
              <Typography
                sx={{
                  flexShrink: 0,
                  fontSize: 9,
                  fontWeight: 700,
                  letterSpacing: 0.4,
                  color: tokens.color.accentBlue,
                  textTransform: "uppercase",
                }}
              >
                LATEST
              </Typography>
            ) : null}
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
              {titleForJob(job)}
            </Typography>
          </Stack>
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
    </motion.div>
  );
}

function titleForJob(job: Job): string {
  const p = job.parameters[FLOWBIT_PROMPT_PARAM];
  return typeof p === "string" && p.trim() !== "" ? p : job.job_type;
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
