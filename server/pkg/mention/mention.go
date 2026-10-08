package mention

import (
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var excludedPatterns = []*regexp.Regexp{
	regexp.MustCompile("(?is)<(script|style|textarea|pre|code)\\b[^>]*>.*?</(script|style|textarea|pre|code)\\s*>"),
	regexp.MustCompile("(?is)<a\\b[^>]*>.*?</a\\s*>"),
	regexp.MustCompile("(?s)`+[^`]*`+"),
	regexp.MustCompile("!?\\[[^\\]\\n]*\\]\\([^\\)\\n]*\\)"),
	regexp.MustCompile("!?\\[[^\\]\\n]*\\]\\[[^\\]\\n]*\\]"),
	regexp.MustCompile("(?s)<[^>]*>"),
}

func Replace(content, format string, resolve func(string) int64) (string, []int64) {
	excluded := make([]bool, len(content))
	markFences(content, excluded)
	for _, pattern := range excludedPatterns {
		for _, match := range pattern.FindAllStringIndex(content, -1) {
			mark(excluded, match[0], match[1])
		}
	}
	markMarkdownLinks(content, excluded)

	var result strings.Builder
	var users []int64
	seen := map[int64]bool{}
	last := 0
	for index := 0; index < len(content); index++ {
		if content[index] != '@' || excluded[index] || !validBoundary(content[:index]) {
			continue
		}
		end := index + 1
		for end < len(content) && content[end] != ' ' && content[end] != '\n' && content[end] != '\t' && content[end] != '@' && !excluded[end] {
			end++
		}
		if end == index+1 || end >= len(content) || content[end] != ' ' || excluded[end] {
			continue
		}
		userId := resolve(content[index+1 : end])
		if userId <= 0 {
			continue
		}
		result.WriteString(escape(content[last:index], format))
		label := content[index:end]
		url := "/user/" + strconv.FormatInt(userId, 10)
		if format == "markdown" {
			result.WriteString("[" + escapeMarkdown(label) + "](" + url + ` "bbs-mention")`)
		} else {
			result.WriteString(`<a href="` + url + `" title="bbs-mention">` + html.EscapeString(label) + `</a>`)
		}
		if !seen[userId] {
			seen[userId] = true
			users = append(users, userId)
		}
		last = end
		index = end
	}
	if len(users) == 0 {
		return content, nil
	}
	result.WriteString(escape(content[last:], format))
	return result.String(), users
}

func escape(value, format string) string {
	if format == "text" {
		return html.EscapeString(value)
	}
	return value
}

func escapeMarkdown(value string) string {
	return strings.NewReplacer(`\`, `\\`, `[`, `\[`, `]`, `\]`).Replace(value)
}

func validBoundary(before string) bool {
	if before == "" {
		return true
	}
	char, _ := utf8.DecodeLastRuneInString(before)
	return !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '.' && char != '/' && char != '@'
}

func markFences(content string, excluded []bool) {
	fence := byte(0)
	fenceLength := 0
	for start := 0; start < len(content); {
		end := strings.IndexByte(content[start:], '\n')
		if end < 0 {
			end = len(content)
		} else {
			end += start + 1
		}
		line := strings.TrimLeft(content[start:end], " ")
		indent := len(content[start:end]) - len(line)
		for strings.HasPrefix(line, ">") {
			line = strings.TrimPrefix(line, ">")
			if strings.HasPrefix(line, " ") {
				line = line[1:]
			}
			indent = len(line) - len(strings.TrimLeft(line, " "))
			line = strings.TrimLeft(line, " ")
		}
		if fence == 0 && (indent >= 4 || strings.HasPrefix(content[start:end], "\t")) {
			mark(excluded, start, end)
		}
		marker := byte(0)
		count := 0
		if indent <= 3 && len(line) > 0 && (line[0] == '`' || line[0] == '~') {
			marker = line[0]
			for count < len(line) && line[count] == marker {
				count++
			}
		}
		if fence != 0 || count >= 3 {
			mark(excluded, start, end)
			if fence == 0 {
				fence, fenceLength = marker, count
			} else if marker == fence && count >= fenceLength {
				fence = 0
			}
		}
		start = end
	}
}

func mark(excluded []bool, start, end int) {
	for index := start; index < end; index++ {
		excluded[index] = true
	}
}

func markMarkdownLinks(content string, excluded []bool) {
	for start := 0; start < len(content); start++ {
		if content[start] != '[' || excluded[start] {
			continue
		}
		end := closingBracket(content, start, '[', ']')
		if end < 0 || end+1 >= len(content) {
			continue
		}
		var destinationEnd int
		switch content[end+1] {
		case '(':
			destinationEnd = closingBracket(content, end+1, '(', ')')
		case '[':
			destinationEnd = closingBracket(content, end+1, '[', ']')
		default:
			continue
		}
		if destinationEnd >= 0 {
			mark(excluded, start, destinationEnd+1)
			start = destinationEnd
		}
	}
}

func closingBracket(content string, start int, opening, closing byte) int {
	depth := 0
	for index := start; index < len(content); index++ {
		if content[index] == '\\' {
			index++
			continue
		}
		if content[index] == opening {
			depth++
		} else if content[index] == closing {
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}
