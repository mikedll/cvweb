package main

import (
	"testing"
)

func TestRegex(t *testing.T)  {

	line := "./file_storage/c3682d2c-0a69-4343-b94b-6642bf456791/data.json"
	matches := UUIDRegex.FindStringSubmatch(line)
	
	// \w{12}
	if matches == nil {
		t.Fatalf("Failed to match UUID string")
	} else {
		t.Logf("Matched: %s\n", matches[1])
	}
	
}
