package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/store"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestUserStore(t *testing.T) store.UserStore {
	t.Helper()
	tmp, err := os.CreateTemp("", "duolabs-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	db, err := gorm.Open(sqlite.Open(tmp.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatal(err)
	}
	return store.NewSQLiteUserStore(db)
}

func TestGetUsers_Empty(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	h.GetAll(w, req)

	var users []model.User
	json.NewDecoder(w.Body).Decode(&users)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}

	if len(users) != 0 {
		t.Fatalf("expected empty array, got %d courses", len(users))
	}
}

func TestGetUsers_ReturnsAll(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))

	for _, u := range []model.User{
		{Name: "Dede", Password: "123"},
		{Name: "Jhon", Password: "456"},
	} {
		body, _ := json.Marshal(u)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.Create(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, w.Code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	h.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var users []model.User
	json.NewDecoder(w.Body).Decode(&users)
	if len(users) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(users))
	}
}

func TestPostUser_Success(t *testing.T) {
	name := "dede"
	password := "123"
	h := handler.NewUserHandler(setupTestUserStore(t))
	body, _ := json.Marshal(model.User{Name: name, Password: password})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, w.Code)
	}

	var user model.User
	json.NewDecoder(w.Body).Decode(&user)
	if user.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if user.Name != name {
		t.Fatalf("expected %q, got %q", name, user.Name)
	}
}

func TestRegister_Success(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))
	body, _ := json.Marshal(map[string]string{"name": "alice", "password": "secret123"})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var user model.User
	json.NewDecoder(w.Body).Decode(&user)
	if user.Name != "alice" {
		t.Fatalf("expected name %q, got %q", "alice", user.Name)
	}
	if user.Password != "" {
		t.Fatal("password should not be in response")
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))

	body, _ := json.Marshal(map[string]string{"name": "bob", "password": "pass123"})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	regW := httptest.NewRecorder()
	h.Register(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", regW.Code)
	}

	loginBody, _ := json.Marshal(map[string]string{"name": "bob", "password": "pass123"})
	loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	loginW := httptest.NewRecorder()
	h.Login(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", loginW.Code)
	}
	var resp map[string]string
	json.NewDecoder(loginW.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))

	body, _ := json.Marshal(map[string]string{"name": "charlie", "password": "correct"})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	regW := httptest.NewRecorder()
	h.Register(regW, regReq)

	loginBody, _ := json.Marshal(map[string]string{"name": "charlie", "password": "wrong"})
	loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	loginW := httptest.NewRecorder()
	h.Login(loginW, loginReq)

	if loginW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", loginW.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	h := handler.NewUserHandler(setupTestUserStore(t))
	body, _ := json.Marshal(map[string]string{"name": "nobody", "password": "pass"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
