import { Box, Stack, Typography } from "@mui/material";
import type { ReactNode } from "react";

import { JobCard } from "./JobCard";
import type { Job, JobStatus } from "../jobs/types";
import { statusColor, tokens } from "../theme";
import { AlertIcon, CheckIcon, ClockIcon, PlayIcon, RefreshIcon } from "./icons";

interface StatusColumnProps {
  title: string;
  status: JobStatus;
  jobs: Job[];
  latestTrackedJobId: string | null;
}

const statusIcon: Record<JobStatus, ReactNode> = {
  pending: <ClockIcon />,
  running: <PlayIcon />,
  retrying: <RefreshIcon />,
  succeeded: <CheckIcon />,
  failed: <AlertIcon />,
};

export function StatusColumn({
  title,
  status,
  jobs,
  latestTrackedJobId,
}: StatusColumnProps) {
  const color = statusColor[status];
  return (
    <Box
      sx={{
        flex: "1 1 0",
        minWidth: 0,
        backgroundColor: tokens.color.bgSurface,
        border: `1px solid ${tokens.color.borderSubtle}`,
        borderRadius: `${tokens.radius.lg}px`,
        display: "flex",
        flexDirection: "column",
        minHeight: 320,
        overflow: "hidden",
      }}
    >
      <Stack
        direction="row"
        alignItems="center"
        spacing={1}
        sx={{
          px: 1.5,
          py: 1.25,
          borderBottom: `1px solid ${tokens.color.borderSubtle}`,
        }}
      >
        <Box sx={{ color: color.main, display: "flex" }}>{statusIcon[status]}</Box>
        <Typography
          sx={{
            fontFamily: tokens.font.sans,
            fontSize: 12,
            fontWeight: 500,
            color: color.main,
            textTransform: "uppercase",
            letterSpacing: 0.6,
            flex: 1,
          }}
        >
          {title}
        </Typography>
        <Box
          sx={{
            backgroundColor: color.muted,
            color: color.main,
            fontFamily: tokens.font.mono,
            fontSize: 11,
            fontWeight: 600,
            px: "6px",
            py: "1px",
            borderRadius: `${tokens.radius.sm}px`,
            minWidth: 20,
            textAlign: "center",
          }}
        >
          {jobs.length}
        </Box>
      </Stack>
      <Stack spacing={1} sx={{ p: 1.25, overflowY: "auto", flex: 1 }}>
        {jobs.length === 0 ? (
          <Typography
            sx={{
              color: tokens.color.textMuted,
              fontSize: 11,
              fontStyle: "italic",
              mt: 0.5,
            }}
          >
            empty
          </Typography>
        ) : (
          jobs.map((job) => (
            <JobCard
              key={job.id}
              job={job}
              isLatestDispatch={job.id === latestTrackedJobId}
            />
          ))
        )}
      </Stack>
    </Box>
  );
}
