/**
 * React bindings for the singleton WebSocket client. Handlers auto-unsubscribe
 * on unmount.
 */
import { useEffect } from 'react';
import { getWsClient, type WsFrame, type WsHandler } from './client';

export function useWsEvent<T = unknown>(
  type: string,
  handler: WsHandler<T>,
  deps: unknown[] = [],
): void {
  useEffect(() => {
    const client = getWsClient();
    if (!client) return;
    const off = client.on<T>(type, handler);
    return off;
    // Intentional — consumers pass their own deps. The `handler` is allowed
    // to close over per-render values; we rebind when deps change.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [type, ...deps]);
}

export type { WsFrame, WsHandler };
