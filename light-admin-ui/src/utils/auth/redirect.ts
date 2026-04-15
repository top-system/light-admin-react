/**
 * Login-redirect helpers. Kept small and side-effect-free at import time so
 * the request interceptor can call them during early bootstrap.
 */
export const LOGIN_PATH = '/user/login';

/** Build the URL to bounce to, preserving the current location as ?redirect=. */
export function buildLoginRedirect(): string {
  if (typeof window === 'undefined') return LOGIN_PATH;
  const { pathname, search, hash } = window.location;
  if (pathname === LOGIN_PATH) return LOGIN_PATH;
  const redirect = `${pathname}${search}${hash}`;
  return `${LOGIN_PATH}?redirect=${encodeURIComponent(redirect)}`;
}

/** Navigate to login, preserving redirect. Uses location.assign so interceptors can fire before the SPA router renders. */
export function redirectToLogin(): void {
  if (typeof window === 'undefined') return;
  window.location.href = buildLoginRedirect();
}

/**
 * Validate a user-supplied redirect target to prevent open-redirect abuse.
 * Only same-origin absolute paths are accepted.
 */
export function safeRedirectTarget(raw: string | null | undefined): string {
  if (!raw?.startsWith('/') || raw.startsWith('//')) return '/';
  try {
    const parsed = new URL(raw, window.location.origin);
    if (parsed.origin !== window.location.origin) return '/';
    return `${parsed.pathname}${parsed.search}${parsed.hash}`;
  } catch {
    return '/';
  }
}
