package data

import (
	"testing"
)

func TestDatabaseIntegrity(t *testing.T) {
	all := All()
	t.Logf("Total acronyms in Acronomicon: %d", len(all))

	if len(all) < 300 {
		t.Errorf("Expected at least 300 acronyms, got %d", len(all))
	}

	seen := make(map[string]string)
	letters := make(map[rune]int)

	for i, e := range all {
		if e.Acronym == "" {
			t.Errorf("Entry %d has empty Acronym", i)
		}
		if e.FullName == "" {
			t.Errorf("Entry %q has empty FullName", e.Acronym)
		}
		if e.Category == "" {
			t.Errorf("Entry %q has empty Category", e.Acronym)
		}
		if e.Provider == "" {
			t.Errorf("Entry %q has empty Provider", e.Acronym)
		}
		if e.Meaning == "" {
			t.Errorf("Entry %q has empty Meaning", e.Acronym)
		}
		if e.WhatItDoes == "" {
			t.Errorf("Entry %q has empty WhatItDoes", e.Acronym)
		}
		if e.Example == "" {
			t.Errorf("Entry %q has empty Example", e.Acronym)
		}
		if len(e.Related) == 0 {
			t.Errorf("Entry %q has empty Related", e.Acronym)
		}

		norm := NormalizeKey(e.Acronym)
		if prev, exists := seen[norm]; exists {
			t.Errorf("Duplicate acronym key %q (%q and %q)", norm, prev, e.Acronym)
		}
		seen[norm] = e.Acronym

		letters[e.FirstLetter()]++
	}

	for ch := 'A'; ch <= 'Z'; ch++ {
		count := letters[ch]
		if count == 0 {
			t.Errorf("Letter %c has 0 acronyms!", ch)
		}
	}
}

func TestLookupAndSearch(t *testing.T) {
	// Exact lookup
	res := Lookup("k8s")
	if len(res) == 0 || res[0].Acronym != "K8s" {
		t.Fatalf("Expected Lookup(\"k8s\") to find K8s, got %+v", res)
	}

	// Alias lookup
	resAlias := Lookup("kubernetes")
	if len(resAlias) == 0 || resAlias[0].Acronym != "K8s" {
		t.Fatalf("Expected alias lookup 'kubernetes' to find K8s, got %+v", resAlias)
	}

	// Punctuation-tolerant lookup
	resPci := Lookup("pci dss")
	if len(resPci) == 0 || resPci[0].Acronym != "PCI-DSS" {
		t.Fatalf("Expected Lookup(\"pci dss\") to find PCI-DSS, got %+v", resPci)
	}

	// Suggest (fuzzy)
	sug := Suggest("kubernet", 5)
	if len(sug) == 0 {
		t.Fatalf("Expected Suggest(\"kubernet\") to return suggestions")
	}

	// Search
	searchRes := Search("network")
	if len(searchRes) < 10 {
		t.Fatalf("Expected Search(\"network\") to return at least 10 entries, got %d", len(searchRes))
	}
}
