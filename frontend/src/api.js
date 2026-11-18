// Everything that talks to the Go API lives here, so components never call
// fetch directly.

// In dev the API runs separately on :8081; the production build is served by
// the Go server itself, so it uses same-origin requests. Override with
// VITE_API_BASE when the API lives elsewhere.
const API_BASE =
  import.meta.env.VITE_API_BASE ??
  (import.meta.env.DEV ? 'http://localhost:8081' : '');

// Error thrown for any non-2xx reply. `status` lets callers react to specific
// codes (401/403 = not allowed, 409 = conflict...).
export class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.status = status;
  }
}

// The API answers errors as plain text (http.Error), not JSON; accept both.
async function errorMessage(res) {
  const raw = (await res.text()).trim();
  try {
    return JSON.parse(raw).error || raw;
  } catch {
    return raw;
  }
}

// Sends one request and returns the parsed JSON body (null when empty, e.g.
// 204 No Content). `token` adds the Authorization header.
async function request(path, { method = 'GET', token, body } = {}) {
  const headers = {};
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body)
  });
  if (!res.ok) {
    throw new ApiError((await errorMessage(res)) || 'Erro na requisição', res.status);
  }
  const text = await res.text();
  return text ? JSON.parse(text) : null;
}

// --- Auth ---
export const login = (name, password) =>
  request('/login', { method: 'POST', body: { name, password } });

export const register = (name, password) =>
  request('/register', { method: 'POST', body: { name, password } });

// --- Courses ---
export const listCourses = () => request('/courses');

export const getCourse = (id) => request(`/courses/${id}`);

export const createCourse = (token, course) =>
  request('/courses', { method: 'POST', token, body: course });

export const deleteCourse = (token, id) =>
  request(`/courses/${id}`, { method: 'DELETE', token });

// --- Enrollments ---
export const listEnrollments = () => request('/enrollments');

export const enroll = (token, courseId) =>
  request('/enrollments', { method: 'POST', token, body: { course_id: courseId } });

export const unenroll = (token, enrollmentId) =>
  request(`/enrollments/${enrollmentId}`, { method: 'DELETE', token });

// --- Progress ---
export const getProgress = (token, courseId) =>
  request(`/courses/${courseId}/progress`, { token });

export const submitAnswer = (token, lessonId, choiceId) =>
  request(`/lessons/${lessonId}/answer`, { method: 'POST', token, body: { choice_id: choiceId } });
