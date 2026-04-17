import { useState } from "react";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from "@mui/material";
import type { Job } from "../jobs/types";
import { tokens } from "../theme";
import { ChevronDownIcon } from "./icons";

export function DeadLetterQueue({ jobs }: { jobs: Job[] }) {
  const dlq = jobs.filter((j) => j.status === "failed");
  const [payload, setPayload] = useState<Job | null>(null);

  return (
    <>
      <Accordion
        defaultExpanded
        disableGutters
        elevation={0}
        sx={{
          backgroundColor: "transparent",
          borderTop: `1px solid ${tokens.color.borderSubtle}`,
          "&:before": { display: "none" },
        }}
      >
        <AccordionSummary
          expandIcon={<ChevronDownIcon style={{ color: tokens.color.textSecondary }} />}
          sx={{ px: `${tokens.spacing.lg}px`, minHeight: 48 }}
        >
          <Stack direction="row" alignItems="center" spacing={1} sx={{ flex: 1 }}>
            <Typography
              sx={{
                fontFamily: tokens.font.sans,
                fontSize: 12,
                fontWeight: 600,
                color: tokens.color.textPrimary,
                textTransform: "uppercase",
                letterSpacing: 0.6,
              }}
            >
              Dead Letter Queue
            </Typography>
            <Box
              sx={{
                backgroundColor: tokens.color.statusFailedMuted,
                color: tokens.color.statusFailed,
                fontFamily: tokens.font.mono,
                fontSize: 11,
                fontWeight: 600,
                px: "6px",
                borderRadius: `${tokens.radius.sm}px`,
                minWidth: 18,
                textAlign: "center",
              }}
            >
              {dlq.length}
            </Box>
          </Stack>
        </AccordionSummary>
        <AccordionDetails sx={{ px: `${tokens.spacing.lg}px`, pt: 0 }}>
          {dlq.length === 0 ? (
            <Typography
              sx={{ color: tokens.color.textMuted, fontSize: 12, py: 1 }}
            >
              No failed jobs.
            </Typography>
          ) : (
            <Stack spacing={0.5}>
              {dlq.map((job) => (
                <Stack
                  key={job.id}
                  direction="row"
                  alignItems="center"
                  spacing={1}
                  sx={{ py: 0.75 }}
                >
                  <Typography
                    sx={{
                      fontFamily: tokens.font.mono,
                      fontSize: 11,
                      color: tokens.color.textSecondary,
                    }}
                  >
                    {job.id.slice(0, 8)}
                  </Typography>
                  <Box
                    sx={{
                      backgroundColor: tokens.color.statusFailedMuted,
                      color: tokens.color.statusFailed,
                      fontFamily: tokens.font.mono,
                      fontSize: 10,
                      fontWeight: 500,
                      px: "6px",
                      py: "2px",
                      borderRadius: `${tokens.radius.sm}px`,
                    }}
                  >
                    {job.job_type}
                  </Box>
                  <Box sx={{ flex: 1 }} />
                  <Button
                    variant="text"
                    size="small"
                    onClick={() => setPayload(job)}
                    sx={{
                      color: tokens.color.accentBlue,
                      fontSize: 11,
                      fontWeight: 500,
                      minWidth: 0,
                      px: 0.5,
                    }}
                  >
                    view payload
                  </Button>
                </Stack>
              ))}
            </Stack>
          )}
        </AccordionDetails>
      </Accordion>

      <Dialog
        open={!!payload}
        onClose={() => setPayload(null)}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: {
            backgroundColor: tokens.color.bgSurface,
            border: `1px solid ${tokens.color.borderSubtle}`,
          },
        }}
      >
        <DialogTitle sx={{ fontFamily: tokens.font.mono, fontSize: 13 }}>
          {payload?.id}
        </DialogTitle>
        <DialogContent>
          <Box
            component="pre"
            sx={{
              backgroundColor: tokens.color.bgElevated,
              color: tokens.color.textPrimary,
              fontFamily: tokens.font.mono,
              fontSize: 12,
              p: 2,
              borderRadius: `${tokens.radius.md}px`,
              overflowX: "auto",
              m: 0,
            }}
          >
            {payload ? JSON.stringify(payload.parameters, null, 2) : ""}
          </Box>
          {payload?.last_error ? (
            <Typography
              sx={{
                mt: 2,
                color: tokens.color.statusFailed,
                fontFamily: tokens.font.mono,
                fontSize: 12,
              }}
            >
              {payload.last_error}
            </Typography>
          ) : null}
        </DialogContent>
      </Dialog>
    </>
  );
}
