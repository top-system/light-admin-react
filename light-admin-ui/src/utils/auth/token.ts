/**
 * Local-storage-backed JWT store. A single key; no refresh-token flow yet.
 * Reads are defensive (SSR / disabled storage) so callers never see exceptions.
 */
const TOKEN_KEY = 'light-admin-token';

export function getToken(): string | null {
  try {
    return window.localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

export function setToken(token: string): void {
  try {
    window.localStorage.setItem(TOKEN_KEY, token);
  } catch {
    /* ignore */
  }
}

export function clearToken(): void {
  try {
    window.localStorage.removeItem(TOKEN_KEY);
  } catch {
    /* ignore */
  }
}
