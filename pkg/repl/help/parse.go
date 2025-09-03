package replhelp

import (
	"errors"
	"strings"
)

// HelpArgs is the parsed representation of a /help command input.
type HelpArgs struct {
	Slug     string
	ShowAll  bool
	Query    string
	Types    []string
	Topics   []string
	Flags    []string
	Commands []string
	Search   string
	ListOnly bool
}

// ParseHelpInput parses a raw input line, expected to start with "/help".
func ParseHelpInput(raw string) (HelpArgs, error) {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, "/help") {
		return HelpArgs{}, errors.New("not a help command")
	}
	s = strings.TrimSpace(strings.TrimPrefix(s, "/help"))
	if s == "" {
		return HelpArgs{}, nil
	}

	tokens := splitArgsPreserveQuotes(s)
	var args HelpArgs
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if strings.HasPrefix(t, "--") {
			// flags
			switch {
			case t == "--all":
				args.ShowAll = true
			case t == "--list":
				args.ListOnly = true
			case strings.HasPrefix(t, "--query="):
				args.Query = strings.TrimPrefix(t, "--query=")
				args.Query = trimOptionalQuotes(args.Query)
			case t == "--query" && i+1 < len(tokens):
				i++
				args.Query = trimOptionalQuotes(tokens[i])
			case strings.HasPrefix(t, "--type="):
				v := strings.TrimPrefix(t, "--type=")
				v = trimOptionalQuotes(v)
				if v != "" {
					args.Types = append(args.Types, v)
				}
			case strings.HasPrefix(t, "--topic="):
				v := strings.TrimPrefix(t, "--topic=")
				v = trimOptionalQuotes(v)
				if v != "" {
					args.Topics = append(args.Topics, v)
				}
			case strings.HasPrefix(t, "--flag="):
				v := strings.TrimPrefix(t, "--flag=")
				v = trimOptionalQuotes(v)
				if v != "" {
					args.Flags = append(args.Flags, v)
				}
			case strings.HasPrefix(t, "--command="):
				v := strings.TrimPrefix(t, "--command=")
				v = trimOptionalQuotes(v)
				if v != "" {
					args.Commands = append(args.Commands, v)
				}
			case strings.HasPrefix(t, "--search="):
				v := strings.TrimPrefix(t, "--search=")
				v = trimOptionalQuotes(v)
				if v != "" {
					args.Search = v
				}
			default:
				// unknown flag: ignore to be permissive
			}
			continue
		}
		// first non-flag token is the slug if not already set
		if args.Slug == "" {
			args.Slug = t
		}
	}
	return args, nil
}

// BuildDSL returns a DSL query and true if a DSL should be used.
// If false, the caller should fallback to slug/top-level handling.
func BuildDSL(args HelpArgs) (string, bool) {
	if strings.TrimSpace(args.Query) != "" {
		return strings.TrimSpace(args.Query), true
	}

	var groups []string

	// OR within the same field
	if len(args.Types) > 0 {
		ors := make([]string, 0, len(args.Types))
		for _, v := range args.Types {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			ors = append(ors, "type:"+v)
		}
		if len(ors) > 0 {
			groups = append(groups, orJoin(ors))
		}
	}
	if len(args.Topics) > 0 {
		ors := make([]string, 0, len(args.Topics))
		for _, v := range args.Topics {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			ors = append(ors, "topic:"+v)
		}
		if len(ors) > 0 {
			groups = append(groups, orJoin(ors))
		}
	}
	if len(args.Flags) > 0 {
		ors := make([]string, 0, len(args.Flags))
		for _, v := range args.Flags {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			ors = append(ors, "flag:"+v)
		}
		if len(ors) > 0 {
			groups = append(groups, orJoin(ors))
		}
	}
	if len(args.Commands) > 0 {
		ors := make([]string, 0, len(args.Commands))
		for _, v := range args.Commands {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			ors = append(ors, "command:"+v)
		}
		if len(ors) > 0 {
			groups = append(groups, orJoin(ors))
		}
	}
	if strings.TrimSpace(args.Search) != "" {
		groups = append(groups, "\""+escapeQuotes(args.Search)+"\"")
	}

	if len(groups) == 0 {
		return "", false
	}
	return andJoin(groups), true
}

// Helpers

func splitArgsPreserveQuotes(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	quoteChar := byte(0)
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			cur.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if inQuote {
			if c == quoteChar {
				inQuote = false
				continue
			}
			cur.WriteByte(c)
			continue
		}
		if c == '"' || c == '\'' {
			inQuote = true
			quoteChar = c
			continue
		}
		if c == ' ' || c == '\t' || c == '\n' {
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(c)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func trimOptionalQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func escapeQuotes(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func orJoin(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func andJoin(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, " AND ")
}
