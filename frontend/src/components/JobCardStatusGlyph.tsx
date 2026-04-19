import { motion, useReducedMotion } from "motion/react";

import type { JobStatus } from "../jobs/types";
import { AlertIcon, CheckIcon, RefreshIcon } from "./icons";

const svgBase = {
  width: 16,
  height: 16,
  viewBox: "0 0 24 24",
  fill: "none" as const,
  stroke: "currentColor",
  strokeWidth: 2,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
};

interface JobCardStatusGlyphProps {
  status: JobStatus;
  color: string;
}

export function JobCardStatusGlyph({ status, color }: JobCardStatusGlyphProps) {
  const prefersReducedMotion = useReducedMotion();
  const reduced = !!prefersReducedMotion;

  if (status === "retrying") {
    if (reduced) {
      return (
        <span style={{ color, display: "flex", lineHeight: 0 }}>
          <RefreshIcon />
        </span>
      );
    }
    return (
      <motion.div
        initial={{ rotate: 0 }}
        animate={{ rotate: 720 }}
        transition={{ duration: 1.1, ease: "linear" }}
        style={{ color, display: "flex", lineHeight: 0 }}
      >
        <RefreshIcon />
      </motion.div>
    );
  }

  if (status === "succeeded") {
    if (reduced) {
      return (
        <span style={{ color, display: "flex", lineHeight: 0 }}>
          <CheckIcon />
        </span>
      );
    }
    return (
      <motion.svg
        {...svgBase}
        style={{ color, display: "block" }}
        initial={{ opacity: 0.6 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.35 }}
      >
        <motion.polyline
          points="20 6 9 17 4 12"
          initial={{ pathLength: 0 }}
          animate={{ pathLength: 1 }}
          transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
        />
      </motion.svg>
    );
  }

  if (status === "failed") {
    if (reduced) {
      return (
        <span style={{ color, display: "flex", lineHeight: 0 }}>
          <AlertIcon />
        </span>
      );
    }
    return (
      <motion.div
        initial={{ x: 0, rotate: 0 }}
        animate={{
          x: [0, -3, 3, -3, 3, 0],
          rotate: [0, -4, 4, -4, 0],
        }}
        transition={{
          delay: 0.45,
          duration: 0.38,
          ease: "easeInOut",
        }}
        style={{ color, display: "flex", lineHeight: 0 }}
      >
        <svg {...svgBase}>
          <motion.circle
            cx={12}
            cy={12}
            r={10}
            initial={{ opacity: 0, scale: 0.88 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.15, ease: "easeOut" }}
            style={{ transformOrigin: "12px 12px" }}
          />
          <motion.line
            x1={12}
            y1={8}
            x2={12}
            y2={12}
            initial={{ opacity: 0, pathLength: 0, y1: 10, y2: 10 }}
            animate={{ opacity: 1, pathLength: 1, y1: 8, y2: 12 }}
            transition={{ delay: 0.2, duration: 0.22, ease: "easeOut" }}
          />
          <motion.line
            x1={12}
            y1={16}
            x2={12.01}
            y2={16}
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.32, duration: 0.12, ease: "easeOut" }}
            style={{ transformOrigin: "12px 16px" }}
          />
        </svg>
      </motion.div>
    );
  }

  return null;
}
