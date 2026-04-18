import { useState } from "react";
import {
  Alert,
  Box,
  Button,
  InputAdornment,
  Snackbar,
  Stack,
  TextField,
  Typography,
} from "@mui/material";

import { postDispatch } from "../api/client";
import type { Job } from "../jobs/types";
import { tokens } from "../theme";
import { ConnectionBadge } from "./ConnectionBadge";
import { SendIcon, TerminalIcon } from "./icons";
import type { ConnectionStatus } from "../hooks/useJobSocket";

interface CommandBarProps {
  onJobCreated: (job: Job) => void;
  connectionStatus: ConnectionStatus;
}

const HELPER_EXAMPLES =
  "Try: 'resize all images to 512×512' · 'send weekly digest' · 'POST /api/nuke --force'";

export function CommandBar({ onJobCreated, connectionStatus }: CommandBarProps) {
  const [prompt, setPrompt] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    const trimmed = prompt.trim();
    if (!trimmed || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      const job = await postDispatch(trimmed);
      onJobCreated(job);
      setPrompt("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Dispatch failed");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Box
      sx={{
        flexShrink: 0,
        height: 80,
        backgroundColor: tokens.color.bgSurface,
        borderTop: `1px solid ${tokens.color.borderSubtle}`,
        px: `${tokens.spacing.xl}px`,
        display: "flex",
        alignItems: "center",
        gap: `${tokens.spacing.lg}px`,
      }}
    >
      <Stack spacing={0.5} sx={{ flex: 1, minWidth: 0 }}>
        <TextField
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              handleSubmit();
            }
          }}
          placeholder="send an email to bob@example.com about tomorrow's launch"
          fullWidth
          size="small"
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <TerminalIcon style={{ color: tokens.color.textMuted }} />
              </InputAdornment>
            ),
            sx: {
              height: 40,
              fontFamily: tokens.font.sans,
              fontSize: 13,
              backgroundColor: tokens.color.bgElevated,
              "& fieldset": { borderColor: tokens.color.borderSubtle },
              "&:hover fieldset": { borderColor: tokens.color.borderDefault },
            },
          }}
        />
        <Typography
          sx={{
            fontFamily: tokens.font.mono,
            fontSize: 11,
            color: tokens.color.textMuted,
          }}
        >
          {HELPER_EXAMPLES}
        </Typography>
      </Stack>
      <Button
        variant="contained"
        onClick={handleSubmit}
        disabled={submitting || !prompt.trim()}
        startIcon={<SendIcon />}
        sx={{
          height: 40,
          px: 2,
          backgroundColor: tokens.color.accentBlue,
          color: "#fff",
          fontWeight: 500,
          "&:hover": { backgroundColor: "#2563EB" },
        }}
      >
        {submitting ? "Dispatching…" : "Dispatch"}
      </Button>
      <ConnectionBadge status={connectionStatus} />
      <Snackbar
        open={!!error}
        autoHideDuration={4000}
        onClose={() => setError(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
      >
        <Alert severity="error" variant="filled" onClose={() => setError(null)}>
          {error}
        </Alert>
      </Snackbar>
    </Box>
  );
}
