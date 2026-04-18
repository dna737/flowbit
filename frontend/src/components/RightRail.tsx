import { Box, Stack, Typography } from "@mui/material";

import { DeadLetterQueue } from "./DeadLetterQueue";
import { JobCard } from "./JobCard";
import type { Job } from "../jobs/types";
import { tokens } from "../theme";

function sortFeedJobs(jobs: Job[], tracked: Set<string>): Job[] {
  const trackedJobs = jobs.filter((j) => tracked.has(j.id));
  const other = jobs.filter((j) => !tracked.has(j.id));
  const byUpdated = (a: Job, b: Job) =>
    new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
  trackedJobs.sort(byUpdated);
  other.sort(byUpdated);
  return [...trackedJobs, ...other].slice(0, 50);
}

export function RightRail({
  jobs,
  trackedJobIds,
  latestTrackedJobId,
}: {
  jobs: Job[];
  trackedJobIds: Set<string>;
  latestTrackedJobId: string | null;
}) {
  const feed = sortFeedJobs(jobs, trackedJobIds);

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
              <JobCard
                key={j.id}
                job={j}
                isLatestDispatch={j.id === latestTrackedJobId}
              />
            ))}
          </Stack>
        )}
      </Box>
      <DeadLetterQueue jobs={jobs} />
    </Box>
  );
}
