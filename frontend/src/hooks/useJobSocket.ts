import { useEffect, useRef, useState, type Dispatch } from "react";

import type { Job, JobsAction, SnapshotMessage } from "../jobs/types";

const defaultWsUrl  = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/ws`;
const maxReconnectDelayMs = 30000;

export type ConnectionStatus =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected";

export function useJobSocket(dispatch: Dispatch<JobsAction>) {
  const reconnectDelayRef = useRef(1000);
  const reconnectTimerRef = useRef<number | null>(null);
  const manualCloseRef = useRef(false);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");

  useEffect(() => {
    let socket: WebSocket | null = null;

    const clearReconnectTimer = () => {
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
    };

    const connect = () => {
      clearReconnectTimer();
      setStatus(socket ? "reconnecting" : "connecting");

      socket = new WebSocket(defaultWsUrl);

      socket.onopen = () => {
        reconnectDelayRef.current = 1000;
        setStatus("connected");
      };

      socket.onmessage = (event) => {
        const parsed = JSON.parse(event.data) as SnapshotMessage | Job;
        if ("type" in parsed && parsed.type === "snapshot") {
          dispatch({ type: "SNAPSHOT", jobs: parsed.jobs });
          return;
        }

        dispatch({ type: "UPSERT", job: parsed as Job });
      };

      socket.onclose = () => {
        if (manualCloseRef.current) {
          setStatus("disconnected");
          return;
        }

        setStatus("reconnecting");
        reconnectTimerRef.current = window.setTimeout(connect, reconnectDelayRef.current);
        reconnectDelayRef.current = Math.min(
          reconnectDelayRef.current * 2,
          maxReconnectDelayMs,
        );
      };

      socket.onerror = () => {
        socket?.close();
      };
    };

    connect();

    return () => {
      manualCloseRef.current = true;
      clearReconnectTimer();
      socket?.close();
    };
  }, [dispatch]);

  return status;
}
