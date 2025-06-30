package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/greatdaveo/Schoolly/internal/models"
	"github.com/greatdaveo/Schoolly/internal/models/repositories/sqlconnect"
)

func TeacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTeachers(w, r)
	case http.MethodPost:
		addTeacher(w, r)
	case http.MethodPut:
		editAndUpdateTeacher(w, r)
	case http.MethodPatch:
		editTeacherSingleData(w, r)
	case http.MethodDelete:
		deleteTeacher(w, r)
	}
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFIelds := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}

	return validFIelds[field]
}

// ::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::

func getTeachers(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	fmt.Println(idStr)

	if idStr == "" {
		query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
		var args []interface{}

		// To Filter
		query, args = addFilters(r, query, args)
		// To Sort
		query = addSorting(r, query)

		rows, err := db.Query(query, args...)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "❌ Database query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// To change teacher to a slice
		teacherList := make([]models.Teacher, 0)
		// To loop through any possible rows if it is more than one rows
		for rows.Next() {
			var teacher models.Teacher
			err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
			if err != nil {
				http.Error(w, "❌ Error scanning Database results", http.StatusInternalServerError)
				return
			}
			teacherList = append(teacherList, teacher)
		}

		response := struct {
			Status string           `json:"status"`
			Count  int              `json:"count"`
			Data   []models.Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teacherList),
			Data:   teacherList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		return
	}

	// To handle path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		http.Error(w, "❌ Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "❌ Database query error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(teacher)
}

func addSorting(r *http.Request, query string) string {
	// https: //localhost:3000/teachers/?subject=Mathematics&sortby=last_name:asc&sortby=subject:desc
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY "
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func addFilters(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

// :::::::::::::::::::::::::::::::::::::::::::::::::::::::::

func addTeacher(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "❌ Invalid Request Body", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?);")
	// fmt.Println("DATABASE STMT Err ----", err)

	if err != nil {
		http.Error(w, "❌ Error in preparing SQL query", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			http.Error(w, "❌ Error inserting data into database", http.StatusInternalServerError)
			return
		}

		// To get the id of this entry
		lastID, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "❌ Error getting last insert ID", http.StatusInternalServerError)
			return
		}

		newTeacher.ID = int(lastID)
		// To add to the teacher list
		addedTeachers[i] = newTeacher
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}

// :::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// To update all details of a teacher
func editAndUpdateTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Invalid request payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Unable to connect to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id,
	).Scan(
		&existingTeacher.ID,
		&existingTeacher.FirstName,
		&existingTeacher.LastName,
		&existingTeacher.Email,
		&existingTeacher.Class,
		&existingTeacher.Subject,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "❌ Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(w, "❌ Unable to retrieve data", http.StatusInternalServerError)
		return
	}

	updatedTeacher.ID = existingTeacher.ID
	_, err = db.Exec(
		"UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		updatedTeacher.FirstName,
		updatedTeacher.LastName,
		updatedTeacher.Email,
		updatedTeacher.Class,
		updatedTeacher.Subject,
		updatedTeacher.ID,
	)

	if err != nil {
		fmt.Println("❌ Error updating teacher: ", err)
		http.Error(w, "❌ Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)
}

// :::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// To update all details of a teacher
func editTeacherSingleData(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Invalid teacher id", http.StatusBadRequest)
		return
	}

	var input map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Invalid request payload", http.StatusBadRequest)
		return

	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	db.QueryRow(
		"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id,
	).Scan(
		&existingTeacher.ID,
		&existingTeacher.FirstName,
		&existingTeacher.LastName,
		&existingTeacher.Email,
		&existingTeacher.Class,
		&existingTeacher.Subject,
	)

	// To update the teacher data
	// for k, v := range input {
	// 	switch k {
	// 	case "first_name":
	// 		existingTeacher.FirstName = v.(string)
	// 	case "last_name":
	// 		existingTeacher.LastName = v.(string)
	// 	case "email":
	// 		existingTeacher.Email = v.(string)
	// 	case "class":
	// 		existingTeacher.Class = v.(string)
	// 	case "subject":
	// 		existingTeacher.Subject = v.(string)
	// 	}
	// }

	// To apply update using reflect package
	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range input {
		for i := 0; i < teacherVal.NumField(); i++ {
			// fmt.Println("k from reflect mechanism", k)
			field := teacherType.Field(i)
			field.Tag.Get("json")

			if field.Tag.Get("json") == k+" ,omitempty" {
				if teacherVal.Field((i)).CanSet() {
					teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec(
		"UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		existingTeacher.FirstName,
		existingTeacher.LastName,
		existingTeacher.Email,
		existingTeacher.Class,
		existingTeacher.Subject,
		existingTeacher.ID,
	)

	if err != nil {
		fmt.Println("❌ Error updating teacher: ", err)
		http.Error(w, "❌ Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTeacher)
}

func deleteTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Invalid teacher id", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "❌ Unable delete teacher", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "❌ Unable retrieve deleted teacher", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "❌ Teacher not found", http.StatusNotFound)
		return
	}

	// w.WriteHeader(http.StatusNoContent)
	
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Teacher successfully deleted",
		ID:     id,
	}

	json.NewEncoder(w).Encode(response)
}
