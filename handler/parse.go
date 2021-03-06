package handler

import (
	"regexp"
	"strings"
	"unicode"
)

// grab grabs i from r, or returns 0 if i >= end.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// findSpace finds first space rune in r, returning end if not found.
func findSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if unicode.IsSpace(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findNonSpace finds first non space rune in r, returning end if not found.
func findNonSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if !unicode.IsSpace(r[i]) {
			return i, true
		}
	}

	return i, false
}

// isEmptyLine returns true when r is empty or composed of only whitespace.
func isEmptyLine(r []rune, i, end int) bool {
	_, ok := findNonSpace(r, i, end)
	return !ok
}

// startsWithHelp determines if r starts with "help", skipping initial
// whitespace and returning -1 if r does not start with "help".
func startsWithHelp(r []rune, i, end int) bool {
	// find start
	var found bool
	i, found = findNonSpace(r, i, end)
	if found && i+4 > end {
		return false
	}

	// check
	if strings.ToLower(string(r[i:i+4])) == "help" {
		return true
	}

	return false
}

// trimSplit splits r by whitespace into a string slice.
func trimSplit(r []rune, i, end int) []string {
	var a []string

	for i < end {
		n, found := findNonSpace(r, i, end)
		if !found || n == end {
			// empty
			return a
		}
		m, _ := findSpace(r, n, end)
		a = append(a, string(r[n:m]))
		i = m
	}

	return a
}

var identifierRE = regexp.MustCompile(`(?i)^[a-z][a-z0-9_]{0,127}$`)

// readDollarAndTag reads a dollar style $tag$ in r, starting at i, returning
// the enclosed "tag" and position, or -1 if the dollar and tag was invalid.
func readDollarAndTag(r []rune, i, end int) (string, int, bool) {
	start, found := i, false
	i++
	for ; i < end; i++ {
		if r[i] == '$' {
			found = true
			break
		}
		if i-start > 128 {
			break
		}
	}
	if !found {
		return "", i, false
	}

	// check valid identifier
	id := string(r[start+1 : i])
	if id != "" && !identifierRE.MatchString(id) {
		return "", i, false
	}

	return id, i, true
}

// readString seeks to the end of a string (depending on the state of h)
// returning the position and whether or not the string's end was found.
//
// If the string's terminator was not found, then the result will be the passed
// end.
func readString(r []rune, i, end int, h *Handler) (int, bool) {
	var prev, c rune
	for ; i < end; i++ {
		c = r[i]
		switch {
		case h.allowdollar && h.qdollar && c == '$':
			if id, pos, ok := readDollarAndTag(r, i, end); ok && h.qid == id {
				return pos, true
			}

		case h.qdbl && c == '"':
			return i, true

		case !h.qdbl && !h.qdollar && c == '\'' && prev != '\'':
			return i, true
		}
		prev = r[i]
	}

	return end, false
}

// readMultilineComment finds the end of a multiline comment (ie, '*/').
func readMultilineComment(r []rune, i, end int) (int, bool) {
	i++
	for ; i < end; i++ {
		if r[i-1] == '*' && r[i] == '/' {
			return i, true
		}
	}
	return end, false
}

// readCommand reads the command and any parameters from r.
func readCommand(r []rune, i, end int) (string, []string, int) {
	i++

	// find end (either end of r, or the next command)
	start, found := i, false
	for ; i < end-1; i++ {
		if unicode.IsSpace(r[i]) && r[i+1] == '\\' {
			found = true
			break
		}
	}

	// fix i
	if found {
		i++
	} else {
		i = end
	}

	// split values
	a := trimSplit(r, start, i)
	if len(a) == 0 {
		return "", nil, i
	}

	return a[0], a[1:], i
}
