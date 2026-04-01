package main

import (
    "crypto/tls"
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"
    
    _ "github.com/lib/pq"
)

type Student struct {
    ID         int64  `json:"id"`
    FullName   string `json:"full_name"`
    StudyGroup string `json:"study_group"`
    Email      string `json:"email"`
}

// Логгер с временными метками
var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

func main() {
    logger.Println("========================================")
    logger.Println("Starting HTTPS Server")
    logger.Println("========================================")
    
    // Конфигурация
    serverIP := "72.56.69.145"
    serverPort := "8443"
    serverAddr := serverIP + ":" + serverPort
    
    // Подключение к БД
    dsn := "postgres://elcakog:5179@localhost:5432/study_security?sslmode=disable"
    logger.Printf("Connecting to database with DSN: postgres://elcakog:****@localhost:5432/study_security")
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        logger.Fatalf("❌ Failed to open database: %v", err)
    }
    defer db.Close()
    
    // Проверка подключения к БД
    logger.Print("Testing database connection...")
    if err := db.Ping(); err != nil {
        logger.Fatalf("❌ Failed to ping database: %v", err)
    }
    logger.Println("✅ Database connected successfully")
    
    // Проверяем наличие таблицы
    var tableExists bool
    err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'students')").Scan(&tableExists)
    if err != nil {
        logger.Printf("⚠️ Warning: Could not check if table exists: %v", err)
    } else if tableExists {
        logger.Println("✅ Table 'students' exists")
        
        // Считаем количество записей
        var count int
        db.QueryRow("SELECT COUNT(*) FROM students").Scan(&count)
        logger.Printf("📊 Found %d students in database", count)
    } else {
        logger.Println("⚠️ Warning: Table 'students' does not exist!")
    }
    
    // Обработчики
    mux := http.NewServeMux()
    
    // Middleware для логирования
    loggingMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            logger.Printf("📥 %s %s - Request from %s", r.Method, r.URL.Path, r.RemoteAddr)
            next(w, r)
            logger.Printf("📤 %s %s - Completed in %v", r.Method, r.URL.Path, time.Since(start))
        }
    }
    
    // Health check
    mux.HandleFunc("/health", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
        logger.Println("🏥 Health check requested")
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "ok",
            "scheme": "https",
            "server": serverIP,
            "time":   time.Now().Format(time.RFC3339),
        })
    }))
    
    // Получение студента по ID
    mux.HandleFunc("/students", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
        rawID := r.URL.Query().Get("id")
        logger.Printf("🔍 Looking for student with ID: %s", rawID)
        
        if rawID == "" {
            logger.Println("⚠️ Missing id parameter")
            http.Error(w, "id required", http.StatusBadRequest)
            return
        }
        
        id, err := strconv.ParseInt(rawID, 10, 64)
        if err != nil {
            logger.Printf("❌ Invalid ID format: %s", rawID)
            http.Error(w, "invalid id", http.StatusBadRequest)
            return
        }
        
        var st Student
        err = db.QueryRow(
            "SELECT id, full_name, study_group, email FROM students WHERE id = $1",
            id,
        ).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
        
        if err == sql.ErrNoRows {
            logger.Printf("❌ Student with ID %d not found", id)
            http.Error(w, "student not found", http.StatusNotFound)
            return
        }
        if err != nil {
            logger.Printf("❌ Database error: %v", err)
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }
        
        logger.Printf("✅ Found student: %s (ID: %d)", st.FullName, st.ID)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(st)
    }))
    
    // Получение студента по email
    mux.HandleFunc("/students/by-email", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
        email := r.URL.Query().Get("email")
        logger.Printf("🔍 Looking for student with email: %s", email)
        
        if email == "" {
            logger.Println("⚠️ Missing email parameter")
            http.Error(w, "email required", http.StatusBadRequest)
            return
        }
        
        var st Student
        err := db.QueryRow(
            "SELECT id, full_name, study_group, email FROM students WHERE email = $1",
            email,
        ).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
        
        if err == sql.ErrNoRows {
            logger.Printf("❌ Student with email %s not found", email)
            http.Error(w, "student not found", http.StatusNotFound)
            return
        }
        if err != nil {
            logger.Printf("❌ Database error: %v", err)
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }
        
        logger.Printf("✅ Found student: %s (Email: %s)", st.FullName, st.Email)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(st)
    }))
    
    // Демонстрация SQL-инъекции (опасный эндпоинт)
    mux.HandleFunc("/students/unsafe", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
        rawID := r.URL.Query().Get("id")
        logger.Printf("⚠️ UNSAFE query requested with ID: %s", rawID)
        
        if rawID == "" {
            http.Error(w, "id required", http.StatusBadRequest)
            return
        }
        
        // ОПАСНО: конкатенация строки SQL с пользовательским вводом
        query := "SELECT id, full_name, study_group, email FROM students WHERE id = " + rawID
        logger.Printf("🚨 DANGEROUS SQL QUERY: %s", query)
        
        var st Student
        err := db.QueryRow(query).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
        
        if err == sql.ErrNoRows {
            http.Error(w, "student not found", http.StatusNotFound)
            return
        }
        if err != nil {
            logger.Printf("❌ Error: %v", err)
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(st)
    }))
    
    // Загрузка сертификатов
    logger.Println("Loading TLS certificates...")
    cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
    if err != nil {
        logger.Fatalf("❌ Failed to load certificates: %v", err)
    }
    logger.Println("✅ Certificates loaded successfully")
    
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion:   tls.VersionTLS12,
    }
    
    server := &http.Server{
        Addr:      "0.0.0.0:8443",  // Слушаем на всех интерфейсах
        Handler:   mux,
        TLSConfig: tlsConfig,
    }
    
    logger.Println("========================================")
    logger.Printf("🚀 HTTPS Server starting on https://%s", serverAddr)
    logger.Printf("📍 Local access: https://localhost:%s", serverPort)
    logger.Printf("🌐 Remote access: https://%s", serverAddr)
    logger.Println("========================================")
    logger.Println("Available endpoints:")
    logger.Println("  GET /health - Health check")
    logger.Println("  GET /students?id=<id> - Get student by ID (safe)")
    logger.Println("  GET /students/by-email?email=<email> - Get student by email (safe)")
    logger.Println("  GET /students/unsafe?id=<id> - DEMO: SQL injection vulnerable")
    logger.Println("========================================")
    
    log.Fatal(server.ListenAndServeTLS("", ""))
}
