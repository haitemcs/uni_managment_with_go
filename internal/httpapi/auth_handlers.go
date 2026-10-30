package httpapi

import (
	"net/http"
	"strings"

	"github.com/haitemcs/uni_managment_with_go/internal/auth"
)

type registerRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	FullName       string `json:"full_name"`
	StudentNumber  string `json:"student_number"`
	DepartmentID   int64  `json:"department_id"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)
	if req.Email == "" || len(req.Password) < 8 || req.FullName == "" || req.StudentNumber == "" || req.DepartmentID < 1 {
		writeError(w, http.StatusBadRequest, "email, full_name, student_number, department_id and password (8+ characters) are required")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not secure password")
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start registration")
		return
	}
	defer tx.Rollback(r.Context())

	var userID int64
	err = tx.QueryRow(r.Context(),
		"INSERT INTO users(email, password_hash, full_name, role) VALUES($1,$2,$3,'student') RETURNING id",
		req.Email, hash, req.FullName).Scan(&userID)
	if err != nil {
		writeError(w, http.StatusConflict, "email already exists or user data is invalid")
		return
	}
	_, err = tx.Exec(r.Context(),
		"INSERT INTO students(user_id, student_number, department_id) VALUES($1,$2,$3)",
		userID, strings.TrimSpace(req.StudentNumber), req.DepartmentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not create student profile; check department and student number")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "registration failed")
		return
	}
	token, err := a.tokens.Generate(userID, req.Email, "student")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": userID, "email": req.Email, "full_name": req.FullName, "role": "student", "access_token": token, "token_type": "bearer"})
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	var id int64
	var email, fullName, role, hash string
	err := a.db.QueryRow(r.Context(),
		"SELECT id,email,full_name,role,password_hash FROM users WHERE email=$1",
		strings.ToLower(strings.TrimSpace(req.Email))).Scan(&id, &email, &fullName, &role, &hash)
	if err != nil || !auth.CheckPassword(hash, req.Password) {
		writeError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}
	token, err := a.tokens.Generate(id, email, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"access_token": token, "token_type": "bearer", "user": map[string]any{"id": id, "email": email, "full_name": fullName, "role": role}})
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := currentClaims(r.Context().Value(claimsKey))
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid authentication context")
		return
	}
	var fullName string
	err := a.db.QueryRow(r.Context(), "SELECT full_name FROM users WHERE id=$1", claims.UserID).Scan(&fullName)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": claims.UserID, "email": claims.Email, "full_name": fullName, "role": claims.Role})
}
