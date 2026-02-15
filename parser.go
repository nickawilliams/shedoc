package shedoc

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"
)

// Parse parses shedoc documentation from a shell script file at the given path.
func Parse(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := ParseReader(f)
	if err != nil {
		return nil, err
	}
	doc.Path = path
	return doc, nil
}

// ParseReader parses shedoc documentation from a reader.
func ParseReader(r io.Reader) (*Document, error) {
	p := &parser{
		scanner: bufio.NewScanner(r),
		doc:     &Document{},
	}
	p.parse()
	return p.doc, nil
}

type parseState int

const (
	stateTop      parseState = iota // outside any block
	stateShedoc                     // inside a #?/ multi-line block
	stateSheblock                   // inside a #@/ block
)

// Compiled patterns for line classification.
var (
	reShebang       = regexp.MustCompile(`^#!(.+)$`)
	reShedocInline  = regexp.MustCompile(`^#\?/(\w+)\s+(.+)$`)
	reShedocOpen    = regexp.MustCompile(`^#\?/(\w+)\s*$`)
	reSheblockOpen  = regexp.MustCompile(`^#@/(\w*)\s*(.*)$`)
	reContinuation  = regexp.MustCompile(`^ # ?(.*)$`)
	reBlockClose    = regexp.MustCompile(`^ ##\s*$`)
	reFuncParen     = regexp.MustCompile(`^\s*(\w[\w-]*)\s*\(\)\s*\{?`)
	reFuncKeyword   = regexp.MustCompile(`^\s*function\s+(\w[\w-]*)`)
)

type parser struct {
	scanner       *bufio.Scanner
	doc           *Document
	line          int
	state         parseState
	shedocTag     string   // current #?/ tag being accumulated
	shedocLines   []string // accumulated lines for multi-line shedoc

	// sheblock accumulation
	block         *Block
	blockDesc     []string // description lines before first @tag
	inTags        bool     // true once we've seen the first @tag
	currentTag    string   // name of current @tag being accumulated
	currentResult any      // parsed result of current @tag
	tagContLines  []string // continuation lines for current @tag
}

func (p *parser) parse() {
	for p.scanner.Scan() {
		p.line++
		line := p.scanner.Text()

		switch p.state {
		case stateTop:
			p.handleTop(line)
		case stateShedoc:
			p.handleShedoc(line)
		case stateSheblock:
			p.handleSheblock(line)
		}
	}

	// If we're mid-block at EOF, finalize what we have.
	switch p.state {
	case stateShedoc:
		p.finalizeShedoc()
	case stateSheblock:
		p.finalizeCurrentTag()
		p.finalizeBlock()
	}
}

func (p *parser) handleTop(line string) {
	// Shebang
	if m := reShebang.FindStringSubmatch(line); m != nil {
		p.doc.Shebang = strings.TrimSpace(m[1])
		return
	}

	// Shedoc single-line: #?/tag value
	if m := reShedocInline.FindStringSubmatch(line); m != nil {
		p.setShedocMeta(m[1], strings.TrimSpace(m[2]))
		return
	}

	// Shedoc block open: #?/tag
	if m := reShedocOpen.FindStringSubmatch(line); m != nil {
		p.state = stateShedoc
		p.shedocTag = m[1]
		p.shedocLines = nil
		return
	}

	// Sheblock open: #@/visibility [name]
	if m := reSheblockOpen.FindStringSubmatch(line); m != nil {
		visibility, name := parseSheblockHeader(m[1], strings.TrimSpace(m[2]))
		p.state = stateSheblock
		p.block = &Block{
			Visibility: visibility,
			Name:       name,
			Line:       p.line,
		}
		p.blockDesc = nil
		p.inTags = false
		p.currentTag = ""
		p.currentResult = nil
		p.tagContLines = nil
		return
	}

	// Function declaration — attach to most recent block if applicable.
	if funcName := matchFuncDecl(line); funcName != "" {
		if len(p.doc.Blocks) > 0 {
			last := &p.doc.Blocks[len(p.doc.Blocks)-1]
			if last.FunctionName == "" {
				last.FunctionName = funcName
			}
		}
	}
}

func (p *parser) handleShedoc(line string) {
	if reBlockClose.MatchString(line) {
		p.finalizeShedoc()
		p.state = stateTop
		return
	}

	if m := reContinuation.FindStringSubmatch(line); m != nil {
		p.shedocLines = append(p.shedocLines, m[1])
		return
	}

	// Unexpected line inside shedoc block — finalize and reprocess.
	p.finalizeShedoc()
	p.state = stateTop
	p.handleTop(line)
}

func (p *parser) handleSheblock(line string) {
	// Block close
	if reBlockClose.MatchString(line) {
		p.finalizeCurrentTag()
		p.finalizeBlock()
		p.state = stateTop
		return
	}

	// Continuation line
	m := reContinuation.FindStringSubmatch(line)
	if m == nil {
		// Non-continuation line — finalize block and reprocess.
		p.finalizeCurrentTag()
		p.finalizeBlock()
		p.state = stateTop
		p.handleTop(line)
		return
	}

	content := m[1]

	// Check for @tag
	if tagName, tagText, ok := splitTag(content); ok {
		p.finalizeCurrentTag()
		p.inTags = true

		name, result, err := parseTag(tagName, tagText, p.line)
		if err != nil {
			p.doc.Warnings = append(p.doc.Warnings, Warning{
				Line:    p.line,
				Message: err.Error(),
			})
			return
		}
		p.currentTag = name
		p.currentResult = result
		p.tagContLines = nil
		return
	}

	// Blank continuation line (just " #")
	if content == "" {
		if p.currentTag != "" {
			p.finalizeCurrentTag()
		}
		return
	}

	// Content line
	if p.currentTag != "" {
		// Tag continuation
		p.tagContLines = append(p.tagContLines, strings.TrimSpace(content))
	} else if !p.inTags {
		// Block description
		p.blockDesc = append(p.blockDesc, content)
	}
}

func (p *parser) finalizeShedoc() {
	if p.shedocTag != "" {
		value := strings.Join(p.shedocLines, "\n")
		p.setShedocMeta(p.shedocTag, value)
	}
	p.shedocTag = ""
	p.shedocLines = nil
}

func (p *parser) finalizeCurrentTag() {
	if p.currentTag == "" || p.currentResult == nil {
		p.currentTag = ""
		p.currentResult = nil
		p.tagContLines = nil
		return
	}

	// Append continuation lines to the tag's description.
	if len(p.tagContLines) > 0 {
		cont := strings.Join(p.tagContLines, " ")
		appendTagDescription(p.currentResult, cont)
	}

	p.applyTagToBlock(p.currentTag, p.currentResult)
	p.currentTag = ""
	p.currentResult = nil
	p.tagContLines = nil
}

func (p *parser) finalizeBlock() {
	if p.block == nil {
		return
	}
	if len(p.blockDesc) > 0 {
		p.block.Description = strings.Join(p.blockDesc, "\n")
	}
	p.doc.Blocks = append(p.doc.Blocks, *p.block)
	p.block = nil
}

func (p *parser) setShedocMeta(tag, value string) {
	switch tag {
	case "name":
		p.doc.Meta.Name = value
	case "version":
		p.doc.Meta.Version = value
	case "synopsis":
		p.doc.Meta.Synopsis = value
	case "description":
		p.doc.Meta.Description = value
	case "examples":
		p.doc.Meta.Examples = value
	case "section":
		p.doc.Meta.Section = value
	case "author":
		p.doc.Meta.Author = value
	case "license":
		p.doc.Meta.License = value
	default:
		p.doc.Warnings = append(p.doc.Warnings, Warning{
			Line:    p.line,
			Message: "unknown shedoc tag: #?/" + tag,
		})
	}
}

func (p *parser) applyTagToBlock(name string, result any) {
	b := p.block
	switch name {
	case "flag":
		if v, ok := result.(*Flag); ok {
			b.Flags = append(b.Flags, *v)
		}
	case "option":
		if v, ok := result.(*Option); ok {
			b.Options = append(b.Options, *v)
		}
	case "operand":
		if v, ok := result.(*Operand); ok {
			b.Operands = append(b.Operands, *v)
		}
	case "env":
		if v, ok := result.(*Env); ok {
			b.Env = append(b.Env, *v)
		}
	case "reads":
		if v, ok := result.(*Reads); ok {
			b.Reads = append(b.Reads, *v)
		}
	case "stdin":
		if v, ok := result.(*Stdin); ok {
			b.Stdin = v
		}
	case "exit":
		if v, ok := result.(*Exit); ok {
			b.Exit = append(b.Exit, *v)
		}
	case "stdout":
		if v, ok := result.(*Stdout); ok {
			b.Stdout = v
		}
	case "stderr":
		if v, ok := result.(*Stderr); ok {
			b.Stderr = v
		}
	case "sets":
		if v, ok := result.(*Sets); ok {
			b.Sets = append(b.Sets, *v)
		}
	case "writes":
		if v, ok := result.(*Writes); ok {
			b.Writes = append(b.Writes, *v)
		}
	case "deprecated":
		if v, ok := result.(*Deprecated); ok {
			b.Deprecated = v
		}
	}
}

// parseSheblockHeader interprets the visibility and optional name from a
// sheblock opening line.
func parseSheblockHeader(vis, extra string) (Visibility, string) {
	switch vis {
	case "command":
		return VisibilityCommand, ""
	case "subcommand":
		return VisibilitySubcommand, extra
	case "public":
		return VisibilityPublic, ""
	case "private":
		return VisibilityPrivate, ""
	case "":
		return VisibilityPublic, ""
	default:
		// Treat unknown visibility as public.
		return VisibilityPublic, ""
	}
}

// matchFuncDecl returns the function name if line is a function declaration.
func matchFuncDecl(line string) string {
	if m := reFuncKeyword.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	if m := reFuncParen.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	return ""
}

// splitTag checks if a content line starts with @tagname and returns the tag
// name and remaining text.
func splitTag(content string) (name, text string, ok bool) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "@") {
		return "", "", false
	}
	rest := trimmed[1:]
	idx := strings.IndexAny(rest, " \t")
	if idx < 0 {
		return rest, "", true
	}
	return rest[:idx], strings.TrimSpace(rest[idx+1:]), true
}

// appendTagDescription appends continuation text to a parsed tag's description.
func appendTagDescription(result any, text string) {
	switch v := result.(type) {
	case *Flag:
		v.Description = joinDesc(v.Description, text)
	case *Option:
		v.Description = joinDesc(v.Description, text)
	case *Operand:
		v.Description = joinDesc(v.Description, text)
	case *Env:
		v.Description = joinDesc(v.Description, text)
	case *Reads:
		v.Description = joinDesc(v.Description, text)
	case *Stdin:
		v.Description = joinDesc(v.Description, text)
	case *Exit:
		v.Description = joinDesc(v.Description, text)
	case *Stdout:
		v.Description = joinDesc(v.Description, text)
	case *Stderr:
		v.Description = joinDesc(v.Description, text)
	case *Sets:
		v.Description = joinDesc(v.Description, text)
	case *Writes:
		v.Description = joinDesc(v.Description, text)
	case *Deprecated:
		v.Message = joinDesc(v.Message, text)
	}
}

func joinDesc(existing, addition string) string {
	if existing == "" {
		return addition
	}
	return existing + " " + addition
}
