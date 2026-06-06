import { useEffect, useRef, useState, type Dispatch } from "react";

import type { AuthTokenGetter } from "../api/client";
import type { Job, JobsAction, SnapshotMessage } from "../jobs/types";

const defaultWsUrl =
  import.meta.env.VITE_WS_URL ||
  `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}/api/ws`;
const maxReconnectDelayMs = 30000;

export type ConnectionStatus =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected";

export function useJobSocket(
  dispatch: Dispatch<JobsAction>,
  getToken: AuthTokenGetter,
  enabled: boolean,
) {
  const reconnectDelayRef = useRef(1000);
  const reconnectTimerRef = useRef<number | null>(null);
  const manualCloseRef = useRef(false);
  const [status, setStatus] = useState<ConnectionStatus>(enabled ? "connecting" : "disconnected");

  useEffect(() => {
    let socket: WebSocket | null = null;
    let cancelled = false;
    manualCloseRef.current = false;

    const clearReconnectTimer = () => {
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
    };

    const connect = async () => {
      if (!enabled) {
        setStatus("disconnected");
        return;
      }
      clearReconnectTimer();
      setStatus(socket ? "reconnecting" : "connecting");

      const token = await getToken();
      if (cancelled || !token) {
        setStatus("disconnected");
        return;
      }

      const url = new URL(defaultWsUrl, window.location.href);
      url.searchParams.set("token", token);
      socket = new WebSocket(url.toString());

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
        reconnectTimerRef.current = window.setTimeout(() => void connect(), reconnectDelayRef.current);
        reconnectDelayRef.current = Math.min(
          reconnectDelayRef.current * 2,
          maxReconnectDelayMs,
        );
      };

      socket.onerror = () => {
        socket?.close();
      };
    };

    void connect();

    return () => {
      cancelled = true;
      manualCloseRef.current = true;
      clearReconnectTimer();
      socket?.close();
    };
  }, [dispatch, enabled, getToken]);

  return status;
}
