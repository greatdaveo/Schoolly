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

// To get multiple teachers
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

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

}

// To get single teacher
func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	idStr := r.PathValue("id")

	fmt.Println(idStr)

	// To handle path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id,
	).Scan(
		&teacher.ID,
		&teacher.FirstName,
		&teacher.LastName,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
	)
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

// To add a teacher to the DB
func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
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

// To edit and update multiple entries of a teacher data
func EditTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
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

// To edit multiple teachers
func EditMultipleTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Unable to connect to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var inputs []map[string]interface{}

	err = json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		// log.Println(err)
		http.Error(w, "❌ Invalid request payload", http.StatusBadRequest)
		return
	}

	// To set up transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "❌ Error starting transaction", http.StatusInternalServerError)
		return
	}

	for _, input := range inputs {
		idStr, ok := input["id"].(string)
		if !ok {
			tx.Rollback()
			http.Error(w, "❌ Invalid teacher ID in input field", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		// log.Println(id)
		// log.Printf("Type: %T", id)
		// log.Println(err)

		if err != nil {
			tx.Rollback()
			http.Error(w, "❌ Error converting ID to int", http.StatusBadRequest)
			return
		}

		var teacher models.Teacher
		err = db.QueryRow(
			"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id,
		).Scan(
			&teacher.ID,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.Email,
			&teacher.Class,
			&teacher.Subject,
		)

		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				http.Error(w, "❌ Teacher not found", http.StatusNotFound)
				return
			}
			http.Error(w, "❌ Error retrieving teacher", http.StatusInternalServerError)
			return
		}

		// To update using reflect
		teacherVal := reflect.ValueOf(&teacher).Elem()
		teacherType := teacherVal.Type()

		for k, v := range input {
			if k == "id" {
				continue // To skip updating the id field
			}

			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := teacherVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert((fieldVal.Type())))
						} else {
							tx.Rollback()
							log.Printf("Cannot convert %v to %v", val.Type(), fieldVal.Type())
							return
						}
					}
					break
				}
			}
		}

		// To execute and update the values in the transaction
		_, err = tx.Exec(
			"UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
			teacher.FirstName,
			teacher.LastName,
			teacher.Email,
			teacher.Class,
			teacher.Subject,
			teacher.ID,
		)

		if err != nil {
			fmt.Println(err)
			tx.Rollback()
			http.Error(w, "❌ Error updating teacher", http.StatusInternalServerError)
			return
		}
	}
	// To commit the transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		http.Error(w, "❌ Error committing transaction ", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// To update specific entries of a teacher data
func EditTeacherSingleDataHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
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

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
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
