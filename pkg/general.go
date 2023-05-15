
package pkg

import (
	"os"
	"time"
	"log"
	"github.com/joho/godotenv"
)

var Debug = false
var Env string
var timeZone *time.Location
const TimeLayout = "Mon Jan 2, 2006 at 3:04pm MST"

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func Init() {
	if(fileExists(".env")) {
		loadErr := godotenv.Load()
		if loadErr != nil {
			log.Fatal("Error loading .env file")
		}
	}

	Debug = os.Getenv("DEBUG") == "true"
	Env = os.Getenv("APP_ENV")

	var err error
	timeZone, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatalf("Error when loading location: %s\n", err)
	}
}
