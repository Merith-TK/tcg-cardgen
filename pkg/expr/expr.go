// Package expr provides template expression evaluation for card style templates.
//
// Expressions are embedded inside {{ }} delimiters anywhere in a template string.
// Supported syntax inside a {{ }} block:
//
//	{{key}}                         simple variable lookup (empty string if missing)
//	{{key|fallback}}                lookup with literal fallback
//	{{key == 'val' ? 'a' : 'b'}}   ternary with equality test
//	{{key != 'val' ? 'a' : 'b'}}   ternary with inequality test
//	{{key ? 'a' : 'b'}}            ternary with truthy test
//	{{key}}                         (in boolean context) truthy if non-empty/non-false/non-0
//
// Boolean conditions (used in layer condition fields) additionally support:
//
//	{{a}} && {{b}}      both must be truthy
//	{{a}} || {{b}}      either must be truthy
package expr

import (
	"strings"
)

// Processor evaluates template expressions against a variable map.
type Processor struct {
	vars map[string]string
}

// New creates an Processor bound to the provided variable map.
// The map is not copied; it is read directly during evaluation.
func New(vars map[string]string) *Processor {
	return &Processor{vars: vars}
}

// Substitute replaces all {{ expr }} blocks in s with their evaluated string
// values, leaving any non-{{ }} text unchanged.
func (p *Processor) Substitute(s string) string {
	var b strings.Builder
	rest := s
	for {
		open := strings.Index(rest, "{{")
		if open == -1 {
			b.WriteString(rest)
			break
		}
		b.WriteString(rest[:open])
		rest = rest[open+2:]

		close := strings.Index(rest, "}}")
		if close == -1 {
			// Unclosed block — treat the rest as literal
			b.WriteString("{{")
			b.WriteString(rest)
			break
		}
		block := strings.TrimSpace(rest[:close])
		rest = rest[close+2:]
		b.WriteString(p.evalBlock(block))
	}
	return b.String()
}

// EvalBool evaluates a condition string as a boolean.
// The condition may contain multiple {{ }} blocks joined by && / || at the
// top level (outside any {{ }} delimiters).
func (p *Processor) EvalBool(condition string) bool {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return false
	}

	// Tokenise into a list of {value, op} pairs where op is &&/||.
	// We walk the string, extract {{...}} blocks and the bare operators between them.
	type token struct {
		val string
		op  string // "&&", "||", or "" for the last token
	}
	var tokens []token
	rest := condition

	for rest != "" {
		rest = strings.TrimSpace(rest)
		if rest == "" {
			break
		}

		var boolVal bool
		var hasBool bool
		if strings.HasPrefix(rest, "{{") {
			close := strings.Index(rest, "}}")
			if close == -1 {
				// Malformed — treat remainder as truthy
				boolVal = isTruthy(strings.TrimSpace(rest[2:]))
				hasBool = true
				rest = ""
			} else {
				block := strings.TrimSpace(rest[2:close])
				// In boolean context, evaluate the block as a condition expression
				// (supports key, key==val, key!=val, key|fallback).
				boolVal = p.evalCondExpr(block)
				hasBool = true
				rest = strings.TrimSpace(rest[close+2:])
			}
		}
		_ = hasBool

		var val string
		if hasBool {
			if boolVal {
				val = "true"
			} else {
				val = "false"
			}
		} else {
			// Bare identifier or literal (no {{ }} wrapper)
			// Consume until next {{, &&, or ||
			end := len(rest)
			for _, sep := range []string{"{{", "&&", "||"} {
				if idx := strings.Index(rest, sep); idx != -1 && idx < end {
					end = idx
				}
			}
			val = p.lookupVar(strings.TrimSpace(rest[:end]))
			rest = strings.TrimSpace(rest[end:])
		}

		// Check for operator
		op := ""
		if strings.HasPrefix(rest, "&&") {
			op = "&&"
			rest = strings.TrimSpace(rest[2:])
		} else if strings.HasPrefix(rest, "||") {
			op = "||"
			rest = strings.TrimSpace(rest[2:])
		}

		tokens = append(tokens, token{val: val, op: op})
	}

	if len(tokens) == 0 {
		return false
	}

	// Evaluate left-to-right with short-circuit.
	result := isTruthy(tokens[0].val)
	for i := 0; i < len(tokens)-1; i++ {
		next := isTruthy(tokens[i+1].val)
		switch tokens[i].op {
		case "&&":
			result = result && next
		case "||":
			result = result || next
		}
	}
	return result
}

// ─── internal helpers ────────────────────────────────────────────────────────

// evalBlock evaluates the content of a single {{ }} block and returns its
// string value.
func (p *Processor) evalBlock(block string) string {
	// Ternary: condition ? true_val : false_val
	if idx := indexOutsideQuotes(block, '?'); idx != -1 {
		condPart := strings.TrimSpace(block[:idx])
		rest := strings.TrimSpace(block[idx+1:])
		colonIdx := indexOutsideQuotes(rest, ':')
		if colonIdx == -1 {
			// Malformed ternary — fall through to simple lookup
			goto simple
		}
		trueVal := strings.TrimSpace(rest[:colonIdx])
		falseVal := strings.TrimSpace(rest[colonIdx+1:])

		if p.evalCondExpr(condPart) {
			return unquote(p.resolveValue(trueVal))
		}
		return unquote(p.resolveValue(falseVal))
	}

simple:
	// Fallback: key|default
	if idx := strings.Index(block, "|"); idx != -1 {
		key := strings.TrimSpace(block[:idx])
		fallback := strings.TrimSpace(block[idx+1:])
		v := p.lookupVar(key)
		if v == "" {
			return unquote(fallback)
		}
		return v
	}

	// Simple lookup
	return p.lookupVar(block)
}

// evalCondExpr evaluates a condition expression (the part before ? in a ternary).
// Supports: key, key == 'val', key != 'val', key|fallback (truthy check).
func (p *Processor) evalCondExpr(expr string) bool {
	expr = strings.TrimSpace(expr)

	if idx := strings.Index(expr, "=="); idx != -1 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+2:])
		return p.resolveValue(left) == unquote(p.resolveValue(right))
	}

	if idx := strings.Index(expr, "!="); idx != -1 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+2:])
		return p.resolveValue(left) != unquote(p.resolveValue(right))
	}

	// Truthy check (with optional |fallback)
	v := p.evalBlock(expr)
	return isTruthy(v)
}

// resolveValue resolves a value that may be a quoted literal or a variable key.
func (p *Processor) resolveValue(s string) string {
	s = strings.TrimSpace(s)
	if (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) {
		return s[1 : len(s)-1]
	}
	// Check for fallback syntax
	if idx := strings.Index(s, "|"); idx != -1 {
		key := strings.TrimSpace(s[:idx])
		fallback := strings.TrimSpace(s[idx+1:])
		v := p.lookupVar(key)
		if v == "" {
			return unquote(fallback)
		}
		return v
	}
	return p.lookupVar(s)
}

// lookupVar returns the value of a variable key, or "" if not found.
func (p *Processor) lookupVar(key string) string {
	if p.vars == nil {
		return ""
	}
	return p.vars[key]
}

// isTruthy returns true when s represents a non-empty, non-false, non-zero value.
func isTruthy(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && s != "false" && s != "0" && s != "null"
}

// unquote removes surrounding single or double quotes from a string.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') ||
			(s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// indexOutsideQuotes finds the first occurrence of ch in s that is not inside
// single or double quotes.  Returns -1 if not found.
func indexOutsideQuotes(s string, ch byte) int {
	inSingle := false
	inDouble := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case c == ch && !inSingle && !inDouble:
			return i
		}
	}
	return -1
}
