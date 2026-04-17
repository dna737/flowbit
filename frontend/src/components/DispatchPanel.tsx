import { useState } from "react";
import { Alert, Button, Paper, Stack, TextField, Typography } from "@mui/material";

import { postDispatch } from "../api/client";
import type { Job } from "../jobs/types";

interface DispatchPanelProps {
  onJobCreated: (job: Job) => void;
}

export function DispatchPanel({ onJobCreated }: DispatchPanelProps) {
  const [prompt, setPrompt] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!prompt.trim()) {
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const job = await postDispatch(prompt.trim());
      onJobCreated(job);
      setPrompt("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "dispatch failed");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Dispatch a job</Typography>
        <TextField
          label="Plain-English prompt"
          multiline
          minRows={3}
          value={prompt}
          onChange={(event) => setPrompt(event.target.value)}
          placeholder="Send an email to bob@example.com about tomorrow's launch"
        />
        {error ? <Alert severity="error">{error}</Alert> : null}
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={submitting || !prompt.trim()}
          sx={{ alignSelf: "flex-start" }}
        >
          {submitting ? "Dispatching..." : "Submit"}
        </Button>
      </Stack>
    </Paper>
  );
}
