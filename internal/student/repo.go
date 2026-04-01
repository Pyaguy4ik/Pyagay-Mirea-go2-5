package student

import (
    "database/sql"
    "errors"
    "fmt"
)

var ErrStudentNotFound = errors.New("student not found")

type Repo struct {
    db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
    return &Repo{db: db}
}

// ⚠️ ОПАСНЫЙ ПРИМЕР - НЕ ИСПОЛЬЗОВАТЬ В ПРОДАКШЕНЕ ⚠️
// Этот метод демонстрирует уязвимость к SQL-инъекциям
func (r *Repo) UnsafeGetByID(rawID string) (*Student, error) {
    // ВНИМАНИЕ: Конкатенация строк с пользовательским вводом опасна!
    query := "SELECT id, full_name, study_group, email FROM students WHERE id = " + rawID
    
    fmt.Printf("[UNSAFE] Executing query: %s\n", query)
    
    row := r.db.QueryRow(query)
    
    var st Student
    err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrStudentNotFound
        }
        return nil, err
    }
    return &st, nil
}

// ✅ БЕЗОПАСНЫЙ ПРИМЕР - Использование параметризованного запроса
func (r *Repo) GetByID(id int64) (*Student, error) {
    // Используем параметризованный запрос с плейсхолдером $1
    row := r.db.QueryRow(
        "SELECT id, full_name, study_group, email FROM students WHERE id = $1",
        id,
    )
    
    var st Student
    err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrStudentNotFound
        }
        return nil, err
    }
    return &st, nil
}

// ✅ БЕЗОПАСНЫЙ ПРИМЕР - Использование подготовленного выражения (prepared statement)
func (r *Repo) PrepareGetByID() (*sql.Stmt, error) {
    return r.db.Prepare("SELECT id, full_name, study_group, email FROM students WHERE id = $1")
}

// ✅ БЕЗОПАСНЫЙ ПРИМЕР - Поиск по email
func (r *Repo) GetByEmail(email string) (*Student, error) {
    row := r.db.QueryRow(
        "SELECT id, full_name, study_group, email FROM students WHERE email = $1",
        email,
    )
    
    var st Student
    err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrStudentNotFound
        }
        return nil, err
    }
    return &st, nil
}

// ✅ БЕЗОПАСНЫЙ ПРИМЕР - Подготовленное выражение для поиска по email
func (r *Repo) PrepareGetByEmail() (*sql.Stmt, error) {
    return r.db.Prepare("SELECT id, full_name, study_group, email FROM students WHERE email = $1")
}
