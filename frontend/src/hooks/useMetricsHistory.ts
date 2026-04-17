import { useEffect, useRef, useState } from "react";

import type { Job } from "../jobs/types";

const BUCKETS = 6;
const BUCKET_MS = 5000;

export interface MetricsSnapshot {
  queueDepth: number;
  throughput: number;
  successRate: number;
  errorRate: number;
  queueHistory: number[];
  throughputHistory: number[];
  successHistory: number[];
  errorHistory: number[];
}

export function useMetricsHistory(jobs: Job[]): MetricsSnapshot {
  const queueRef = useRef<number[]>(new Array(BUCKETS).fill(0));
  const throughputRef = useRef<number[]>(new Array(BUCKETS).fill(0));
  const successRef = useRef<number[]>(new Array(BUCKETS).fill(0));
  const errorRef = useRef<number[]>(new Array(BUCKETS).fill(0));
  const lastSeenRef = useRef<Map<string, string>>(new Map());

  const [tick, setTick] = useState(0);

  const pending = jobs.filter((j) => j.status === "pending" || j.status === "retrying").length;
  const running = jobs.filter((j) => j.status === "running").length;
  const succeeded = jobs.filter((j) => j.status === "succeeded").length;
  const failed = jobs.filter((j) => j.status === "failed").length;
  const total = succeeded + failed;
  const successRate = total === 0 ? 0 : Math.round((succeeded / total) * 100);
  const errorRate = total === 0 ? 0 : Math.round((failed / total) * 100);

  useEffect(() => {
    let completedDelta = 0;
    let succeededDelta = 0;
    let failedDelta = 0;
    const seen = lastSeenRef.current;
    for (const job of jobs) {
      const prev = seen.get(job.id);
      if (prev !== job.status) {
        if (prev !== undefined && prev !== "succeeded" && prev !== "failed") {
          if (job.status === "succeeded") {
            completedDelta += 1;
            succeededDelta += 1;
          } else if (job.status === "failed") {
            completedDelta += 1;
            failedDelta += 1;
          }
        }
        seen.set(job.id, job.status);
      }
    }
    if (completedDelta > 0 || succeededDelta > 0 || failedDelta > 0) {
      const tLast = throughputRef.current.length - 1;
      throughputRef.current[tLast] += completedDelta;
      successRef.current[tLast] += succeededDelta;
      errorRef.current[tLast] += failedDelta;
    }
  }, [jobs]);

  useEffect(() => {
    const id = window.setInterval(() => {
      queueRef.current = [...queueRef.current.slice(1), pending + running];
      throughputRef.current = [...throughputRef.current.slice(1), 0];
      successRef.current = [...successRef.current.slice(1), 0];
      errorRef.current = [...errorRef.current.slice(1), 0];
      setTick((t) => t + 1);
    }, BUCKET_MS);
    return () => window.clearInterval(id);
  }, [pending, running]);

  void tick;

  return {
    queueDepth: pending + running,
    throughput: throughputRef.current.reduce((a, b) => a + b, 0),
    successRate,
    errorRate,
    queueHistory: [...queueRef.current],
    throughputHistory: [...throughputRef.current],
    successHistory: [...successRef.current],
    errorHistory: [...errorRef.current],
  };
}
