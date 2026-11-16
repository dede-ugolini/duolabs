package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"duolabs/handler"
	"duolabs/model"

	"github.com/golang-jwt/jwt/v5"
)

func roleToken(t *testing.T, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"name":    "someone",
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	s, err := token.SignedString(testJWTSecret)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func callAdminProtected(t *testing.T, authHeader string) (int, bool) {
	t.Helper()
	called := false
	h := handler.AdminMiddleware(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})
	req := httptest.NewRequest(http.MethodPost, "/courses", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	h(w, req)
	return w.Code, called
}

func TestAdminMiddleware_NoToken(t *testing.T) {
	code, called := callAdminProtected(t, "")
	if code != http.StatusUnauthorized || called {
		t.Fatalf("expected 401 and handler not called, got %d (called=%v)", code, called)
	}
}

func TestAdminMiddleware_UserRoleForbidden(t *testing.T) {
	code, called := callAdminProtected(t, "Bearer "+roleToken(t, model.RoleUser))
	if code != http.StatusForbidden || called {
		t.Fatalf("expected 403 and handler not called, got %d (called=%v)", code, called)
	}
}

func TestAdminMiddleware_MissingRoleForbidden(t *testing.T) {
	code, called := callAdminProtected(t, "Bearer "+roleToken(t, ""))
	if code != http.StatusForbidden || called {
		t.Fatalf("expected 403 and handler not called, got %d (called=%v)", code, called)
	}
}

func TestAdminMiddleware_AdminAllowed(t *testing.T) {
	code, called := callAdminProtected(t, "Bearer "+roleToken(t, model.RoleAdmin))
	if code != http.StatusCreated || !called {
		t.Fatalf("expected 201 and handler called, got %d (called=%v)", code, called)
	}
}

func postJSON(h http.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	w := httptest.NewRecorder()
	h(w, req)
	return w
}

func loginToken(t *testing.T, h *handler.UserHandler, name, password string) string {
	t.Helper()
	w := postJSON(h.Login, "/login", map[string]string{"name": name, "password": password})
	if w.Code != http.StatusOK {
		t.Fatalf("login %q: expected 200, got %d", name, w.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	return resp["token"]
}

func TestEnsureAdmin_LoginPassesAdminMiddleware(t *testing.T) {
	s := setupTestUserStore(t)
	h := handler.NewUserHandler(s)
	if err := handler.EnsureAdmin(s, "admin", "admin123"); err != nil {
		t.Fatal(err)
	}
	code, called := callAdminProtected(t, "Bearer "+loginToken(t, h, "admin", "admin123"))
	if code != http.StatusCreated || !called {
		t.Fatalf("admin login should reach protected handler, got %d (called=%v)", code, called)
	}
}

func TestRegisteredUser_CannotPassAdminMiddleware(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))
	if w := postJSON(h.Register, "/register", map[string]string{"name": "bob", "password": "pw"}); w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", w.Code)
	}
	code, called := callAdminProtected(t, "Bearer "+loginToken(t, h, "bob", "pw"))
	if code != http.StatusForbidden || called {
		t.Fatalf("expected 403 and handler not called, got %d (called=%v)", code, called)
	}
}

func TestRegister_IgnoresRoleInBody(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))
	w := postJSON(h.Register, "/register", map[string]string{"name": "mallory", "password": "pw", "role": "admin"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", w.Code)
	}
	var user model.User
	_ = json.NewDecoder(w.Body).Decode(&user)
	if user.Role != model.RoleUser {
		t.Fatalf("expected role %q, got %q", model.RoleUser, user.Role)
	}
}

func TestRegister_DuplicateNameConflict(t *testing.T) {
	s := setupTestUserStore(t)
	h := handler.NewUserHandler(s)
	if err := handler.EnsureAdmin(s, "admin", "admin123"); err != nil {
		t.Fatal(err)
	}
	if w := postJSON(h.Register, "/register", map[string]string{"name": "admin", "password": "x"}); w.Code != http.StatusConflict {
		t.Fatalf("expected 409 registering the admin name, got %d", w.Code)
	}
}

func TestEnsureAdmin_PromotesAndResetsExistingUser(t *testing.T) {
	s := setupTestUserStore(t)
	h := handler.NewUserHandler(s)
	if w := postJSON(h.Register, "/register", map[string]string{"name": "admin", "password": "squatter"}); w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", w.Code)
	}
	if err := handler.EnsureAdmin(s, "admin", "admin123"); err != nil {
		t.Fatal(err)
	}
	if w := postJSON(h.Login, "/login", map[string]string{"name": "admin", "password": "squatter"}); w.Code != http.StatusUnauthorized {
		t.Fatalf("old password should stop working, got %d", w.Code)
	}
	code, called := callAdminProtected(t, "Bearer "+loginToken(t, h, "admin", "admin123"))
	if code != http.StatusCreated || !called {
		t.Fatalf("expected admin access, got %d (called=%v)", code, called)
	}
}
