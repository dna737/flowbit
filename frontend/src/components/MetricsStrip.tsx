import { Box } from "@mui/material";

import { MetricTile } from "./MetricTile";
import { useMetricsHistory } from "../hooks/useMetricsHistory";
import type { Job } from "../jobs/types";
import { tokens } from "../theme";

export function MetricsStrip({ jobs }: { jobs: Job[] }) {
  const m = useMetricsHistory(jobs);

  return (
    <Box
      sx={{
        height: 72,
        backgroundColor: tokens.color.bgSurface,
        borderBottom: `1px solid ${tokens.color.borderSubtle}`,
        px: `${tokens.spacing.xl}px`,
        display: "flex",
        alignItems: "center",
        gap: `${tokens.spacing.xl}px`,
      }}
    >
      <MetricTile
        label="Queue Depth"
        value={m.queueDepth.toLocaleString()}
        history={m.queueHistory}
        accent="blue"
      />
      <MetricTile
        label="Throughput (30s)"
        value={m.throughput.toLocaleString()}
        history={m.throughputHistory}
        accent="blue"
      />
      <MetricTile
        label="Success Rate"
        value={`${m.successRate}%`}
        history={m.successHistory}
        accent="success"
      />
      <MetricTile
        label="Error Rate"
        value={`${m.errorRate}%`}
        history={m.errorHistory}
        accent="failure"
      />
    </Box>
  );
}
