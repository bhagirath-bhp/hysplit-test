package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <json_payload_file>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var payload Payload
	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	controlContent, err := GenerateHysplitControlFile(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating CONTROL file: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(controlContent)
}
