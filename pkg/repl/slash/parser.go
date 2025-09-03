package slash

import (
	"strings"
)

// very small tokenizer and parser sufficient for completion and basic validation

type token struct {
	text  string
	start int
	end   int // exclusive
}

func tokenizeWithSpans(s string) []token {
	var toks []token
	var cur strings.Builder
	inQuote := byte(0)
	escaped := false
	tokStart := -1
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			if tokStart < 0 {
				tokStart = i
			}
			cur.WriteByte(c)
			escaped = false
			continue
		}
		switch c {
		case '\\':
			escaped = true
			if tokStart < 0 {
				tokStart = i
			}
		case '\'', '"':
			switch inQuote {
			case 0:
				inQuote = c
				if tokStart < 0 {
					tokStart = i + 1
				}
			case c:
				inQuote = 0
				// keep quote out of token content
			default:
				if tokStart < 0 {
					tokStart = i
				}
				cur.WriteByte(c)
			}
		case ' ', '\t', '\n':
			if inQuote != 0 {
				if tokStart < 0 {
					tokStart = i
				}
				cur.WriteByte(c)
				continue
			}
			if cur.Len() > 0 {
				toks = append(toks, token{text: cur.String(), start: tokStart, end: i})
				cur.Reset()
				tokStart = -1
			}
		default:
			if tokStart < 0 {
				tokStart = i
			}
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		toks = append(toks, token{text: cur.String(), start: tokStart, end: len(s)})
	}
	return toks
}

func parseInput(raw string) (Input, []token) {
	toks := tokenizeWithSpans(strings.TrimSpace(raw))
	out := Input{Raw: raw, Flags: map[string][]string{}}
	if len(toks) == 0 {
		return out, toks
	}
	if !strings.HasPrefix(toks[0].text, "/") {
		return out, toks
	}

	// name
	out.Name = strings.TrimPrefix(toks[0].text, "/")

	// rest
	for i := 1; i < len(toks); i++ {
		t := toks[i].text
		if strings.HasPrefix(t, "--") {
			// flag or --flag=value
			eq := strings.IndexByte(t, '=')
			if eq >= 0 {
				name := t[2:eq]
				val := t[eq+1:]
				out.Flags[name] = append(out.Flags[name], val)
			} else {
				name := t[2:]
				// next token value if exists and not another flag
				if i+1 < len(toks) && !strings.HasPrefix(toks[i+1].text, "--") {
					out.Flags[name] = append(out.Flags[name], toks[i+1].text)
					i++
				} else {
					out.Flags[name] = append(out.Flags[name], "true")
				}
			}
			continue
		}
		out.Positionals = append(out.Positionals, t)
	}
	return out, toks
}

func completionState(raw string, caret int) (CompletionState, []token) {
	st := CompletionState{Raw: raw, Caret: caret}
	in, toks := parseInput(raw)
	st.Parsed = in
	if caret < 0 || caret > len(raw) {
		caret = len(raw)
	}
	// find token under caret
	tokIndex := -1
	for i, t := range toks {
		if caret >= t.start && caret <= t.end {
			tokIndex = i
			st.Partial = t.text[:max(0, caret-t.start)]
			st.TokenStart = t.start
			st.TokenEnd = t.end
			break
		}
	}
	if len(toks) == 0 || !strings.HasPrefix(raw, "/") {
		st.Phase = PhaseName
		return st, toks
	}
	if tokIndex == 0 {
		st.Phase = PhaseName
		st.Name = strings.TrimPrefix(toks[0].text, "/")
		return st, toks
	}
	// after name
	st.Name = in.Name
	// detect if we are on a flag token or value
	t := toks[tokIndex]
	if strings.HasPrefix(t.text, "--") {
		eq := strings.IndexByte(t.text, '=')
		if eq >= 0 {
			st.Phase = PhaseFlagValue
			st.CurrentFlag = t.text[2:eq]
			st.Partial = t.text[eq+1:][:max(0, caret-(t.start+eq+1))]
		} else {
			st.Phase = PhaseFlag
			st.CurrentFlag = strings.TrimPrefix(t.text, "--")
		}
		return st, toks
	}
	// positional or value after flag
	// check previous token for a flag without value consumed
	if tokIndex-1 >= 0 && strings.HasPrefix(toks[tokIndex-1].text, "--") && !strings.Contains(toks[tokIndex-1].text, "=") {
		st.Phase = PhaseFlagValue
		st.CurrentFlag = strings.TrimPrefix(toks[tokIndex-1].text, "--")
		return st, toks
	}
	st.Phase = PhasePositional
	st.PositionalIndex = len(in.Positionals) - 1
	if st.PositionalIndex < 0 {
		st.PositionalIndex = 0
	}
	return st, toks
}
