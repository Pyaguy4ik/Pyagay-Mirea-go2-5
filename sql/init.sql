-- Создание таблицы students
CREATE TABLE IF NOT EXISTS students (
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    study_group TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

-- Вставка тестовых данных
INSERT INTO students (full_name, study_group, email)
VALUES 
    ('Иванов Иван Иванович', 'ИВБО-01-25', 'ivanov@example.com'),
    ('Петрова Мария Сергеевна', 'ИВБО-02-25', 'petrova@example.com'),
    ('Сидоров Алексей Андреевич', 'ИВБО-03-25', 'sidorov@example.com')
ON CONFLICT (email) DO NOTHING;
