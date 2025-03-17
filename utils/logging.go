// In utils/logging.go
package utils

import (
	"log"
)

func LogWarning(message string) {
	log.Printf("[WARNING] %s", message)
}

func LogError(message string) {
	log.Printf("[ERROR] %s", message)
}
