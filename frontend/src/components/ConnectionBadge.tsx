import { Box, Chip, Stack, Typography } from "@mui/material";

import type { ConnectionStatus } from "../hooks/useJobSocket";

const badgeColor: Record<ConnectionStatus, "default" | "success" | "warning" | "error"> = {
  connecting: "warning",
  connected: "success",
  reconnecting: "warning",
  disconnected: "error",
};

export function ConnectionBadge({ status }: { status: ConnectionStatus }) {
  return (
    <Stack direction="row" spacing={1} alignItems="center">
      <Typography variant="body2" color="text.secondary">
        Realtime
      </Typography>
      <Chip
        icon={
          <Box
            sx={{
              width: 10,
              height: 10,
              borderRadius: "50%",
              backgroundColor: "currentColor",
            }}
          />
        }
        label={status}
        color={badgeColor[status]}
        size="small"
        variant="outlined"
      />
    </Stack>
  );
}
