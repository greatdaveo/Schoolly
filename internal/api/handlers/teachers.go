package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/greatdaveo/Schoolly/internal/models"
	"github.com/greatdaveo/Schoolly/internal/models/repositories/sqlconnect"
	"github.com/greatdaveo/Schoolly/pkg/utils"
)

// To get multiple teachers
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error connecting to DB")
		return
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}

	// To Filter
	query, args = utils.AddFilters(r, query, args)
	// To Sort
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		// http.Error(w, "❌ Database query error", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Database query error")
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
			// http.Error(w, "❌ Error scanning Database results", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error scanning Database results")
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
		// http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error connecting to DB")
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
		// http.Error(w, "❌ Teacher not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Teacher not found")
		return
	} else if err != nil {
		// http.Error(w, "❌ Database query error", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Database query error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(teacher)
}

// To add a teacher to the DB
func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error connecting to DB")
		return
	}

	defer db.Close()

	var newTeachers []models.Teacher
	var rawTeacher []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "❌ Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawTeacher)
	if err != nil {
		// http.Error(w, "❌ Invalid Request Body", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid Request Body")
		return
	}

	// To perform Data Validation

	fields := GetFieldsName(models.Teacher{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, teacher := range rawTeacher {
		for key := range teacher {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "❌ Unacceptable field found in request. Only use allowed fields", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newTeachers)

	if err != nil {
		// http.Error(w, "❌ Invalid Request Body", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid Request Body")
		return
	}

	for _, teacher := range newTeachers {
		err := CheckBlankFields(teacher)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("teachers", models.Teacher{}))

	if err != nil {
		// http.Error(w, "❌ Error in preparing SQL query", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error in preparing SQL query")
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		// res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		values := utils.GetStructValues(newTeacher)
		res, err := stmt.Exec(values...)
		if err != nil {
			// http.Error(w, "❌ Error inserting data into database", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error inserting data into database")
			return
		}

		// To get the id of this entry
		lastID, err := res.LastInsertId()
		if err != nil {
			// http.Error(w, "❌ Error getting last insert ID", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error getting last insert ID")
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
		// http.Error(w, "❌ Invalid teacher ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid teacher ID")
		return
	}

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		// http.Error(w, "❌ Invalid request payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid request payload")
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Unable to connect to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to connect to DB")
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
		// http.Error(w, "❌ Teacher not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Teacher not found")
		return
	} else if err != nil {
		// http.Error(w, "❌ Unable to retrieve data", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to retrieve data")
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
		// http.Error(w, "❌ Error updating teacher", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error updating teacher")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)
}

// To edit multiple teachers
func EditMultipleTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Unable to connect to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to connect to DB")
		return
	}
	defer db.Close()

	var inputs []map[string]interface{}

	err = json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		// http.Error(w, "❌ Invalid request payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid request payload")
		return
	}

	// To set up transaction
	tx, err := db.Begin()
	if err != nil {
		// http.Error(w, "❌ Error starting transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error starting transaction")
		return
	}

	for _, input := range inputs {
		idStr, ok := input["id"].(string)
		if !ok {
			tx.Rollback()
			// http.Error(w, "❌ Invalid teacher ID in input field", http.StatusBadRequest)
			utils.ErrorHandler(err, "❌ Invalid teacher ID in input field")
			return
		}

		id, err := strconv.Atoi(idStr)
		// log.Println(id)
		// log.Printf("Type: %T", id)
		// log.Println(err)

		if err != nil {
			tx.Rollback()
			// http.Error(w, "❌ Error converting ID to int", http.StatusBadRequest)
			utils.ErrorHandler(err, "❌ Error converting ID to int")
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
				// http.Error(w, "❌ Teacher not found", http.StatusNotFound)
				utils.ErrorHandler(err, "❌ Teacher not found")

				return
			}
			// http.Error(w, "❌ Error retrieving teacher", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error retrieving teacher")
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
			// http.Error(w, "❌ Error updating teacher", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error updating teacher")
			return
		}
	}
	// To commit the transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		// http.Error(w, "❌ Error committing transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error committing transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// To update specific entries of a teacher data
func EditTeacherSingleDataHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "❌ Invalid teacher id", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid teacher id")
		return
	}

	var input map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// http.Error(w, "❌ Invalid request payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid request payload")
		return

	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to connect to database")
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
		// http.Error(w, "❌ Error updating teacher", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error updating teacher")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTeacher)
}

func DeleteOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "❌ Invalid teacher id", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid teacher id")

		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to connect to database")
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		// http.Error(w, "❌ Unable delete teacher", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable delete teacher")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// http.Error(w, "❌ Unable retrieve deleted teacher", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable retrieve deleted teacher")
		return
	}

	if rowsAffected == 0 {
		// http.Error(w, "❌ Teacher not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Teacher not found")
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

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to connect to database")
		return
	}
	defer db.Close()

	var ids []int
	json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		// http.Error(w, "❌ Invalid request payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid request payload")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		// http.Error(w, "❌ Error starting transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error starting transaction")
		return
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		tx.Rollback()
		// http.Error(w, "❌ Error preparing deleting statement", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error preparing deleting statement")
		return
	}
	defer stmt.Close()

	deletedIds := []int{}

	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			// http.Error(w, "❌ Error deleting teacher", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error deleting teacher")
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			// http.Error(w, "❌ Error retrieving deleted result", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error retrieving deleted result")
			return
		}

		// If teacher was delete, add the ID to the deletedIDs slice
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
		if rowsAffected < 1 {
			tx.Rollback()
			// http.Error(w, fmt.Sprintf("❌ ID %d does not exist", id), http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ ID does not exist")
			return
		}
	}

	// To commit transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		// http.Error(w, "❌ Error committing transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error committing transaction")
		return
	}

	if len(deletedIds) < 1 {
		// http.Error(w, "❌ IDs do not exist", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ IDs does not exist")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_ids"`
	}{
		Status:     "Teacher successfully deleted",
		DeletedIDs: deletedIds,
	}

	json.NewEncoder(w).Encode(response)
}

// To get the list of students for a specific teacher
func GetStudentsForATeacher(w http.ResponseWriter, r *http.Request) {
	teacherId := r.PathValue("id")

	var students []models.Student
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	query := `SELECT id, first_name, last_name, email, class FROM students WHERE class = (SELECT class from teachers WHERE id = ?)`
	rows, err := db.Query(query, teacherId)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	for rows.Next() {
		var student models.Student
		err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.Email,
			&student.Class,
		)

		if err != nil {
			log.Println(err)
			return
		}

		students = append(students, student)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(students),
		Data:   students,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func CountStudentsForATeacher(w http.ResponseWriter, r *http.Request) {
	teacherId := r.PathValue("id")

	var studentCount int

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// log.Println(err)
		return
	}
	defer db.Close()

	query := `SELECT COUNT(*) FROM students WHERE class = (SELECT class FROM teachers WHERE id = ?)`

	err = db.QueryRow(query, teacherId).Scan(&studentCount)
	if err != nil {
		// log.Println(err)
		return
	}

	response := struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}{
		Status: "success",
		Count:  studentCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
