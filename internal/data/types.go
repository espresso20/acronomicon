package data

import (
	"math/rand"
	"sort"
	"strings"
	"unicode"
)

// Entry represents a single cloud technology acronym in The Acronomicon.
type Entry struct {
	Acronym    string   // Canonical display name (e.g., "VPC", "K8s", "eBPF", "mTLS")
	Aliases    []string // Optional alternative names or spellings (e.g., ["KUBERNETES"])
	FullName   string   // What the acronym stands for
	Category   string   // Domain category (e.g., "Networking", "Security & Identity")
	Provider   string   // Cloud provider or ecosystem (e.g., "Cloud-Agnostic", "AWS", "Azure", "GCP", "CNCF")
	Meaning    string   // Crisp 1-sentence summary of what it is
	WhatItDoes string   // Detailed 2-4 sentence explanation of how it works and why engineers use it
	Example    string   // Practical real-world example or command/scenario
	Related    []string // Related acronyms to explore next
}

// FirstLetter returns the uppercase A-Z bucket for this acronym.
func (e Entry) FirstLetter() rune {
	for _, r := range e.Acronym {
		if unicode.IsLetter(r) {
			return unicode.ToUpper(r)
		}
	}
	return '#'
}

// All returns the complete sorted slice of all acronyms in the database.
func All() []Entry {
	all := make([]Entry, 0, 360)
	all = append(all, entriesAC...)
	all = append(all, entriesDF...)
	all = append(all, entriesGL...)
	all = append(all, entriesMP...)
	all = append(all, entriesQS...)
	all = append(all, entriesTZ...)
	all = append(all, entriesExtra...)
	all = append(all, entriesExtra2...)

	sort.Slice(all, func(i, j int) bool {
		return strings.ToUpper(all[i].Acronym) < strings.ToUpper(all[j].Acronym)
	})
	return all
}

// NormalizeKey strips spaces, hyphens, and slashes and uppercases for flexible lookup.
func NormalizeKey(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if r != ' ' && r != '-' && r != '_' && r != '/' && r != '.' {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

// Lookup finds an exact match (by Acronym or Alias, case-insensitive and punctuation-tolerant).
// If multiple entries share the same acronym, it returns all exact matches.
func Lookup(query string) []Entry {
	norm := NormalizeKey(query)
	if norm == "" {
		return nil
	}
	all := All()
	var exact []Entry
	for _, e := range all {
		if NormalizeKey(e.Acronym) == norm {
			exact = append(exact, e)
			continue
		}
		for _, alias := range e.Aliases {
			if NormalizeKey(alias) == norm {
				exact = append(exact, e)
				break
			}
		}
	}
	return exact
}

// Suggest finds close or partial matches when an exact lookup fails.
func Suggest(query string, limit int) []Entry {
	norm := NormalizeKey(query)
	qLower := strings.ToLower(strings.TrimSpace(query))
	if norm == "" {
		return nil
	}
	all := All()

	type scored struct {
		entry Entry
		score int
	}
	var matches []scored

	for _, e := range all {
		eNorm := NormalizeKey(e.Acronym)
		score := 0

		if strings.HasPrefix(eNorm, norm) || strings.HasPrefix(norm, eNorm) {
			score += 100
		} else if strings.Contains(eNorm, norm) {
			score += 80
		}

		for _, a := range e.Aliases {
			aNorm := NormalizeKey(a)
			if strings.Contains(aNorm, norm) || strings.Contains(norm, aNorm) {
				score += 75
			}
		}

		if strings.Contains(strings.ToLower(e.FullName), qLower) {
			score += 60
		}

		// Levenshtein distance for short acronym typos
		dist := levenshtein(eNorm, norm)
		if dist == 1 {
			score += 90
		} else if dist == 2 && len(norm) >= 3 {
			score += 50
		}

		if score > 0 {
			matches = append(matches, scored{entry: e, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return strings.ToUpper(matches[i].entry.Acronym) < strings.ToUpper(matches[j].entry.Acronym)
	})

	if len(matches) > limit {
		matches = matches[:limit]
	}
	res := make([]Entry, len(matches))
	for i, m := range matches {
		res[i] = m.entry
	}
	return res
}

// Search performs a ranked full-text search across Acronym, Aliases, FullName, Category, Provider, Meaning, and WhatItDoes.
func Search(query string) []Entry {
	qLower := strings.ToLower(strings.TrimSpace(query))
	norm := NormalizeKey(query)
	if qLower == "" {
		return nil
	}

	terms := strings.Fields(qLower)
	all := All()

	type scored struct {
		entry Entry
		score int
	}
	var results []scored

	for _, e := range all {
		score := 0
		eNorm := NormalizeKey(e.Acronym)
		acrLower := strings.ToLower(e.Acronym)
		fullLower := strings.ToLower(e.FullName)
		catLower := strings.ToLower(e.Category)
		provLower := strings.ToLower(e.Provider)
		meaningLower := strings.ToLower(e.Meaning)
		whatLower := strings.ToLower(e.WhatItDoes)
		exLower := strings.ToLower(e.Example)

		if eNorm == norm {
			score += 500
		} else if strings.HasPrefix(acrLower, qLower) {
			score += 250
		} else if strings.Contains(acrLower, qLower) {
			score += 150
		}

		for _, a := range e.Aliases {
			if NormalizeKey(a) == norm {
				score += 400
			} else if strings.Contains(strings.ToLower(a), qLower) {
				score += 120
			}
		}

		if strings.Contains(fullLower, qLower) {
			score += 100
		}
		if strings.Contains(catLower, qLower) {
			score += 80
		}
		if strings.Contains(provLower, qLower) {
			score += 80
		}
		if strings.Contains(meaningLower, qLower) {
			score += 50
		}
		if strings.Contains(whatLower, qLower) {
			score += 30
		}
		if strings.Contains(exLower, qLower) {
			score += 20
		}

		// Multi-word query support: require all terms to match somewhere if not already matched
		if score == 0 && len(terms) > 1 {
			combined := acrLower + " " + fullLower + " " + catLower + " " + provLower + " " + meaningLower + " " + whatLower
			allMatched := true
			for _, t := range terms {
				if !strings.Contains(combined, t) {
					allMatched = false
					break
				}
			}
			if allMatched {
				score = 40
			}
		}

		if score > 0 {
			results = append(results, scored{entry: e, score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return strings.ToUpper(results[i].entry.Acronym) < strings.ToUpper(results[j].entry.Acronym)
	})

	out := make([]Entry, len(results))
	for i, r := range results {
		out[i] = r.entry
	}
	return out
}

// ByLetter filters entries by their first letter (A-Z).
func ByLetter(letter string) []Entry {
	letter = strings.TrimSpace(letter)
	if letter == "" {
		return All()
	}
	target := unicode.ToUpper(rune(letter[0]))
	var out []Entry
	for _, e := range All() {
		if e.FirstLetter() == target {
			out = append(out, e)
		}
	}
	return out
}

// CategoryCount holds a category name and number of entries.
type CategoryCount struct {
	Name  string
	Count int
}

// Categories returns all categories sorted by count descending, then name.
func Categories() []CategoryCount {
	counts := make(map[string]int)
	for _, e := range All() {
		counts[e.Category]++
	}
	var out []CategoryCount
	for k, v := range counts {
		out = append(out, CategoryCount{Name: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Providers returns all providers sorted by count descending.
func Providers() []CategoryCount {
	counts := make(map[string]int)
	for _, e := range All() {
		counts[e.Provider]++
	}
	var out []CategoryCount
	for k, v := range counts {
		out = append(out, CategoryCount{Name: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// ByCategory returns all entries matching a category substring (case-insensitive).
func ByCategory(catQuery string) (string, []Entry) {
	q := strings.ToLower(strings.TrimSpace(catQuery))
	all := All()
	var matchedCat string
	// Try exact or prefix match on category name first
	for _, c := range Categories() {
		if strings.ToLower(c.Name) == q || strings.Contains(strings.ToLower(c.Name), q) {
			matchedCat = c.Name
			break
		}
	}
	if matchedCat == "" {
		return "", nil
	}
	var out []Entry
	for _, e := range all {
		if e.Category == matchedCat {
			out = append(out, e)
		}
	}
	return matchedCat, out
}

// Random returns a random Entry from the database.
func Random() Entry {
	all := All()
	return all[rand.Intn(len(all))]
}

func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	dp := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		dp[j] = j
	}
	for i := 1; i <= la; i++ {
		prev := dp[0]
		dp[0] = i
		for j := 1; j <= lb; j++ {
			temp := dp[j]
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			del := dp[j] + 1
			ins := dp[j-1] + 1
			sub := prev + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			dp[j] = m
			prev = temp
		}
	}
	return dp[lb]
}
