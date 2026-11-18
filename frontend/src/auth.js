const TOKEN_KEY = 'duolabs_token';

// Reads the claims from the JWT payload. Only used to adapt the UI (admin
// buttons, which courses the user is enrolled in); the server enforces
// the permissions.
export function claimsFromToken(token) {
  try {
    const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
    return JSON.parse(atob(payload));
  } catch {
    return {};
  }
}

// Returns the stored token, or '' when there is none or it has expired.
export function loadToken() {
  const token = localStorage.getItem(TOKEN_KEY) || '';
  if (!token) return '';
  const { exp } = claimsFromToken(token);
  if (!exp || exp * 1000 < Date.now()) {
    clearToken();
    return '';
  }
  return token;
}

export const saveToken = (token) => localStorage.setItem(TOKEN_KEY, token);

export const clearToken = () => localStorage.removeItem(TOKEN_KEY);
