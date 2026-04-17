import { Box, Stack, Typography } from "@mui/material";

import type { ConnectionStatus } from "../hooks/useJobSocket";
import { tokens } from "../theme";

const palette: Record<ConnectionStatus, { dot: string; label: string; text: string }> = {
  connecting: {
    dot: tokens.color.statusRetrying,
    label: "Connecting",
    text: tokens.color.statusRetrying,
  },
  connected: {
    dot: tokens.color.statusSucceeded,
    label: "Connected",
    text: tokens.color.statusSucceeded,
  },
  reconnecting: {
    dot: tokens.color.statusRetrying,
    label: "Reconnecting",
    text: tokens.color.statusRetrying,
  },
  disconnected: {
    dot: tokens.color.statusFailed,
    label: "Disconnected",
    text: tokens.color.statusFailed,
  },
};

export function ConnectionBadge({ status }: { status: ConnectionStatus }) {
  const p = palette[status];
  return (
    <Stack
      direction="row"
      spacing={1}
      alignItems="center"
      sx={{
        height: 28,
        px: 1.25,
        borderRadius: `${tokens.radius.md}px`,
        border: `1px solid ${tokens.color.borderSubtle}`,
        backgroundColor: tokens.color.bgElevated,
      }}
    >
      <Box
        sx={{
          width: 8,
          height: 8,
          borderRadius: "50%",
          backgroundColor: p.dot,
          boxShadow: `0 0 8px ${p.dot}`,
        }}
      />
      <Typography
        sx={{
          fontFamily: tokens.font.sans,
          fontSize: 11,
          fontWeight: 500,
          color: p.text,
        }}
      >
        {p.label}
      </Typography>
    </Stack>
  );
}
