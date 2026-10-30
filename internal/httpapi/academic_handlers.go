package httpapi

import (
	"net/http"
	"strings"
)

type departmentInput struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type studentInput struct {
	UserID       int64  `json:"user_id"`
	StudentNumber string `json:"student_number"`
	DepartmentID int64  `json:"department_id"`
}

type courseInput struct {
	Code         string `json:"code"`
	Title        string `json:"title"`
	Credits      int    `json:"credits"`
	DepartmentID int64  `json:"department_id"`
	TeacherID    *int64 `json:"teacher_id"`
}

type enrollmentInput struct {
	StudentID int64 `json:"student_id"`
	CourseID  int64 `json:"course_id"`
}

type gradeInput struct {
	EnrollmentID int64   `json:"enrollment_id"`
	Score        float64 `json:"score"`
	Weight       float64 `json:"weight"`
}

func (a *API) listDepartments(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), "SELECT id,name,code,created_at FROM departments ORDER BY name")
	if err != nil { writeError(w, 500, "could not list departments"); return }
	defer rows.Close()
	result := make([]map[string]any, 0)
	for rows.Next() {
		var id int64; var name, code string; var created any
		if err := rows.Scan(&id,&name,&code,&created); err != nil { writeError(w,500,"could not read departments"); return }
		result = append(result,map[string]any{"id":id,"name":name,"code":code,"created_at":created})
	}
	writeJSON(w,200,result)
}

func (a *API) createDepartment(w http.ResponseWriter, r *http.Request) {
	var in departmentInput
	if err := decodeJSON(r,&in); err != nil { writeError(w,400,"invalid JSON body"); return }
	in.Name=strings.TrimSpace(in.Name); in.Code=strings.ToUpper(strings.TrimSpace(in.Code))
	if in.Name=="" || in.Code=="" { writeError(w,400,"name and code are required"); return }
	var id int64
	err:=a.db.QueryRow(r.Context(),"INSERT INTO departments(name,code) VALUES($1,$2) RETURNING id",in.Name,in.Code).Scan(&id)
	if err!=nil { writeError(w,409,"department code already exists or data is invalid"); return }
	writeJSON(w,201,map[string]any{"id":id,"name":in.Name,"code":in.Code})
}

func (a *API) listStudents(w http.ResponseWriter, r *http.Request) {
	rows,err:=a.db.Query(r.Context(),`SELECT s.id,s.student_number,u.full_name,u.email,d.name
		FROM students s JOIN users u ON u.id=s.user_id JOIN departments d ON d.id=s.department_id ORDER BY s.id`)
	if err!=nil { writeError(w,500,"could not list students"); return }; defer rows.Close()
	result:=make([]map[string]any,0)
	for rows.Next(){ var id int64; var number,name,email,department string
		if err:=rows.Scan(&id,&number,&name,&email,&department);err!=nil{writeError(w,500,"could not read students");return}
		result=append(result,map[string]any{"id":id,"student_number":number,"full_name":name,"email":email,"department":department})
	}
	writeJSON(w,200,result)
}

func (a *API) createStudent(w http.ResponseWriter, r *http.Request) {
	var in studentInput
	if err:=decodeJSON(r,&in);err!=nil{writeError(w,400,"invalid JSON body");return}
	if in.UserID<1 || in.DepartmentID<1 || strings.TrimSpace(in.StudentNumber)=="" {writeError(w,400,"user_id, department_id and student_number are required");return}
	var id int64
	err:=a.db.QueryRow(r.Context(),"INSERT INTO students(user_id,student_number,department_id) VALUES($1,$2,$3) RETURNING id",in.UserID,in.StudentNumber,in.DepartmentID).Scan(&id)
	if err!=nil{writeError(w,409,"student profile could not be created; check user, department and student number");return}
	writeJSON(w,201,map[string]any{"id":id,"student_number":in.StudentNumber})
}

func (a *API) listCourses(w http.ResponseWriter, r *http.Request) {
	rows,err:=a.db.Query(r.Context(),`SELECT c.id,c.code,c.title,c.credits,d.name
		FROM courses c JOIN departments d ON d.id=c.department_id ORDER BY c.code`)
	if err!=nil{writeError(w,500,"could not list courses");return};defer rows.Close()
	result:=make([]map[string]any,0)
	for rows.Next(){var id int64;var credits int;var code,title,department string
		if err:=rows.Scan(&id,&code,&title,&credits,&department);err!=nil{writeError(w,500,"could not read courses");return}
		result=append(result,map[string]any{"id":id,"code":code,"title":title,"credits":credits,"department":department})
	}
	writeJSON(w,200,result)
}

func (a *API) createCourse(w http.ResponseWriter, r *http.Request) {
	var in courseInput
	if err:=decodeJSON(r,&in);err!=nil{writeError(w,400,"invalid JSON body");return}
	if strings.TrimSpace(in.Code)=="" || strings.TrimSpace(in.Title)=="" || in.Credits<1 || in.DepartmentID<1 {writeError(w,400,"code, title, positive credits and department_id are required");return}
	var id int64
	err:=a.db.QueryRow(r.Context(),"INSERT INTO courses(code,title,credits,department_id,teacher_id) VALUES($1,$2,$3,$4,$5) RETURNING id",strings.ToUpper(in.Code),in.Title,in.Credits,in.DepartmentID,in.TeacherID).Scan(&id)
	if err!=nil{writeError(w,409,"course could not be created; check course code, department and teacher");return}
	writeJSON(w,201,map[string]any{"id":id,"code":strings.ToUpper(in.Code),"title":in.Title})
}

func (a *API) enrollStudent(w http.ResponseWriter, r *http.Request) {
	var in enrollmentInput
	if err:=decodeJSON(r,&in);err!=nil{writeError(w,400,"invalid JSON body");return}
	var id int64
	err:=a.db.QueryRow(r.Context(),"INSERT INTO enrollments(student_id,course_id) VALUES($1,$2) RETURNING id",in.StudentID,in.CourseID).Scan(&id)
	if err!=nil{writeError(w,409,"enrollment failed; check IDs, duplicate enrollment or course capacity");return}
	writeJSON(w,201,map[string]any{"id":id,"student_id":in.StudentID,"course_id":in.CourseID})
}

func (a *API) addGrade(w http.ResponseWriter, r *http.Request) {
	var in gradeInput
	if err:=decodeJSON(r,&in);err!=nil{writeError(w,400,"invalid JSON body");return}
	if in.EnrollmentID<1 || in.Score<0 || in.Score>20 || in.Weight<=0 || in.Weight>1 {writeError(w,400,"enrollment_id, score (0-20) and weight (0-1] are required");return}
	var id int64
	err:=a.db.QueryRow(r.Context(),"INSERT INTO grades(enrollment_id,score,weight) VALUES($1,$2,$3) RETURNING id",in.EnrollmentID,in.Score,in.Weight).Scan(&id)
	if err!=nil{writeError(w,409,"grade could not be added; check enrollment and weight total");return}
	writeJSON(w,201,map[string]any{"id":id,"enrollment_id":in.EnrollmentID,"score":in.Score,"weight":in.Weight})
}
