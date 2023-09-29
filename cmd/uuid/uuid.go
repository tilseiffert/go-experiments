package main

// Import required packages for generating UUIDs using different libraries
import (
	"fmt" // For formatted I/O operations

	"github.com/aidarkhanov/nanoid/v2" // For generating nanoid-based UUIDs
	"github.com/jakehl/goid"           // For generating v4 UUIDs
	"github.com/oklog/ulid/v2"         // For generating ULID-based UUIDs
	"github.com/twharmon/gouid"        // For generating custom UUIDs
)

// Showcase prints a table of UUIDs generated using different libraries.
// It demonstrates how to generate UUIDs using various configurations and libraries.
func Showcase() {
	// Helper function to print the library name, configuration, and generated UUID in a formatted manner
	print := func(lib string, name string, uuid string) {
		fmt.Printf("%-30s %-25s %s\n", lib, name, uuid)
	}

	// Print table headers
	print("lib", "config", "uuid")

	// Generate and print a UUID using gouid library with LowerCaseAlphaNum configuration
	print("github.com/twharmon/gouid", "LowerCaseAlphaNum", gouid.String(16, gouid.LowerCaseAlphaNum))

	// Generate and print a UUID using gouid library with MixedCaseAlpha configuration
	print("github.com/twharmon/gouid", "MixedCaseAlpha", gouid.String(16, gouid.MixedCaseAlpha))

	// Generate and print a ULID using oklog/ulid library
	ulid := ulid.Make()
	print("github.com/oklog/ulid/v2", "", ulid.String())

	// Generate and print a nanoid using aidarkhanov/nanoid library
	nid, _ := nanoid.New()
	print("github.com/aidarkhanov/nanoid/v2", "", nid)

	// Generate and print a v4 UUID using jakehl/goid library
	v4UUID := goid.NewV4UUID()
	print("github.com/jakehl/goid", "NewV4UUID", v4UUID.String())
}

// Entry point of the program
func main() {
	// Call the Showcase function to display the table of UUIDs
	Showcase()
}
