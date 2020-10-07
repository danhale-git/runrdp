package config

import "github.com/arbovm/levenshtein"

type levenshteinSort struct {
	items   []string
	pattern string
}

// Len implements sort.Interface and returns the length of the string slice to be sorted.
func (s levenshteinSort) Len() int {
	return len(s.items)
}

// Swap implements sort.Interface and swaps one item with another.
func (s levenshteinSort) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

// Less implements sort.Interface and returns a true if the levenshtein distance of i from pattern is less than that of
// j.
func (s levenshteinSort) Less(i, j int) bool {
	return levenshtein.Distance(s.pattern, s.items[i]) < levenshtein.Distance(s.pattern, s.items[j])
}
