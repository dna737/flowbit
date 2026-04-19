import { Box, Stack } from "@mui/material";

import { ConnectionBadge } from "./ConnectionBadge";
import { MetricTile } from "./MetricTile";
import { useMetricsHistory } from "../hooks/useMetricsHistory";
import type { ConnectionStatus } from "../hooks/useJobSocket";
import type { Job } from "../jobs/types";
import { tokens } from "../theme";

export function MetricsStrip({
  jobs,
  connectionStatus,
}: {
  jobs: Job[];
  connectionStatus: ConnectionStatus;
}) {
  const m = useMetricsHistory(jobs);

  return (
    <Box
      sx={{
        flexShrink: 0,
        height: 72,
        backgroundColor: tokens.color.bgSurface,
        borderBottom: `1px solid ${tokens.color.borderSubtle}`,
        px: `${tokens.spacing.xl}px`,
        display: "flex",
        alignItems: "stretch",
        justifyContent: "space-between",
        gap: `${tokens.spacing.xl}px`,
      }}
    >
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          gap: `${tokens.spacing.xl}px`,
          flex: 1,
          minWidth: 0,
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
      <Stack alignItems="flex-end" justifyContent="flex-start" sx={{ pt: 0.5 }}>
        <ConnectionBadge status={connectionStatus} />
      </Stack>
    </Box>
  );
}
