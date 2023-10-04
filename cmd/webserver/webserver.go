package main

// Import required packages
import (
	"encoding/json" // For JSON encoding and decoding
	"fmt"           // For formatted I/O
	"log"           // Logging package for logging errors and information
	"net/http"      // HTTP package for building HTTP servers
	"sync"          // For using mutex to handle concurrency
	"time"          // For working with time

	"github.com/oklog/ulid/v2" // External package for generating ULIDs
)

// DefaultPort is the default port for the webserver
const DefaultPort = "8080"

// Global variables to keep track of counter and mutex
var counter = 0
var mutexCounter = &sync.Mutex{}

// defaultAnswer is a struct for encapsulating the default response
type defaultAnswer struct {
	Time      string `json:"time"` // Current time
	Message   string `json:"msg"`  // Message to display
	Increment int    `json:"inc"`  // Counter increment
	UUID      string `json:"uuid"` // Unique Identifier
}

// handleDefault handles the default HTTP request and returns a JSON response
func handleDefault(w http.ResponseWriter, r *http.Request) {

	// Lock and update the counter in a thread-safe manner
	mutexCounter.Lock()
	counter++
	c := counter
	mutexCounter.Unlock()

	// Initialize a defaultAnswer struct
	da := defaultAnswer{
		Time:      time.Now().Format(time.RFC3339),
		Message:   "Hello World!",
		Increment: c,
		UUID:      ulid.Make().String(),
	}

	// Convert the struct to a JSON string
	jsonData, err := json.Marshal(da)
	if err != nil {
		log.Println("JSON Marshaling Error:", err)
		return
	}

	// Write the JSON response to the client
	fmt.Fprintf(w, "%s", jsonData)

	// Log that the request was successfully handled
	log.Printf("Request %d handled successfully\n", da.Increment)
}

// main is the entry point of the application
func main() {

	// Log that the webserver is starting
	log.Println("Starting webserver...")

	// Defer the stopping log message until the main function returns
	defer log.Println("Stopping webserver...")

	// Register the default handler function for HTTP requests
	http.HandleFunc("/", handleDefault)

	// Start listening for HTTP requests
	log.Printf("Listening on http://localhost:%s/\n", DefaultPort)
	log.Fatal(http.ListenAndServe(":"+DefaultPort, nil))
}
