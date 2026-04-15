# Releases

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
