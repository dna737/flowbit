# AI dispatcher

See also: [Scheduler](30-scheduler.md) · [Vision and demo](00-vision-and-demo.md) · [Build checklist](BUILD-CHECKLIST.md)

---

## Block 4 — AI dispatcher

- `POST /dispatch` endpoint accepts plain English
- Calls Gemini API (`gemini-3-flash-preview` by default; fallback chain on quota/5xx), extracts `job_type` + `parameters` as JSON
- Forwards structured payload to `POST /jobs`
- Test with: “send an email”, “resize this image”, “scrape this URL”

---

## Integration

The dispatcher validates or normalizes Gemini output, then submits the same shape the scheduler expects from a direct `POST /jobs` call. Keep the scheduler runnable and testable without the AI layer.
