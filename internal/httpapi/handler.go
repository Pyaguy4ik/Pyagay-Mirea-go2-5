package httpapi

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    
    "example.com/pz5-security/internal/student"
)

type Handler struct {
    repo         *student.Repo
    stmtByID     *sql.Stmt
    stmtByEmail  *sql.Stmt
}

func NewHandler(repo *student.Repo, stmtByID, stmtByEmail *sql.Stmt) *Handler {
    return &Handler{
        repo:        repo,
        stmtByID:    stmtByID,
        stmtByEmail: stmtByEmail,
    }
}

// Health проверяет работоспособность сервиса
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
        "scheme": "https",
    })
}

// GetStudentByID возвращает студента по ID (через prepared statement)
func (h *Handler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    rawID := r.URL.Query().Get("id")
    if rawID == "" {
        http.Error(w, "id is required", http.StatusBadRequest)
        return
    }
    
    id, err := strconv.ParseInt(rawID, 10, 64)
    if err != nil || id <= 0 {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    
    var st student.Student
    err = h.stmtByID.QueryRow(id).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "student not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(st)
}

// GetStudentByEmail возвращает студента по email (через prepared statement)
func (h *Handler) GetStudentByEmail(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    email := r.URL.Query().Get("email")
    if email == "" {
        http.Error(w, "email is required", http.StatusBadRequest)
        return
    }
    
    var st student.Student
    err := h.stmtByEmail.QueryRow(email).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "student not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(st)
}

// GetStudentUnsafe ДЕМОНСТРАЦИЯ УЯЗВИМОСТИ - НЕ ИСПОЛЬЗОВАТЬ
// Этот эндпоинт показывает, как работает SQL-инъекция
func (h *Handler) GetStudentUnsafe(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    rawID := r.URL.Query().Get("id")
    if rawID == "" {
        http.Error(w, "id is required", http.StatusBadRequest)
        return
    }
    
    // ⚠️ ОПАСНО: использование небезопасного метода
    st, err := h.repo.UnsafeGetByID(rawID)
    if err != nil {
        if err == student.ErrStudentNotFound {
            http.Error(w, "student not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(st)
}
