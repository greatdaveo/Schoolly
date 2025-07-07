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

// To get multiple students
func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error connecting to DB")
		return
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class FROM students WHERE 1=1"
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

	// To change student to a slice
	studentList := make([]models.Student, 0)
	// To loop through any possible rows if it is more than one rows
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			// http.Error(w, "❌ Error scanning Database results", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error scanning Database results")
			return
		}
		studentList = append(studentList, student)
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(studentList),
		Data:   studentList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

// To get single student
func GetOneStudentsHandler(w http.ResponseWriter, r *http.Request) {
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

	var student models.Student
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id,
	).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Email,
		&student.Class,
	)
	if err == sql.ErrNoRows {
		// http.Error(w, "❌ Student not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Student not found")
		return
	} else if err != nil {
		// http.Error(w, "❌ Database query error", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Database query error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(student)
}

// To add a student to the DB
func AddStudentHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error connecting to DB")
		return
	}

	defer db.Close()

	var newStudents []models.Student
	var rawStudent []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "❌ Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawStudent)
	if err != nil {
		// http.Error(w, "❌ Invalid Request Body", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid Request Body")
		return
	}

	// To perform Data Validation

	fields := GetFieldsName(models.Student{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, student := range rawStudent {
		for key := range student {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "❌ Unacceptable field found in request. Only use allowed fields", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newStudents)

	if err != nil {
		// http.Error(w, "❌ Invalid Request Body", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid Request Body")
		return
	}

	for _, student := range newStudents {
		err := CheckBlankFields(student)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// stmt, err := db.Prepare("INSERT INTO students (first_name, last_name, email, class) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))

	if err != nil {
		// http.Error(w, "❌ Error in preparing SQL query", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error in preparing SQL query")
		return
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		// res, err := stmt.Exec(newStudent.FirstName, newStudent.LastName, newStudent.Email, newStudent.Class)
		values := utils.GetStructValues(newStudent)
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

		newStudent.ID = int(lastID)
		// To add to the student list
		addedStudents[i] = newStudent
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(response)
}

// To edit and update multiple entries of a student data
func EditStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "❌ Invalid student ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid student ID")
		return
	}

	var updatedStudent models.Student
	err = json.NewDecoder(r.Body).Decode(&updatedStudent)
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

	var existingStudent models.Student
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id,
	).Scan(
		&existingStudent.ID,
		&existingStudent.FirstName,
		&existingStudent.LastName,
		&existingStudent.Email,
		&existingStudent.Class,
	)

	if err == sql.ErrNoRows {
		// http.Error(w, "❌ Student not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Student not found")
		return
	} else if err != nil {
		// http.Error(w, "❌ Unable to retrieve data", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to retrieve data")
		return
	}

	updatedStudent.ID = existingStudent.ID
	_, err = db.Exec(
		"UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
		updatedStudent.FirstName,
		updatedStudent.LastName,
		updatedStudent.Email,
		updatedStudent.Class,
		updatedStudent.ID,
	)

	if err != nil {
		// http.Error(w, "❌ Error updating student", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error updating student")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudent)
}

// To edit multiple students
func EditMultipleStudentsHandler(w http.ResponseWriter, r *http.Request) {
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
			// http.Error(w, "❌ Invalid student ID in input field", http.StatusBadRequest)
			utils.ErrorHandler(err, "❌ Invalid student ID in input field")
			return
		}

		id, err := strconv.Atoi(idStr)

		if err != nil {
			tx.Rollback()
			// http.Error(w, "❌ Error converting ID to int", http.StatusBadRequest)
			utils.ErrorHandler(err, "❌ Error converting ID to int")
			return
		}

		var student models.Student
		err = db.QueryRow(
			"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id,
		).Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.Email,
			&student.Class,
		)

		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				// http.Error(w, "❌ Student not found", http.StatusNotFound)
				utils.ErrorHandler(err, "❌ Student not found")

				return
			}
			// http.Error(w, "❌ Error retrieving student", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error retrieving student")
			return
		}

		// To update using reflect
		studentVal := reflect.ValueOf(&student).Elem()
		studentType := studentVal.Type()

		for k, v := range input {
			if k == "id" {
				continue // To skip updating the id field
			}

			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := studentVal.Field(i)
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
			"UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
			student.FirstName,
			student.LastName,
			student.Email,
			student.Class,
			student.ID,
		)

		if err != nil {
			fmt.Println(err)
			tx.Rollback()
			// http.Error(w, "❌ Error updating student", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error updating student")
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

// To update specific entries of a student data
func EditStudentSingleDataHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "❌ Invalid student id", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid student id")
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

	var existingStudent models.Student
	db.QueryRow(
		"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id,
	).Scan(
		&existingStudent.ID,
		&existingStudent.FirstName,
		&existingStudent.LastName,
		&existingStudent.Email,
		&existingStudent.Class,
	)

	// To apply update using reflect package
	studentVal := reflect.ValueOf(&existingStudent).Elem()
	studentType := studentVal.Type()

	for k, v := range input {
		// fmt.Println(k, v)
		for i := 0; i < studentVal.NumField(); i++ {
			// fmt.Println("k from reflect mechanism", k)
			field := studentType.Field(i)
			field.Tag.Get("json")

			if field.Tag.Get("json") == k+",omitempty" {
				if studentVal.Field((i)).CanSet() {
					studentVal.Field(i).Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec(
		"UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
		existingStudent.FirstName,
		existingStudent.LastName,
		existingStudent.Email,
		existingStudent.Class,
		existingStudent.ID,
	)

	if err != nil {
		// http.Error(w, "❌ Error updating student", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error updating student")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingStudent)
}

func DeleteOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "❌ Invalid student id", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid student id")

		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to connect to database")
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		// http.Error(w, "❌ Unable delete student", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable delete student")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// http.Error(w, "❌ Unable retrieve deleted student", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable retrieve deleted student")
		return
	}

	if rowsAffected == 0 {
		// http.Error(w, "❌ Student not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Student not found")
		return
	}

	// w.WriteHeader(http.StatusNoContent)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Student successfully deleted",
		ID:     id,
	}

	json.NewEncoder(w).Encode(response)
}

func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
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

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
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
			// http.Error(w, "❌ Error deleting student", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error deleting student")
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			// http.Error(w, "❌ Error retrieving deleted result", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error retrieving deleted result")
			return
		}

		// If student was delete, add the ID to the deletedIDs slice
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
		Status:     "Student successfully deleted",
		DeletedIDs: deletedIds,
	}

	json.NewEncoder(w).Encode(response)
}
