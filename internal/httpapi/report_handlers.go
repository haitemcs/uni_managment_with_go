package httpapi

import (
	"net/http"
	"strconv"
)

func (a *API) studentAverage(w http.ResponseWriter, r *http.Request) {
	id,err:=strconv.ParseInt(r.PathValue("id"),10,64)
	if err!=nil || id<1 {writeError(w,400,"invalid student id");return}
	var average *float64
	var studentName string
	err=a.db.QueryRow(r.Context(),`SELECT u.full_name, student_average($1)
		FROM students s JOIN users u ON u.id=s.user_id WHERE s.id=$1`,id).Scan(&studentName,&average)
	if err!=nil{writeError(w,404,"student not found");return}
	writeJSON(w,200,map[string]any{"student_id":id,"student_name":studentName,"weighted_average":average})
}

func (a *API) departmentReport(w http.ResponseWriter, r *http.Request) {
	rows,err:=a.db.Query(r.Context(),"SELECT department_id,department_name,student_count,course_count,average_grade FROM department_performance ORDER BY department_name")
	if err!=nil{writeError(w,500,"could not load department report");return};defer rows.Close()
	result:=make([]map[string]any,0)
	for rows.Next(){var id int64;var name string;var students,courses int;var average *float64
		if err:=rows.Scan(&id,&name,&students,&courses,&average);err!=nil{writeError(w,500,"could not read report");return}
		result=append(result,map[string]any{"department_id":id,"department_name":name,"student_count":students,"course_count":courses,"average_grade":average})
	}
	writeJSON(w,200,result)
}

func (a *API) courseEnrollmentReport(w http.ResponseWriter, r *http.Request) {
	rows,err:=a.db.Query(r.Context(),"SELECT course_id,course_code,course_title,enrollment_count FROM course_enrollment_summary ORDER BY course_code")
	if err!=nil{writeError(w,500,"could not load course report");return};defer rows.Close()
	result:=make([]map[string]any,0)
	for rows.Next(){var id int64;var code,title string;var count int
		if err:=rows.Scan(&id,&code,&title,&count);err!=nil{writeError(w,500,"could not read report");return}
		result=append(result,map[string]any{"course_id":id,"course_code":code,"course_title":title,"enrollment_count":count})
	}
	writeJSON(w,200,result)
}
