# WebSocket — Plain JSON Protocol

As of the 2026 migration the server speaks a minimal JSON-over-WebSocket
protocol. STOMP has been retired. Every client interaction is a single JSON
object (a **frame**) sent as a text message.

## Endpoint

```
GET /ws?token=<jwt>
```

- Authentication is performed **before** the HTTP 101 upgrade. A missing or
  invalid `token` returns HTTP 401 and the connection is never upgraded.
- The JWT is the same token returned by `POST /api/v1/auth/login` and sent as
  `Authorization: Bearer <token>` on REST calls. `Authorization` headers are
  also accepted during upgrade, but most browsers cannot set them on a
  WebSocket handshake, hence the query-parameter form.

## Frame shape

```ts
type Frame = {
  type: string;      // discriminator — see table below
  data?: unknown;    // type-specific payload, JSON
  id?: string;       // server-assigned message ID
  ts?: number;       // server-assigned unix millis
};
```

## Server → client frame types

| `type`          | `data` shape                                                 | When emitted                                                   |
|-----------------|--------------------------------------------------------------|----------------------------------------------------------------|
| `ping`          | _none_                                                       | Every 30s. Client must answer with `{type:"pong"}`.            |
| `online-count`  | `number` (total active sessions)                             | On every connect/disconnect.                                   |
| `dict-change`   | `{ dictCode: string; timestamp: number }`                    | After a dict (or dict item) mutation.                          |
| `notice`        | `{ content: string; timestamp: number }` or string payload   | `POST /websocket/sendToAll`, notice module.                    |
| `message`       | `{ sender: string; content: string; timestamp: number }`     | `POST /websocket/sendToUser`.                                  |
| `system`        | `{ sender: "System"; content: string; timestamp: number }`   | Server-initiated system announcements.                         |

## Client → server frame types

| `type`   | `data`  | Semantics                                                 |
|----------|---------|-----------------------------------------------------------|
| `pong`   | _none_  | Heartbeat acknowledgement. Any other frame is also valid. |
| `ping`   | _none_  | Optional client-initiated liveness probe; server replies `{type:"pong"}`. |

Any frame whose `type` is not recognised is **ignored** (a debug log line is
emitted server-side). The canonical way for authenticated users to publish is
still the HTTP API under `/api/v1/websocket/*`.

## Heartbeat contract

- Server pings every **30 s**.
- Server drops connections that do not produce any inbound frame (pong, WS
  control pong, or otherwise) for **60 s**.
- Clients **should** answer every `ping` with `{"type":"pong"}`. WS-level pong
  control frames are also accepted and reset the read deadline.

## Reconnect

Clients should treat any unexpected close as transient and reconnect with
exponential backoff (the reference client in `ant-design-pro-admin/src/utils/ws/client.ts`
uses 1s / 2s / 4s ... capped at 30s).

## HTTP companion endpoints

| Method | Path                              | Purpose                                    |
|--------|-----------------------------------|--------------------------------------------|
| POST   | `/api/v1/websocket/sendToAll`     | Broadcast `notice` frame to everyone.      |
| POST   | `/api/v1/websocket/sendToUser`    | Send `message` frame to one user.          |
| POST   | `/api/v1/websocket/dict-change`   | Broadcast `dict-change` invalidation.      |
| GET    | `/api/v1/websocket/online-users`  | Presence snapshot.                         |
| GET    | `/api/v1/websocket/online-count`  | Distinct-user count.                       |

## Example (browser)

```ts
const ws = new WebSocket(`ws://localhost:9999/ws?token=${jwt}`);
ws.onmessage = (ev) => {
  const frame = JSON.parse(ev.data);
  if (frame.type === 'ping') {
    ws.send(JSON.stringify({ type: 'pong' }));
    return;
  }
  // handle by type
};
```
