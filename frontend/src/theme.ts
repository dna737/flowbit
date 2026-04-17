import { createTheme } from "@mui/material/styles";

export const tokens = {
  color: {
    bgPrimary: "#0A0A0B",
    bgSurface: "#141416",
    bgElevated: "#1C1C1F",
    bgHover: "#232328",
    borderDefault: "#3A3A40",
    borderSubtle: "#2A2A2E",
    textPrimary: "#E8E8EC",
    textSecondary: "#8B8B94",
    textMuted: "#5A5A63",
    accentBlue: "#3B82F6",
    accentBlueMuted: "#1E3A5F",
    statusSucceeded: "#22C55E",
    statusSucceededMuted: "#14532D",
    statusRunning: "#3B82F6",
    statusRetrying: "#F59E0B",
    statusRetryingMuted: "#5C3D0E",
    statusFailed: "#EF4444",
    statusFailedMuted: "#5C1A1A",
    statusPending: "#6B6B73",
  },
  font: {
    sans: '"Inter", "Segoe UI", "Helvetica Neue", Arial, sans-serif',
    mono: '"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace',
  },
  spacing: { xs: 4, sm: 8, md: 12, lg: 16, xl: 24, xxl: 32 },
  radius: { sm: 4, md: 6, lg: 8 },
} as const;

export type StatusKind = "pending" | "running" | "retrying" | "succeeded" | "failed";

export const statusColor: Record<StatusKind, { main: string; muted: string }> = {
  pending: { main: tokens.color.statusPending, muted: tokens.color.borderSubtle },
  running: { main: tokens.color.statusRunning, muted: tokens.color.accentBlueMuted },
  retrying: { main: tokens.color.statusRetrying, muted: tokens.color.statusRetryingMuted },
  succeeded: { main: tokens.color.statusSucceeded, muted: tokens.color.statusSucceededMuted },
  failed: { main: tokens.color.statusFailed, muted: tokens.color.statusFailedMuted },
};

declare module "@mui/material/styles" {
  interface Theme {
    tokens: typeof tokens;
  }
  interface ThemeOptions {
    tokens?: typeof tokens;
  }
}

export const theme = createTheme({
  tokens,
  palette: {
    mode: "dark",
    primary: { main: tokens.color.accentBlue },
    secondary: { main: tokens.color.statusSucceeded },
    error: { main: tokens.color.statusFailed },
    warning: { main: tokens.color.statusRetrying },
    success: { main: tokens.color.statusSucceeded },
    info: { main: tokens.color.accentBlue },
    background: {
      default: tokens.color.bgPrimary,
      paper: tokens.color.bgSurface,
    },
    text: {
      primary: tokens.color.textPrimary,
      secondary: tokens.color.textSecondary,
      disabled: tokens.color.textMuted,
    },
    divider: tokens.color.borderSubtle,
  },
  shape: { borderRadius: tokens.radius.md },
  typography: {
    fontFamily: tokens.font.sans,
    fontSize: 13,
    h6: { fontWeight: 600 },
    subtitle1: { fontWeight: 600, fontSize: 14 },
    subtitle2: { fontWeight: 500, fontSize: 12 },
    body2: { fontSize: 12 },
    button: { textTransform: "none", fontWeight: 500 },
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: { borderRadius: tokens.radius.md },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          backgroundColor: tokens.color.bgElevated,
          borderRadius: tokens.radius.md,
        },
      },
    },
  },
});
