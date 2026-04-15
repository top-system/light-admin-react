/**
 * In-memory dict-item cache. 5-minute stale time. Invalidated by the
 * `dict-change` WebSocket event (wired in app.tsx once WS is running).
 *
 * Design tradeoffs:
 * - Singleton module state. No SSR considerations — this is a CSR admin.
 * - Coalesces concurrent fetches: if 10 `<DictSelect>` instances mount at
 *   once with the same code, only one network call goes out.
 */
import { getDictItemOptions } from '@/services/light-admin/dict';
import type { DictItemOption } from '@/types/light-admin/domain';

type Entry = {
  value: DictItemOption[];
  expiresAt: number;
};

const STALE_MS = 5 * 60 * 1000;
const cache = new Map<string, Entry>();
const inflight = new Map<string, Promise<DictItemOption[]>>();
const subscribers = new Map<string, Set<() => void>>();

function notify(code: string) {
  subscribers.get(code)?.forEach((cb) => {
    try {
      cb();
    } catch {
      /* ignore */
    }
  });
}

export async function loadDictOptions(
  code: string,
  force = false,
): Promise<DictItemOption[]> {
  const now = Date.now();
  if (!force) {
    const cached = cache.get(code);
    if (cached && cached.expiresAt > now) return cached.value;
  }
  const existing = inflight.get(code);
  if (existing) return existing;

  const promise = getDictItemOptions(code)
    .then((value) => {
      cache.set(code, { value, expiresAt: Date.now() + STALE_MS });
      notify(code);
      return value;
    })
    .finally(() => {
      inflight.delete(code);
    });
  inflight.set(code, promise);
  return promise;
}

export function invalidateDict(code: string): void {
  cache.delete(code);
  notify(code);
}

export function invalidateAllDicts(): void {
  cache.clear();
  for (const code of subscribers.keys()) notify(code);
}

export function peekDictOptions(code: string): DictItemOption[] | undefined {
  return cache.get(code)?.value;
}

/** Subscribe to mutations of a specific dict code. Returns an unsubscribe. */
export function subscribeDict(code: string, cb: () => void): () => void {
  let bucket = subscribers.get(code);
  if (!bucket) {
    bucket = new Set();
    subscribers.set(code, bucket);
  }
  bucket.add(cb);
  return () => {
    subscribers.get(code)?.delete(cb);
  };
}
