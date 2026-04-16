# Releases

## 1.29.3-tecnoges-incoming.8

Date: 2026-04-16

Changes:
- `/ui/st-904l` now decodes `RCONF` replies into structured fields (model, ID, mode, APN, IP:port, upload interval, timezone, status flags).
- Keeps raw RCONF text visible under the parsed block for audit/debug.

## 1.29.3-tecnoges-incoming.7

Date: 2026-04-16

Changes:
- `/ui/st-904l` now includes a preloaded command catalog with documentation panel.
- Selecting a command auto-fills the editable SMS payload template.
- Includes common ST-904L/ST-904LA commands with examples and operational notes.

## 1.29.3-tecnoges-incoming.6

Date: 2026-04-16

Changes:
- Adds tracker-focused UI at `/ui/st-904l` (and `/ui/st-904la`) to send commands and view incoming/outgoing timeline for one tracker number.
- Adds `sender` filter to `GET /api/3rdparty/v1/incoming` for consumer-side filtering by source number.

## 1.29.3-tecnoges-incoming.5

Date: 2026-04-15

Changes:
- `/health` now includes `checks["messages:hashing_interval_seconds"]` so you can verify the effective runtime hashing interval via curl.
- Useful to confirm whether `MESSAGES__HASHING_INTERVAL_SECONDS=0` is actually applied in Home Assistant runtime.

## 1.29.3-tecnoges-incoming.4

Date: 2026-04-15

Changes:
- Outgoing API responses now also include `contentPreview` (aligned with incoming endpoint).
- `/ui` now renders incoming/outgoing as readable cards (status, dates, sender/recipient, preview) instead of raw JSON blocks.

## 1.29.3-tecnoges-incoming.3

Date: 2026-04-15

Changes:
- `GET /api/3rdparty/v1/messages` and `GET /api/3rdparty/v1/messages/{id}` now include plain outgoing payload fields when available:
  - `message`
  - `textMessage`
  - `dataMessage`
  - `phoneNumbers`
- Hashed messages keep returning metadata/state, but plain content cannot be restored once already hashed.
- Message hashing can now be disabled by setting `MESSAGES__HASHING_INTERVAL_SECONDS=0` (used to preserve content for future API consumers and UI).

## 1.29.3-tecnoges-incoming.2

Date: 2026-04-15

Changes:
- Adds built-in web UI at `/ui` protected by Basic auth.
- UI supports:
  - listing incoming messages (`/api/3rdparty/v1/incoming`)
  - listing outgoing messages (`/api/3rdparty/v1/messages`)
  - sending SMS (`POST /api/3rdparty/v1/messages`)
- Keeps API behavior from `1.29.3-tecnoges-incoming.1`.

## 1.29.3-tecnoges-incoming.1

Date: 2026-04-14

Changes:
- Adds `POST /api/mobile/v1/incoming` so Android gateways can persist inbound SMS to the backend.
- Adds `GET /api/3rdparty/v1/incoming` so external applications can query inbound SMS history.
- Keeps outgoing SMS behavior unchanged.
- Intended for the Home Assistant build that tracks `Tecnogesfausti/server-ha-postgres` branch `postgres-ha-port`.

Notes:
- If the binary is built without `-ldflags`, the backend now still reports this release version through `/health`.
- To use the new incoming API end to end, the Android gateway APK must also include the matching incoming-sync changes.
