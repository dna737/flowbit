import { useCallback, useState, type Dispatch } from "react";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormHelperText,
  IconButton,
  InputAdornment,
  Snackbar,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";

import { getDispatchCategories, postDispatch, putDispatchCategories } from "../api/client";
import {
  CLIENT_JOB_ID_PREFIX,
  FLOWBIT_PROMPT_PARAM,
  type Job,
  type JobsAction,
} from "../jobs/types";
import { tokens } from "../theme";
import { InfoIcon, SendIcon, TerminalIcon, XIcon } from "./icons";

interface CommandBarProps {
  dispatchJobs: Dispatch<JobsAction>;
  onWatchlistPrepend: (jobId: string, prompt: string) => void;
  onWatchlistReplaceJobId: (fromId: string, toId: string) => void;
}

const EXAMPLES_TOOLTIP =
  "Examples: 'resize all images to 512×512' · 'send weekly digest' · 'POST /api/nuke --force'";

const pillText = "#0A0A0B";
const pillBg = "#E8E8E8";

const MAX_DISPATCH_CATEGORIES = 10;

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

function dedupeCategories(next: string, list: string[]): string[] {
  const t = next.trim();
  if (!t) return list;
  if (list.length >= MAX_DISPATCH_CATEGORIES) return list;
  const lower = t.toLowerCase();
  if (list.some((c) => c.toLowerCase() === lower)) return list;
  return [...list, t];
}

export function CommandBar({
  dispatchJobs,
  onWatchlistPrepend,
  onWatchlistReplaceJobId,
}: CommandBarProps) {
  const [prompt, setPrompt] = useState("");
  const [error, setError] = useState<string | null>(null);

  const [settingsOpen, setSettingsOpen] = useState(false);
  const [categories, setCategories] = useState<string[]>([]);
  const [categoryInput, setCategoryInput] = useState("");
  const [settingsLoadError, setSettingsLoadError] = useState<string | null>(null);
  const [settingsSaveError, setSettingsSaveError] = useState<string | null>(null);
  const [settingsLoading, setSettingsLoading] = useState(false);

  const loadCategories = useCallback(async () => {
    setSettingsLoadError(null);
    setSettingsLoading(true);
    try {
      const list = await getDispatchCategories();
      setCategories(list);
    } catch (e) {
      setSettingsLoadError(e instanceof Error ? e.message : "Failed to load categories");
    } finally {
      setSettingsLoading(false);
    }
  }, []);

  const openSettings = () => {
    setSettingsOpen(true);
    setCategoryInput("");
    setSettingsSaveError(null);
    void loadCategories();
  };

  const handleSaveSettings = async () => {
    setSettingsSaveError(null);
    setSettingsLoading(true);
    try {
      const saved = await putDispatchCategories(categories);
      setCategories(saved);
      setSettingsOpen(false);
    } catch (e) {
      setSettingsSaveError(e instanceof Error ? e.message : "Save failed");
    } finally {
      setSettingsLoading(false);
    }
  };

  const addCategoryFromInput = () => {
    setCategories((prev) => dedupeCategories(categoryInput, prev));
    setCategoryInput("");
  };

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
      <Button
        variant="outlined"
        onClick={openSettings}
        sx={{
          height: 40,
          px: 2,
          fontWeight: 500,
          borderColor: tokens.color.borderDefault,
          color: tokens.color.textSecondary,
          "&:hover": {
            borderColor: tokens.color.borderDefault,
            backgroundColor: tokens.color.bgHover,
          },
        }}
      >
        Configure
      </Button>
      <Dialog
        open={settingsOpen}
        onClose={() => !settingsLoading && setSettingsOpen(false)}
        maxWidth="sm"
        fullWidth
        slotProps={{
          backdrop: {
            sx: {
              backdropFilter: "blur(8px)",
              backgroundColor: "rgba(0,0,0,0.65)",
            },
          },
        }}
        PaperProps={{
          sx: {
            backgroundColor: tokens.color.bgSurface,
            border: `1px solid ${tokens.color.borderSubtle}`,
          },
        }}
      >
        <DialogTitle sx={{ fontFamily: tokens.font.sans, fontSize: 16, fontWeight: 600 }}>
          Settings
        </DialogTitle>
        <DialogContent>
          <Typography
            component="label"
            htmlFor="category-input"
            sx={{ display: "block", mb: 0.5, fontSize: 13, fontWeight: 500, color: tokens.color.textPrimary }}
          >
            Categories
          </Typography>
          <FormHelperText sx={{ m: 0, mb: 1.5, fontSize: 12, color: tokens.color.textMuted }}>
            These labels are what dispatched jobs will be categorized into for routing and display. Maximum{" "}
            {MAX_DISPATCH_CATEGORIES} categories.
          </FormHelperText>
          {settingsLoadError ? (
            <Alert severity="error" sx={{ mb: 2 }}>
              {settingsLoadError}
            </Alert>
          ) : null}
          {settingsSaveError ? (
            <Alert severity="error" sx={{ mb: 2 }}>
              {settingsSaveError}
            </Alert>
          ) : null}
          <TextField
            id="category-input"
            size="small"
            fullWidth
            placeholder={
              categories.length >= MAX_DISPATCH_CATEGORIES
                ? "Maximum categories reached"
                : "e.g. react, postgres"
            }
            value={categoryInput}
            disabled={settingsLoading || categories.length >= MAX_DISPATCH_CATEGORIES}
            onChange={(e) => setCategoryInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                if (categories.length < MAX_DISPATCH_CATEGORIES) {
                  addCategoryFromInput();
                }
              }
            }}
            sx={{
              mb: 2,
              "& .MuiInputBase-input": { fontFamily: tokens.font.sans, fontSize: 13 },
            }}
          />
          <Stack direction="row" flexWrap="wrap" gap={1} useFlexGap>
            {categories.map((cat) => (
              <Box
                key={cat}
                sx={{
                  position: "relative",
                  display: "inline-flex",
                  alignItems: "center",
                  maxWidth: "100%",
                  pr: 0.5,
                  borderRadius: 999,
                  backgroundColor: pillBg,
                  color: pillText,
                  pl: 1.5,
                  py: 0.5,
                  fontFamily: tokens.font.sans,
                  fontSize: 12,
                  fontWeight: 500,
                }}
              >
                <Typography
                  component="span"
                  sx={{
                    fontSize: "inherit",
                    fontFamily: "inherit",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    maxWidth: 240,
                  }}
                >
                  {cat}
                </Typography>
                <IconButton
                  size="small"
                  aria-label={`Remove ${cat}`}
                  onClick={() => setCategories((prev) => prev.filter((c) => c !== cat))}
                  sx={{
                    p: 0.25,
                    ml: 0.25,
                    color: pillText,
                    "&:hover": { backgroundColor: "rgba(0,0,0,0.08)" },
                  }}
                >
                  <XIcon width={14} height={14} />
                </IconButton>
              </Box>
            ))}
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setSettingsOpen(false)} disabled={settingsLoading} color="inherit">
            Cancel
          </Button>
          <Button variant="contained" onClick={() => void handleSaveSettings()} disabled={settingsLoading}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
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
