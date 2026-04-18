export const WATCHLIST_STORAGE_KEY = "flowbit_dispatch_watchlist_v1";
export const MAX_WATCHLIST_ENTRIES = 15;

export interface DispatchWatchlistEntry {
  jobId: string;
  prompt: string;
  addedAt: number;
}

export function loadWatchlistFromSession(): DispatchWatchlistEntry[] {
  if (typeof sessionStorage === "undefined") return [];
  try {
    const raw = sessionStorage.getItem(WATCHLIST_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed
      .filter(
        (row): row is DispatchWatchlistEntry =>
          row !== null &&
          typeof row === "object" &&
          typeof (row as DispatchWatchlistEntry).jobId === "string" &&
          typeof (row as DispatchWatchlistEntry).prompt === "string" &&
          typeof (row as DispatchWatchlistEntry).addedAt === "number",
      )
      .slice(0, MAX_WATCHLIST_ENTRIES);
  } catch {
    return [];
  }
}

export function saveWatchlistToSession(entries: DispatchWatchlistEntry[]): void {
  if (typeof sessionStorage === "undefined") return;
  try {
    sessionStorage.setItem(WATCHLIST_STORAGE_KEY, JSON.stringify(entries));
  } catch {
    // ignore quota / private mode
  }
}

/** Newest first; dedupes by jobId (keeps newest). */
export function prependWatchlistEntry(
  current: DispatchWatchlistEntry[],
  jobId: string,
  prompt: string,
): DispatchWatchlistEntry[] {
  const next: DispatchWatchlistEntry[] = [
    { jobId, prompt, addedAt: Date.now() },
    ...current.filter((e) => e.jobId !== jobId),
  ];
  return next.slice(0, MAX_WATCHLIST_ENTRIES);
}
