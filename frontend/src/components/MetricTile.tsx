import { Box, Stack, Typography } from "@mui/material";

import { tokens } from "../theme";

type Accent = "blue" | "success" | "failure";

const accentColors: Record<Accent, { main: string; muted: string }> = {
  blue: { main: tokens.color.accentBlue, muted: tokens.color.accentBlueMuted },
  success: { main: tokens.color.statusSucceeded, muted: tokens.color.statusSucceededMuted },
  failure: { main: tokens.color.statusFailed, muted: tokens.color.statusFailedMuted },
};

interface MetricTileProps {
  label: string;
  value: string;
  history: number[];
  accent: Accent;
}

export function MetricTile({ label, value, history, accent }: MetricTileProps) {
  const palette = accentColors[accent];
  const max = Math.max(1, ...history);
  const splitIdx = Math.ceil(history.length / 2);

  return (
    <Stack spacing={0.5} sx={{ flex: 1, minWidth: 0 }}>
      <Typography
        sx={{
          fontFamily: tokens.font.sans,
          fontSize: 11,
          fontWeight: 500,
          color: tokens.color.textSecondary,
          letterSpacing: 0.2,
        }}
      >
        {label}
      </Typography>
      <Typography
        sx={{
          fontFamily: tokens.font.mono,
          fontSize: 22,
          fontWeight: 600,
          color: tokens.color.textPrimary,
          lineHeight: 1.1,
        }}
      >
        {value}
      </Typography>
      <Box sx={{ display: "flex", alignItems: "flex-end", gap: "2px", height: 16 }}>
        {history.map((v, i) => {
          const h = Math.max(2, Math.round((v / max) * 14));
          const color = i < splitIdx ? palette.muted : palette.main;
          return (
            <Box
              key={i}
              sx={{
                width: 4,
                height: h,
                borderRadius: "1px",
                backgroundColor: color,
              }}
            />
          );
        })}
      </Box>
    </Stack>
  );
}
