package utils

import (
	"errors"
	"log"
	"os"
)

func ErrorHandler(err error, message string) error {
	errorLogger := log.New(os.Stderr, "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger.Println(message, err)

	// return fmt.Errorf("%s: %w", message, err)
	return errors.New(message)
}
