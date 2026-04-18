import { useState, type Dispatch } from "react";
import {
  Alert,
  Box,
  Button,
  IconButton,
  InputAdornment,
  Snackbar,
  TextField,
  Tooltip,
} from "@mui/material";

import { postDispatch } from "../api/client";
import {
  CLIENT_JOB_ID_PREFIX,
  FLOWBIT_PROMPT_PARAM,
  type Job,
  type JobsAction,
} from "../jobs/types";
import { tokens } from "../theme";
import { ConnectionBadge } from "./ConnectionBadge";
import { InfoIcon, SendIcon, TerminalIcon } from "./icons";
import type { ConnectionStatus } from "../hooks/useJobSocket";

interface CommandBarProps {
  dispatchJobs: Dispatch<JobsAction>;
  onWatchlistPrepend: (jobId: string, prompt: string) => void;
  onWatchlistReplaceJobId: (fromId: string, toId: string) => void;
  connectionStatus: ConnectionStatus;
}

const EXAMPLES_TOOLTIP =
  "Examples: 'resize all images to 512×512' · 'send weekly digest' · 'POST /api/nuke --force'";

function makeClientPendingJob(clientId: string, prompt: string): Job {
  const now = new Date().toISOString();
  return {
    id: clientId,
    job_type: "dispatch",
    parameters: { [FLOWBIT_PROMPT_PARAM]: prompt },
    status: "pending",
    attempts: 0,
    last_error: null,
    created_at: now,
    updated_at: now,
  };
}

export function CommandBar({
  dispatchJobs,
  onWatchlistPrepend,
  onWatchlistReplaceJobId,
  connectionStatus,
}: CommandBarProps) {
  const [prompt, setPrompt] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    const trimmed = prompt.trim();
    if (!trimmed) return;

    const clientId = `${CLIENT_JOB_ID_PREFIX}${crypto.randomUUID()}`;
    const pendingJob = makeClientPendingJob(clientId, trimmed);
    dispatchJobs({ type: "UPSERT", job: pendingJob });
    onWatchlistPrepend(clientId, trimmed);
    setPrompt("");
    setError(null);

    try {
      const job = await postDispatch(trimmed);
      dispatchJobs({ type: "REMOVE", id: clientId });
      dispatchJobs({ type: "UPSERT", job });
      onWatchlistReplaceJobId(clientId, job.id);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Dispatch failed";
      dispatchJobs({
        type: "UPSERT",
        job: {
          ...pendingJob,
          status: "failed",
          last_error: msg,
          updated_at: new Date().toISOString(),
        },
      });
      setError(msg);
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
        sx={{ flex: 1, minWidth: 0 }}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <TerminalIcon style={{ color: tokens.color.textMuted }} />
            </InputAdornment>
          ),
          endAdornment: (
            <InputAdornment position="end">
              <Tooltip
                title={EXAMPLES_TOOLTIP}
                placement="top"
                slotProps={{
                  tooltip: {
                    sx: {
                      maxWidth: 360,
                      fontFamily: tokens.font.sans,
                      fontSize: 12,
                      lineHeight: 1.45,
                    },
                  },
                }}
              >
                <IconButton
                  size="small"
                  edge="end"
                  aria-label="Example prompts"
                  sx={{ color: tokens.color.textMuted }}
                >
                  <InfoIcon width={16} height={16} />
                </IconButton>
              </Tooltip>
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
      <Button
        variant="contained"
        onClick={handleSubmit}
        disabled={!prompt.trim()}
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
        Dispatch
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
