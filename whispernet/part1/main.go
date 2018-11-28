// This program reads from standard input and writes JSON-encoded messages to
// standard output. For example, this input line:
//	Hello!
// Produces this output:
//	{"Body":"Hello!"}
//
package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

// Message ...
type Message struct {
	Body string
}

func main() {
	// Create a new bufio.Scanner reading from the standard input.
	s := bufio.NewScanner(os.Stdin)

	// Create a new json.Encoder writing into the standard output.
	enc := json.NewEncoder(os.Stdout)

	// Iterate over every line in the scanner
	for s.Scan() {
		// Create a new message with the read text.
		msg := Message{s.Text()}
		// Encode the message, and check for errors!
		err := enc.Encode(msg)
		// Check for a scan error.
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}

}
