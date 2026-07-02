package ws

import (
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
)

func TestGenerateClientID(t *testing.T) {
	t.Run("has correct length", func(t *testing.T) {
		id := generateClientID()
		// 4 bytes → 8 hex chars
		if len(id) != 8 {
			t.Errorf("generateClientID() len = %d; want 8", len(id))
		}
	})

	t.Run("is valid hex", func(t *testing.T) {
		id := generateClientID()
		if _, err := hex.DecodeString(id); err != nil {
			t.Errorf("generateClientID() %q is not valid hex: %v", id, err)
		}
	})

	t.Run("ids are unique across calls", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 20; i++ {
			id := generateClientID()
			if seen[id] {
				t.Errorf("generateClientID() returned duplicate ID %q on iteration %d", id, i)
			}
			seen[id] = true
		}
	})

	t.Run("only lowercase hex chars", func(t *testing.T) {
		re := regexp.MustCompile(`^[0-9a-f]+$`)
		for i := 0; i < 10; i++ {
			id := generateClientID()
			if !re.MatchString(id) {
				t.Errorf("generateClientID() %q contains non lowercase-hex characters", id)
			}
		}
	})
}

func TestGenerateName(t *testing.T) {
	t.Run("name has three parts separated by dashes", func(t *testing.T) {
		name := generateName()
		parts := strings.Split(name, "-")
		if len(parts) != 3 {
			t.Errorf("generateName() = %q; want Adjective-Noun-hex format (3 dash-separated parts)", name)
		}
	})

	t.Run("first part is a known adjective", func(t *testing.T) {
		name := generateName()
		parts := strings.Split(name, "-")
		adj := parts[0]
		found := false
		for _, a := range adjectives {
			if a == adj {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("generateName() adjective %q not in adjectives list", adj)
		}
	})

	t.Run("second part is a known noun", func(t *testing.T) {
		name := generateName()
		parts := strings.Split(name, "-")
		noun := parts[1]
		found := false
		for _, n := range nouns {
			if n == noun {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("generateName() noun %q not in nouns list", noun)
		}
	})

	t.Run("third part is 3-char hex suffix", func(t *testing.T) {
		re := regexp.MustCompile(`^[0-9a-f]{3}$`)
		for i := 0; i < 10; i++ {
			name := generateName()
			parts := strings.Split(name, "-")
			suffix := parts[2]
			if !re.MatchString(suffix) {
				t.Errorf("generateName() suffix %q is not 3-char hex", suffix)
			}
		}
	})

	t.Run("names are not always the same", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 50; i++ {
			seen[generateName()] = true
		}
		// With 10×10 adjective/noun combos and 256^2 hex suffixes, should get many unique names
		if len(seen) < 2 {
			t.Errorf("generateName() returned only 1 unique value across 50 calls")
		}
	})
}
