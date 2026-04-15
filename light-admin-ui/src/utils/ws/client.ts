/**
 * Browser WebSocket client for the Light Admin plain-JSON protocol.
 *
 * Heartbeat: server pings every 30s (`{type:"ping"}`) and disconnects after
 * 60s of silence. We reply to every ping with a pong, and independently
 * emit our own ping every 30s. Missing two consecutive pongs from the
 * server triggers a reconnect.
 *
 * Reconnect: exponential backoff (1s, 2s, 4s, ..., cap 30s). The client
 * stops reconnecting only on explicit `close()`.
 */

export type WsFrame<T = unknown> = {
  type: string;
  data?: T;
  id?: string;
  ts?: number;
};

export type WsHandler<T = unknown> = (frame: WsFrame<T>) => void;

export type WsClient = {
  readonly url: string;
  on: <T = unknown>(type: string, handler: WsHandler<T>) => () => void;
  emit: (type: string, data?: unknown) => void;
  close: () => void;
  isOpen: () => boolean;
};

export type WsClientOptions = {
  /** Base ws URL. Token is appended as `?token=...`. */
  url: string;
  /** Bearer JWT minted via `/auth/login`. */
  token: string;
  /** Optional logger — defaults to console.debug / console.warn. */
  logger?: {
    debug: (msg: string, ...args: unknown[]) => void;
    warn: (msg: string, ...args: unknown[]) => void;
  };
};

const PING_INTERVAL_MS = 30_000;
const MAX_MISSED_PONGS = 2;
const RECONNECT_BASE_MS = 1000;
const RECONNECT_CAP_MS = 30_000;

function buildWsUrl(base: string, token: string): string {
  // If base already has a scheme, respect it. Otherwise derive from page.
  let full = base;
  if (!/^wss?:\/\//i.test(full)) {
    if (typeof window !== 'undefined') {
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
      const host = window.location.host;
      const path = full.startsWith('/') ? full : `/${full}`;
      full = `${proto}://${host}${path}`;
    }
  }
  const sep = full.includes('?') ? '&' : '?';
  return `${full}${sep}token=${encodeURIComponent(token)}`;
}

export function createWsClient(opts: WsClientOptions): WsClient {
  const logger = opts.logger ?? {
    debug: (...a) => console.debug('[ws]', ...a),
    warn: (...a) => console.warn('[ws]', ...a),
  };

  let socket: WebSocket | null = null;
  let closed = false;
  let reconnectAttempt = 0;
  let pingTimer: ReturnType<typeof setInterval> | null = null;
  let missedPongs = 0;
  const handlers = new Map<string, Set<WsHandler>>();

  const resolvedUrl = buildWsUrl(opts.url, opts.token);

  const dispatch = (frame: WsFrame) => {
    const bucket = handlers.get(frame.type);
    if (!bucket) return;
    for (const h of bucket) {
      try {
        (h as WsHandler)(frame);
      } catch (err) {
        logger.warn('handler threw', frame.type, err);
      }
    }
  };

  const sendRaw = (frame: WsFrame) => {
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(JSON.stringify(frame));
  };

  const scheduleReconnect = () => {
    if (closed) return;
    const backoff = Math.min(
      RECONNECT_CAP_MS,
      RECONNECT_BASE_MS * 2 ** Math.max(0, reconnectAttempt - 1),
    );
    reconnectAttempt += 1;
    logger.debug(`reconnect in ${backoff}ms (attempt ${reconnectAttempt})`);
    window.setTimeout(connect, backoff);
  };

  const clearPingTimer = () => {
    if (pingTimer !== null) {
      clearInterval(pingTimer);
      pingTimer = null;
    }
  };

  const startPingLoop = () => {
    clearPingTimer();
    missedPongs = 0;
    pingTimer = setInterval(() => {
      if (missedPongs >= MAX_MISSED_PONGS) {
        logger.warn('missed pong threshold, forcing reconnect');
        socket?.close();
        return;
      }
      missedPongs += 1;
      sendRaw({ type: 'ping' });
    }, PING_INTERVAL_MS);
  };

  const connect = () => {
    if (closed) return;
    try {
      socket = new WebSocket(resolvedUrl);
    } catch (err) {
      logger.warn('failed to construct WebSocket', err);
      scheduleReconnect();
      return;
    }
    socket.onopen = () => {
      logger.debug('open');
      reconnectAttempt = 0;
      startPingLoop();
    };
    socket.onmessage = (event) => {
      let frame: WsFrame;
      try {
        frame = JSON.parse(event.data as string) as WsFrame;
      } catch {
        logger.warn('malformed frame', event.data);
        return;
      }
      if (frame.type === 'pong') {
        missedPongs = 0;
        return;
      }
      if (frame.type === 'ping') {
        sendRaw({ type: 'pong' });
        return;
      }
      dispatch(frame);
    };
    socket.onerror = (event) => {
      logger.debug('error', event);
    };
    socket.onclose = () => {
      logger.debug('close');
      clearPingTimer();
      socket = null;
      if (!closed) scheduleReconnect();
    };
  };

  connect();

  return {
    get url() {
      return resolvedUrl;
    },
    on<T = unknown>(type: string, handler: WsHandler<T>) {
      let bucket = handlers.get(type);
      if (!bucket) {
        bucket = new Set();
        handlers.set(type, bucket);
      }
      bucket.add(handler as WsHandler);
      return () => {
        handlers.get(type)?.delete(handler as WsHandler);
      };
    },
    emit(type: string, data?: unknown) {
      sendRaw({ type, data: data as WsFrame['data'] });
    },
    close() {
      closed = true;
      clearPingTimer();
      socket?.close();
      socket = null;
    },
    isOpen() {
      return socket?.readyState === WebSocket.OPEN;
    },
  };
}

// --- Singleton ------------------------------------------------------------

let singleton: WsClient | null = null;

export function startWsClient(opts: WsClientOptions): WsClient {
  stopWsClient();
  singleton = createWsClient(opts);
  return singleton;
}

export function getWsClient(): WsClient | null {
  return singleton;
}

export function stopWsClient(): void {
  singleton?.close();
  singleton = null;
}
