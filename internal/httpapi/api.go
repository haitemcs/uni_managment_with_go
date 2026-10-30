package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/haitemcs/uni_managment_with_go/internal/auth"
)

type API struct {
	db     *pgxpool.Pool
	tokens *auth.TokenManager
}

type contextKey string

const claimsKey contextKey = "claims"

func New(db *pgxpool.Pool, tokens *auth.TokenManager) *API {
	return &API{db: db, tokens: tokens}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.root)
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("POST /auth/register", a.register)
	mux.HandleFunc("POST /auth/login", a.login)
	mux.Handle("GET /auth/me", a.requireAuth(http.HandlerFunc(a.me)))
	mux.Handle("GET /departments", a.requireAuth(http.HandlerFunc(a.listDepartments)))
	mux.Handle("POST /departments", a.requireRoles("admin")(http.HandlerFunc(a.createDepartment)))
	mux.Handle("GET /students", a.requireAuth(http.HandlerFunc(a.listStudents)))
	mux.Handle("POST /students", a.requireRoles("admin")(http.HandlerFunc(a.createStudent)))
	mux.Handle("GET /courses", a.requireAuth(http.HandlerFunc(a.listCourses)))
	mux.Handle("POST /courses", a.requireRoles("admin", "teacher")(http.HandlerFunc(a.createCourse)))
	mux.Handle("POST /enrollments", a.requireRoles("admin", "teacher")(http.HandlerFunc(a.enrollStudent)))
	mux.Handle("POST /grades", a.requireRoles("admin", "teacher")(http.HandlerFunc(a.addGrade)))
	mux.Handle("GET /reports/students/{id}/average", a.requireAuth(http.HandlerFunc(a.studentAverage)))
	mux.Handle("GET /reports/departments", a.requireAuth(http.HandlerFunc(a.departmentReport)))
	mux.Handle("GET /reports/courses/enrollments", a.requireAuth(http.HandlerFunc(a.courseEnrollmentReport)))
	return withCORS(mux)
}

func (a *API) root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name": "University Management API",
		"endpoints": []string{"/health", "/auth/register", "/auth/login", "/auth/me", "/departments", "/students", "/courses", "/enrollments", "/grades", "/reports"},
	})
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	if err := a.db.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "connected"})
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
