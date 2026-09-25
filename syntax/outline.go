package syntax

import (
	"sort"
	"strings"
)

// Symbols returns all symbols recognized in the source.
func (e *LexicalEngine) Symbols() []Symbol {
	if len(e.source) == 0 {
		return nil
	}
	lines := lineRanges(e.source)
	folds := e.Folds()
	foldMap := make(map[int]int, len(folds))
	for _, f := range folds {
		foldMap[f[0]] = f[1]
	}

	lang := e.lang
	switch lang {
	case "go":
		return e.symbolsGo(lines, foldMap)
	case "markdown", "md":
		return e.symbolsMarkdown(lines)
	case "json":
		return e.symbolsJSON(lines, foldMap)
	case "yaml", "yml":
		return e.symbolsYAML(lines, foldMap)
	default:
		return nil
	}
}

func (e *LexicalEngine) symbolsGo(lines [][2]int, foldMap map[int]int) []Symbol {
	var symbols []Symbol
	for lineIdx, rng := range lines {
		line := string(e.source[rng[0]:rng[1]])
		trimmed := strings.TrimSpace(line)
		col := strings.Index(line, strings.TrimLeft(line, " \t"))
		if col < 0 {
			col = 0
		}

		if strings.HasPrefix(trimmed, "func ") {
			rest := strings.TrimSpace(trimmed[len("func "):])
			var symName, symKind string
			if strings.HasPrefix(rest, "(") {
				symKind = "method"
				closeParen := strings.Index(rest, ")")
				if closeParen != -1 {
					recvPart := strings.TrimSpace(rest[1:closeParen])
					fields := strings.Fields(recvPart)
					recvType := ""
					if len(fields) > 0 {
						recvType = strings.Trim(fields[len(fields)-1], "*&")
					}
					afterParen := strings.TrimSpace(rest[closeParen+1:])
					identEnd := strings.IndexAny(afterParen, "([< \t")
					name := afterParen
					if identEnd != -1 {
						name = afterParen[:identEnd]
					}
					if recvType != "" {
						symName = "(" + recvType + ")." + name
					} else {
						symName = name
					}
				}
			} else {
				symKind = "function"
				identEnd := strings.IndexAny(rest, "([< \t")
				if identEnd != -1 {
					symName = rest[:identEnd]
				} else {
					symName = rest
				}
			}
			if symName != "" {
				endLine := lineIdx
				if end, ok := foldMap[lineIdx]; ok {
					endLine = end
				}
				symbols = append(symbols, Symbol{
					Name:    symName,
					Kind:    symKind,
					Line:    lineIdx,
					Column:  col,
					EndLine: endLine,
				})
			}
		} else if strings.HasPrefix(trimmed, "type ") {
			rest := strings.TrimSpace(trimmed[len("type "):])
			fields := strings.Fields(rest)
			if len(fields) > 0 {
				name := fields[0]
				endLine := lineIdx
				if end, ok := foldMap[lineIdx]; ok {
					endLine = end
				}
				symbols = append(symbols, Symbol{
					Name:    name,
					Kind:    "type",
					Line:    lineIdx,
					Column:  col,
					EndLine: endLine,
				})
			}
		}
	}
	return symbols
}

func (e *LexicalEngine) symbolsMarkdown(lines [][2]int) []Symbol {
	var symbols []Symbol
	for lineIdx, rng := range lines {
		line := string(e.source[rng[0]:rng[1]])
		if strings.HasPrefix(line, "#") {
			count := 0
			for count < len(line) && line[count] == '#' {
				count++
			}
			if count < len(line) && line[count] == ' ' {
				name := strings.TrimSpace(line[count:])
				symbols = append(symbols, Symbol{
					Name:    name,
					Kind:    "heading",
					Line:    lineIdx,
					Column:  0,
					EndLine: lineIdx,
				})
			}
		}
	}
	for i := 0; i < len(symbols); i++ {
		if i+1 < len(symbols) {
			symbols[i].EndLine = symbols[i+1].Line - 1
		} else {
			symbols[i].EndLine = len(lines) - 1
		}
	}
	return symbols
}

func (e *LexicalEngine) symbolsJSON(lines [][2]int, foldMap map[int]int) []Symbol {
	var symbols []Symbol
	for lineIdx, rng := range lines {
		line := string(e.source[rng[0]:rng[1]])
		idx := strings.Index(line, "\"")
		if idx != -1 {
			idx2 := strings.Index(line[idx+1:], "\"")
			if idx2 != -1 {
				key := line[idx+1 : idx+1+idx2]
				after := strings.TrimSpace(line[idx+1+idx2+1:])
				if strings.HasPrefix(after, ":") {
					endLine := lineIdx
					if end, ok := foldMap[lineIdx]; ok {
						endLine = end
					}
					symbols = append(symbols, Symbol{
						Name:    key,
						Kind:    "key",
						Line:    lineIdx,
						Column:  idx,
						EndLine: endLine,
					})
				}
			}
		}
	}
	return symbols
}

func (e *LexicalEngine) symbolsYAML(lines [][2]int, foldMap map[int]int) []Symbol {
	var symbols []Symbol
	for lineIdx, rng := range lines {
		line := string(e.source[rng[0]:rng[1]])
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			col := strings.Index(line, strings.TrimLeft(line, " \t"))
			if col < 0 {
				col = 0
			}
			colon := strings.Index(trimmed, ":")
			if colon != -1 {
				key := strings.TrimSpace(trimmed[:colon])
				if !strings.ContainsAny(key, "\"'[]{}") && !strings.HasPrefix(trimmed, "- ") {
					endLine := lineIdx
					if end, ok := foldMap[lineIdx]; ok {
						endLine = end
					}
					symbols = append(symbols, Symbol{
						Name:    key,
						Kind:    "key",
						Line:    lineIdx,
						Column:  col,
						EndLine: endLine,
					})
				}
			}
		}
	}
	return symbols
}

// Breadcrumb returns the scope hierarchy at the given zero-based line and column.
func (e *LexicalEngine) Breadcrumb(line, col int) []string {
	if len(e.source) == 0 {
		return nil
	}
	lang := e.lang
	switch lang {
	case "go":
		symbols := e.Symbols()
		for _, sym := range symbols {
			if sym.Line <= line && line <= sym.EndLine {
				if sym.Kind == "method" && strings.HasPrefix(sym.Name, "(") {
					// Split "(Recv).Method" into ["Recv", "Method"]
					parts := strings.SplitN(sym.Name, ").", 2)
					if len(parts) == 2 {
						recv := strings.TrimPrefix(parts[0], "(")
						return []string{recv, parts[1]}
					}
				}
				return []string{sym.Name}
			}
		}
		return nil

	case "markdown", "md":
		lines := lineRanges(e.source)
		if line < 0 || line >= len(lines) {
			return nil
		}
		var crumbs []string
		var levels []int
		for i := 0; i <= line && i < len(lines); i++ {
			ln := string(e.source[lines[i][0]:lines[i][1]])
			if strings.HasPrefix(ln, "#") {
				lvl := 0
				for lvl < len(ln) && ln[lvl] == '#' {
					lvl++
				}
				if lvl < len(ln) && ln[lvl] == ' ' {
					name := strings.TrimSpace(ln[lvl:])
					for len(levels) > 0 && levels[len(levels)-1] >= lvl {
						levels = levels[:len(levels)-1]
						crumbs = crumbs[:len(crumbs)-1]
					}
					levels = append(levels, lvl)
					crumbs = append(crumbs, name)
				}
			}
		}
		return crumbs

	case "json", "yaml", "yml":
		symbols := e.Symbols()
		var res []string
		for _, sym := range symbols {
			if sym.Line <= line && line <= sym.EndLine {
				res = append(res, sym.Name)
			}
		}
		return res

	default:
		return nil
	}
}

// Folds returns all foldable range pairs [startLine, endLine] (0-based, inclusive).
func (e *LexicalEngine) Folds() [][2]int {
	if len(e.source) == 0 {
		return nil
	}
	lines := lineRanges(e.source)
	var folds [][2]int

	lang := e.lang
	if lang == "markdown" || lang == "md" {
		for i, rng := range lines {
			ln := string(e.source[rng[0]:rng[1]])
			if strings.HasPrefix(ln, "#") {
				lvl := 0
				for lvl < len(ln) && ln[lvl] == '#' {
					lvl++
				}
				if lvl < len(ln) && ln[lvl] == ' ' {
					end := len(lines) - 1
					for j := i + 1; j < len(lines); j++ {
						nextLn := string(e.source[lines[j][0]:lines[j][1]])
						if strings.HasPrefix(nextLn, "#") {
							nextLvl := 0
							for nextLvl < len(nextLn) && nextLn[nextLvl] == '#' {
								nextLvl++
							}
							if nextLvl < len(nextLn) && nextLn[nextLvl] == ' ' && nextLvl <= lvl {
								end = j - 1
								break
							}
						}
					}
					if end > i {
						folds = append(folds, [2]int{i, end})
					}
				}
			}
		}
		return folds
	}

	// Brace-based fold detection for Go, JSON, and YAML
	var stack []int
	inRawString := false

	for lineIdx, rng := range lines {
		b := e.source[rng[0]:rng[1]]
		inQuote := byte(0)

		for i := 0; i < len(b); i++ {
			c := b[i]

			if inRawString {
				if c == '`' {
					inRawString = false
				}
				continue
			}

			if inQuote != 0 {
				if c == '\\' {
					i++
					continue
				}
				if c == inQuote {
					inQuote = 0
				}
				continue
			}

			// Comments
			if (lang == "go" || lang == "json") && c == '/' && i+1 < len(b) && b[i+1] == '/' {
				break
			}
			if (lang == "yaml" || lang == "yml") && c == '#' {
				break
			}

			if c == '`' && lang == "go" {
				inRawString = true
				continue
			}
			if c == '"' || (lang == "yaml" && c == '\'') {
				inQuote = c
				continue
			}

			if c == '{' {
				stack = append(stack, lineIdx)
			} else if c == '}' {
				if len(stack) > 0 {
					startLine := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					if lineIdx > startLine {
						folds = append(folds, [2]int{startLine, lineIdx})
					}
				}
			}
		}
	}

	// Also support YAML indentation-based folds if no/few brace folds found
	if (lang == "yaml" || lang == "yml") && len(folds) == 0 {
		for i := 0; i < len(lines); i++ {
			ln := string(e.source[lines[i][0]:lines[i][1]])
			trimmed := strings.TrimSpace(ln)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.HasSuffix(trimmed, ":") {
				indent := len(ln) - len(strings.TrimLeft(ln, " \t"))
				last := i
				for j := i + 1; j < len(lines); j++ {
					sub := string(e.source[lines[j][0]:lines[j][1]])
					subTrimmed := strings.TrimSpace(sub)
					if subTrimmed == "" || strings.HasPrefix(subTrimmed, "#") {
						continue
					}
					subIndent := len(sub) - len(strings.TrimLeft(sub, " \t"))
					if subIndent > indent {
						last = j
					} else {
						break
					}
				}
				if last > i {
					folds = append(folds, [2]int{i, last})
				}
			}
		}
	}

	sort.Slice(folds, func(i, j int) bool {
		if folds[i][0] != folds[j][0] {
			return folds[i][0] < folds[j][0]
		}
		return folds[i][1] > folds[j][1]
	})

	return folds
}
