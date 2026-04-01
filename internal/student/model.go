package student

type Student struct {
    ID         int64  `json:"id"`
    FullName   string `json:"full_name"`
    StudyGroup string `json:"study_group"`
    Email      string `json:"email"`
}
