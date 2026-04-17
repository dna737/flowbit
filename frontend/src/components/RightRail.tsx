import { Box, Stack, Typography } from "@mui/material";

import { DeadLetterQueue } from "./DeadLetterQueue";
import { JobCard } from "./JobCard";
import type { Job } from "../jobs/types";
import { tokens } from "../theme";

export function RightRail({ jobs }: { jobs: Job[] }) {
  const feed = jobs.slice(0, 50);

  return (
    <Box
      sx={{
        width: 420,
        flexShrink: 0,
        borderLeft: `1px solid ${tokens.color.borderSubtle}`,
        display: "flex",
        flexDirection: "column",
        backgroundColor: tokens.color.bgPrimary,
        minHeight: 0,
      }}
    >
      <Box
        sx={{
          px: `${tokens.spacing.lg}px`,
          py: `${tokens.spacing.md}px`,
          borderBottom: `1px solid ${tokens.color.borderSubtle}`,
        }}
      >
        <Typography
          sx={{
            fontSize: 12,
            fontWeight: 600,
            color: tokens.color.textPrimary,
            textTransform: "uppercase",
            letterSpacing: 0.6,
          }}
        >
          Live Feed
        </Typography>
        <Typography sx={{ fontSize: 11, color: tokens.color.textMuted, mt: 0.25 }}>
          {feed.length} job{feed.length === 1 ? "" : "s"}
        </Typography>
      </Box>
      <Box
        sx={{
          flex: 1,
          overflowY: "auto",
          px: `${tokens.spacing.md}px`,
          py: `${tokens.spacing.md}px`,
        }}
      >
        {feed.length === 0 ? (
          <Typography sx={{ color: tokens.color.textMuted, fontSize: 12 }}>
            No jobs yet. Dispatch one below.
          </Typography>
        ) : (
          <Stack spacing={1}>
            {feed.map((j) => (
              <JobCard key={j.id} job={j} />
            ))}
          </Stack>
        )}
      </Box>
      <DeadLetterQueue jobs={jobs} />
    </Box>
  );
}
