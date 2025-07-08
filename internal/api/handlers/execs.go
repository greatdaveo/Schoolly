package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/greatdaveo/Schoolly/internal/models"
	"github.com/greatdaveo/Schoolly/internal/models/repositories/sqlconnect"
	"github.com/greatdaveo/Schoolly/pkg/utils"
	"golang.org/x/crypto/argon2"
)

// To get multiple execs
func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error connecting to DB")
		return
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email FROM execs WHERE 1=1"
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

	// To change exec to a slice
	execList := make([]models.Exec, 0)
	// To loop through any possible rows if it is more than one rows
	for rows.Next() {
		var exec models.Exec
		err := rows.Scan(
			&exec.ID,
			&exec.FirstName,
			&exec.LastName,
			&exec.Email,
		)
		if err != nil {
			// http.Error(w, "❌ Error scanning Database results", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error scanning Database results")
			return
		}
		execList = append(execList, exec)
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(execList),
		Data:   execList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

// To get single exec
func GetOneExecHandler(w http.ResponseWriter, r *http.Request) {
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

	var exec models.Exec
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email FROM execs WHERE id = ?", id,
	).Scan(
		&exec.ID,
		&exec.FirstName,
		&exec.LastName,
		&exec.Email,
		&exec.Username,
		&exec.UserCreatedAt,
		&exec.InactiveStatus,
		&exec.Role,
	)
	if err == sql.ErrNoRows {
		// http.Error(w, "❌ Exec not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Exec not found")
		return
	} else if err != nil {
		// http.Error(w, "❌ Database query error", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Database query error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(exec)
}

// To add a exec to the DB
func AddExecsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Error connecting to DB", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error connecting to DB")
		return
	}

	defer db.Close()

	var newExecs []models.Exec
	var rawExec []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "❌ Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawExec)
	if err != nil {
		// http.Error(w, "❌ Invalid Request Body", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid Request Body")
		return
	}

	// To perform Data Validation

	fields := GetFieldsName(models.Exec{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, exec := range rawExec {
		for key := range exec {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "❌ Unacceptable field found in request. Only use allowed fields", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newExecs)

	if err != nil {
		// http.Error(w, "❌ Invalid Request Body", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid Request Body")
		return
	}

	for _, exec := range newExecs {
		err := CheckBlankFields(exec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	stmt, err := db.Prepare(utils.GenerateInsertQuery("execs", models.Exec{}))

	if err != nil {
		// http.Error(w, "❌ Error in preparing SQL query", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error in preparing SQL query")
		return
	}
	defer stmt.Close()

	addedExecss := make([]models.Exec, len(newExecs))
	for i, newExec := range newExecs {
		if newExec.Password == "" {
			// http.Error(w, "Please enter password", http.StatusBadRequest)
			utils.ErrorHandler(errors.New("❌ password is blank"), "❌ Please enter password")
			return
		}

		// To hash the password
		salt := make([]byte, 16)
		_, err := rand.Read(salt)
		if err != nil {
			utils.ErrorHandler(errors.New("❌ failed to generate salt"), "❌ Error adding data")
			return
		}
		// For Hashing
		hash := argon2.IDKey([]byte(newExec.Password), salt, 1, 64*1024, 4, 32)
		// To encode the salt
		saltBase64 := base64.StdEncoding.EncodeToString(salt)
		hashBase64 := base64.StdEncoding.EncodeToString(hash)

		encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
		// To override the password field with the hashed password
		newExec.Password = encodedHash

		values := utils.GetStructValues(newExec)
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

		newExec.ID = int(lastID)
		// To add to the exec list
		addedExecss[i] = newExec
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecss),
		Data:   addedExecss,
	}
	json.NewEncoder(w).Encode(response)
}

// To edit multiple execs
func EditMultipleExecsHandler(w http.ResponseWriter, r *http.Request) {
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
			// http.Error(w, "❌ Invalid exec ID in input field", http.StatusBadRequest)
			utils.ErrorHandler(err, "❌ Invalid exec ID in input field")
			return
		}

		id, err := strconv.Atoi(idStr)

		if err != nil {
			tx.Rollback()
			// http.Error(w, "❌ Error converting ID to int", http.StatusBadRequest)
			utils.ErrorHandler(err, "❌ Error converting ID to int")
			return
		}

		var exec models.Exec
		err = db.QueryRow(
			"SELECT id, first_name, last_name, email, username FROM execs WHERE id = ?", id,
		).Scan(
			&exec.ID,
			&exec.FirstName,
			&exec.LastName,
			&exec.Email,
			&exec.Username,
		)

		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				// http.Error(w, "❌ Exec not found", http.StatusNotFound)
				utils.ErrorHandler(err, "❌ Exec not found")

				return
			}
			// http.Error(w, "❌ Error retrieving exec", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error retrieving exec")
			return
		}

		// To update using reflect
		execVal := reflect.ValueOf(&exec).Elem()
		execType := execVal.Type()

		for k, v := range input {
			if k == "id" {
				continue // To skip updating the id field
			}

			for i := 0; i < execVal.NumField(); i++ {
				field := execType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := execVal.Field(i)
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
			"UPDATE execs SET first_name = ?, last_name = ?, email = ? WHERE id = ?",
			exec.FirstName,
			exec.LastName,
			exec.Email,
			exec.ID,
		)

		if err != nil {
			fmt.Println(err)
			tx.Rollback()
			// http.Error(w, "❌ Error updating exec", http.StatusInternalServerError)
			utils.ErrorHandler(err, "❌ Error updating exec")
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

// To update specific entries of a exec data
func EditExecSingleDataHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "❌ Invalid exec id", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid exec id")
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

	var existingExec models.Exec
	db.QueryRow(
		"SELECT id, first_name, last_name, email, username FROM execs WHERE id = ?", id,
	).Scan(
		&existingExec.ID,
		&existingExec.FirstName,
		&existingExec.LastName,
		&existingExec.Email,
		&existingExec.Username,
	)

	// To apply update using reflect package
	execVal := reflect.ValueOf(&existingExec).Elem()
	execType := execVal.Type()

	for k, v := range input {
		// fmt.Println(k, v)
		for i := 0; i < execVal.NumField(); i++ {
			// fmt.Println("k from reflect mechanism", k)
			field := execType.Field(i)
			field.Tag.Get("json")

			if field.Tag.Get("json") == k+",omitempty" {
				if execVal.Field((i)).CanSet() {
					execVal.Field(i).Set(reflect.ValueOf(v).Convert(execVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec(
		"UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ? WHERE id = ?",
		existingExec.FirstName,
		existingExec.LastName,
		existingExec.Email,
		existingExec.Username,
		existingExec.ID,
	)

	if err != nil {
		// http.Error(w, "❌ Error updating exec", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Error updating exec")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingExec)
}

func DeleteOneExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "❌ Invalid exec id", http.StatusBadRequest)
		utils.ErrorHandler(err, "❌ Invalid exec id")

		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		// http.Error(w, "❌ Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable to connect to database")
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM execs WHERE id = ?", id)
	if err != nil {
		// http.Error(w, "❌ Unable delete exec", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable delete exec")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// http.Error(w, "❌ Unable retrieve deleted exec", http.StatusInternalServerError)
		utils.ErrorHandler(err, "❌ Unable retrieve deleted exec")
		return
	}

	if rowsAffected == 0 {
		// http.Error(w, "❌ Exec not found", http.StatusNotFound)
		utils.ErrorHandler(err, "❌ Exec not found")
		return
	}

	// w.WriteHeader(http.StatusNoContent)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Exec successfully deleted",
		ID:     id,
	}

	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Exec
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "❌ Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		http.Error(w, "❌ Username and Password are required", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, "❌ Unable to connect to database")
		http.Error(w, "❌ Unable to connect to database", http.StatusNotFound)

		return
	}

	user := &models.Exec{}

	defer db.Close()

	err = db.QueryRow(
		`SELECT id, first_name, last_name, email, username, password, inactive_status, role FROM execs WHERE username = ?`,
		req.Username,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.InactiveStatus,
		&user.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorHandler(err, "❌ User not found")
			http.Error(w, "User not found", http.StatusNotFound)
		}
		http.Error(w, "database query error", http.StatusBadRequest)
		return
	}

	if user.InactiveStatus {
		http.Error(w, "Account is inactive", http.StatusForbidden)
		return
	}

	parts := strings.Split(user.Password, ".")
	if len(parts) != 2 {
		utils.ErrorHandler(errors.New("❌ invalid Encoded hash format"), "❌ Invalid Encoded hash format")
		http.Error(w, "❌ invalid Encoded hash format", http.StatusForbidden)
		return
	}

	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		utils.ErrorHandler(err, "❌ failed to decode the salt")
		http.Error(w, "❌ failed to decode the salt", http.StatusForbidden)
		return
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		utils.ErrorHandler(err, "❌ failed to decode the hashed password")
		http.Error(w, "❌ failed to decode the hashed password", http.StatusForbidden)
		return
	}

	// For Hashing
	hash := argon2.IDKey([]byte(req.Password), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedPassword) {
		utils.ErrorHandler(errors.New("❌ incorrect password"), "❌ incorrect password")
		http.Error(w, "❌ incorrect password", http.StatusForbidden)
		return
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {
		// do nothing
	} else {
		utils.ErrorHandler(errors.New("❌ incorrect password"), "❌ incorrect password")
		http.Error(w, "❌ incorrect password", http.StatusForbidden)
		return
	}

	// To generate JWT Token
	tokenString, err := utils.SignToken(user.ID, req.Username, user.Role)
	if err != nil {
		http.Error(w, "❌ Could not create login token", http.StatusInternalServerError)
		return
	}

	// To send token as response or as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(72 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "test",
		Value:    "testing",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(72 * time.Hour),
	})

	// Request Body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}

	json.NewEncoder(w).Encode(response)
}
