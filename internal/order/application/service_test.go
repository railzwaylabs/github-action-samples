package application

import (
	"regexp"
	"testing"
)

func TestRandomUUID(t *testing.T) {
	const uuidV4Pattern = `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	pattern := regexp.MustCompile(uuidV4Pattern)

	first, err := randomUUID()
	if err != nil {
		t.Fatalf("generate first UUID: %v", err)
	}
	second, err := randomUUID()
	if err != nil {
		t.Fatalf("generate second UUID: %v", err)
	}
	if !pattern.MatchString(first) {
		t.Fatalf("generated ID %q is not a UUID v4", first)
	}
	if first == second {
		t.Fatal("generated duplicate UUIDs")
	}
}
