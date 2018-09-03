package search

import (
	"unicode/utf8"
)

// build skip table of needle for Boyer-Moore search.
func BuildSkipTable(needle string) map[rune]int {
	l := utf8.RuneCountInString(needle)
	runes := []rune(needle)

	table := make(map[rune]int)

	for i := 0; i < l-1; i++ {
		j := runes[i]
		table[j] = l - i - 1
	}

	return table
}

func SearchBySkipTable(haystack, needle string, table map[rune]int) []int {

	i := 0
	hrunes := []rune(haystack)
	nrunes := []rune(needle)
	hl := utf8.RuneCountInString(haystack)
	nl := utf8.RuneCountInString(needle)

	if hl == 0 || nl == 0 || hl < nl {
		return nil
	}

	if hl == nl && haystack == needle {
		return nil
	}

	matches := make([]int, 0)

loop:
	for i+nl <= hl {
		for j := nl - 1; j >= 0; j-- {
			if hrunes[i+j] != nrunes[j] {
				if _, ok := table[hrunes[i+j]]; !ok {
					if j == nl-1 {
						i += nl
					} else {
						i += nl - j - 1
					}
				} else {
					n := table[hrunes[i+j]] - (nl - j - 1)
					if n <= 0 {
						i++
					} else {
						i += n
					}
				}
				goto loop
			}
		}

		matches = append(matches, i)

		if _, ok := table[hrunes[i+nl-1]]; ok {
			i += table[hrunes[i+nl-1]]
		} else {
			i += nl
		}
	}

	return matches
}

// search a needle in haystack and return count of needle.
func BoyerMooreSearch(haystack, needle string) []int {
	table := BuildSkipTable(needle)
	return SearchBySkipTable(haystack, needle, table)
}
