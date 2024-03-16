package embedded

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"

	. "github.com/stevegt/goadapt"
)

// lt (lexer test) asserts that the next token's type and data are
// equal to the given values.
func lt(t *testing.T, lexer *Lexer, typ, src, payload string) {
	token := lexer.Next()
	pass := true
	msg := ""
	if token.Type != typ {
		msg += fmt.Sprintf("expected token type %q, got %q\n", typ, token.Type)
		pass = false
	}
	if token.Src != src {
		msg += fmt.Sprintf("expected token src %q, got %q\n", src, token.Src)
		pass = false
	}
	if token.Payload != payload {
		msg += fmt.Sprintf("expected token payload %q, got %q\n", payload, token.Payload)
		pass = false
	}
	if !pass {
		msg += Spf("token: %#v", token)
		t.Fatal(msg)
	}
}

// pt (parser test) asserts that the given node's type, content, and
// children count are equal to the given values.  If all assertions
// pass, pt returns the child nodes of the given node.
func pt(t *testing.T, node *ASTNode, typ, content string, childrenCount int) (children []*ASTNode) {
	pass := true
	msg := ""
	if node == nil {
		msg += "expected non-nil node\n"
		pass = false
	} else {
		if node.Type != typ {
			msg += fmt.Sprintf("expected node type %q, got %q\n", typ, node.Type)
			pass = false
		}
		if node.Content != content {
			msg += fmt.Sprintf("expected node content %q, got %q\n", content, node.Content)
			pass = false
		}
		if len(node.Children) != childrenCount {
			msg += fmt.Sprintf("expected %d children, got %d\n", childrenCount, len(node.Children))
			pass = false
		}
	}
	if !pass {
		if node != nil {
			msg += node.AsJSON()
		}
		t.Fatal(msg)
	}
	return node.Children
}

// Helper function to split a byte slice into lines.  Each line
// includes a newline if the original line had one.
func bytesToLines(buf []byte) []string {
	txt := string(buf)
	// use regexp to split on \n or \r\n
	re := regexp.MustCompile(`(?ms)(^.*?(\r\n|\n))`)
	lines := re.FindAllString(txt, -1)
	return lines
}

func TestLexerEmptyInput(t *testing.T) {
	// The lexer should return an EOF token when the input is empty.
	lexer := NewLexer("")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerNewlines(t *testing.T) {
	// The lexer should return a newline token for each empty line in the input.
	lexer := NewLexer("\n\n\n")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerWhitespace(t *testing.T) {
	// The lexer should return a two tokens for each non-empty line in
	// the input, including lines with only whitespace.  The first
	// token may be a Text token with the line's content if any, and the
	// second token should be a Newline token.
	lexer := NewLexer("  \n \n\n")
	lt(t, lexer, "Text", "  ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", " ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
}

func TestLexerText(t *testing.T) {
	// The lexer should return a text token (if not empty) and a
	// newline token for for each line in the input.
	lexer := NewLexer("foo\nbar\n\nbaz\n")
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerTextWithWhitespace(t *testing.T) {
	// The lexer should return a text and newline token for each line in the
	// input, including leading/trailing whitespace.
	lexer := NewLexer("  foo\n  bar \n  baz  \n")
	lt(t, lexer, "Text", "  foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "  bar ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "  baz  ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerTripleBacktick(t *testing.T) {
	// The lexer should return TripleBacktick and Newline tokens for each
	// line in the input that starts with three backticks, and
	// an EOF token at the end of the input.
	lexer := NewLexer("```\n```\n```\n")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerNotTripleBacktick(t *testing.T) {
	// The lexer should not return a TripleBacktick token for a line
	// that does not start with three backticks.
	lexer := NewLexer("```\n ```\n````\n")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", " ```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "````", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerFileStart(t *testing.T) {
	// The lexer should return a FileStart token for each File block start marker
	lexer := NewLexer("File: foo\nFile: bar\n")
	lt(t, lexer, "FileStart", "File: foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerNotFileStart(t *testing.T) {
	// The lexer should not return a FileStart token for a line that does not start with "File: ".
	lexer := NewLexer(" File: foo\nFile: bar\nNotFile: baz\n")
	lt(t, lexer, "Text", " File: foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "NotFile: baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerFileEnd(t *testing.T) {
	// The lexer should return a FileEnd token for each File block end marker
	lexer := NewLexer("EOF_foo\nEOF_bar\n")
	lt(t, lexer, "FileEnd", "EOF_foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileEnd", "EOF_bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

// TestLexerBacktracking tests the lexer's Checkpoint and Rollback methods to ensure they work as expected.
func TestLexerBacktracking(t *testing.T) {
	// The lexer should be able to backtrack and reprocess input from a certain point.
	// Each line in the input file should be one token
	lexer := NewLexer("foo\nbar\nbaz\nbing\nbong\n")
	cp := lexer.Checkpoint()
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lexer.Rollback(cp)
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	cp = lexer.Checkpoint()
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lexer.Rollback(cp)
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bing", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bong", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func LexerMissingNewline(t *testing.T) {
	// The lexer should handle input without a trailing newline.
	lexer := NewLexer("foo\nbar")
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerFunctional(t *testing.T) {
	// Functional test reading input from file
	fn := "input.md"
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}

	lexer := NewLexer(string(buf))
	lt(t, lexer, "Text", "test line before file", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileEnd", "EOF_foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "test line after eof", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerMixedContent(t *testing.T) {
	// The lexer should return a mix of Text, FileStart, FileEnd, and TripleBacktick tokens.
	lexer := NewLexer("foo\nFile: bar\n```\n\nbaz\n```\nEOF_bar\n")
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileEnd", "EOF_bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerBacktickLanguage(t *testing.T) {
	// The lexer should handle input with a language identifier after the opening backticks.
	lexer := NewLexer("```go\npackage main\n```\n")
	lt(t, lexer, "TripleBacktick", "```go", "go")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "package main", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

// The parser uses the backtracking lexer to process input as it
// encounters different tokens.
func TestParseEmptyInput(t *testing.T) {
	lex := NewLexer("")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal("should not have error on empty input")
	}
	rootChildren := pt(t, ast, "Root", "", 1)
	pt(t, rootChildren[0], "EOF", "", 0)
}

// TestParseShowJSON tests the parser's ability to generate a JSON representation of the AST.
func TestParseASTString(t *testing.T) {
	lex := NewLexer("foo\nFile: bar\n```\nbaz\n```\nEOF_bar\n")
	Pl("lexing done")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Pl("parsing done")
	j := ast.AsJSON()
	Pl("AsJSON done")
	Tassert(t, j != "", "expected non-empty JSON string, got %q", j)
}

// TestParseTextOnly tests the parser's behavior when the input contains only text.
func TestParseTextOnly(t *testing.T) {
	lex := NewLexer("test line 1\ntest line 2\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with Text and EOF
	// children.
	rootChildren := pt(t, ast, "Root", "", 2)
	pt(t, rootChildren[0], "Text", "test line 1\ntest line 2\n", 0)
}

// TestParseCodeBlockOnly tests the parser's behavior when the input contains only a code block.
func TestParseCodeBlockOnly(t *testing.T) {
	lex := NewLexer("```\nfoo\nbar\n```\n")
	parser := NewParser(lex)
	node := parser.parseCodeBlock("")
	// Expected behavior is to return a single CodeBlock node with
	// four Text children: "foo", "\n", "bar", and "\n".
	cbChildren := pt(t, node, "CodeBlock", "", 4)
	pt(t, cbChildren[0], "Text", "foo", 0)
	pt(t, cbChildren[1], "Text", "\n", 0)
	pt(t, cbChildren[2], "Text", "bar", 0)
	pt(t, cbChildren[3], "Text", "\n", 0)
}

// TestParseCodeBlockWithLanguage tests the parser's behavior when the input contains a code block with a language identifier.
func TestParseCodeBlockWithLanguage(t *testing.T) {
	lex := NewLexer("```go\npackage main\n```\n")
	parser := NewParser(lex)
	node := parser.parseCodeBlock("")
	// Expected behavior is to return a single CodeBlock node with the
	// Text children "package main" and "\n".
	cbChildren := pt(t, node, "CodeBlock", "", 2)
	pt(t, cbChildren[0], "Text", "package main", 0)
	pt(t, cbChildren[1], "Text", "\n", 0)
}

// TestParseFileBlock tests the parser's behavior when the input contains a single file block.
func TestParseFileBlock(t *testing.T) {
	lex := NewLexer("File: foo\n```\nbar\n```\nEOF_foo\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with a File and EOF
	// children.  The File child should have the name "foo" and a
	// single Text child with the content "bar\n".
	rootChildren := pt(t, ast, "Root", "", 2)
	fileChildren := pt(t, rootChildren[0], "File", "", 1)
	pt(t, fileChildren[0], "Text", "bar\n", 0)

	pt(t, rootChildren[1], "EOF", "", 0)
}

// TestParseFileBlockWithLanguage tests the parser's behavior when the input contains a file block with a language identifier.
func TestParseFileBlockWithLanguage(t *testing.T) {
	lex := NewLexer("File: foo\n```go\npackage main\n```\nEOF_foo\n")
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with File and EOF
	// children.  The File child should have language "go" and a single
	// Text child with content "package main\n".
	rootChildren := pt(t, ast, "Root", "", 2)
	fileChildren := pt(t, rootChildren[0], "File", "", 1)
	pt(t, fileChildren[0], "Text", "package main\n", 0)

	pt(t, rootChildren[1], "EOF", "", 0)
}

// TestParseTripleBacktickOnly tests the parser's behavior when the input contains only triple backticks.
func TestParseTripleBacktickOnly(t *testing.T) {
	lex := NewLexer("```\n```\n```\n")
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with a single
	// empty CodeBlock child followed by a Text child with the content
	// "```\n" and an EOF.
	rootChildren := pt(t, ast, "Root", "", 3)
	pt(t, rootChildren[0], "CodeBlock", "", 0)
	pt(t, rootChildren[1], "Text", "```\n", 0)
	pt(t, rootChildren[2], "EOF", "", 0)
}

// TestParseBacktracking tests the parser's ability to backtrack and reprocess input from a certain point.
func TestParseBacktracking(t *testing.T) {
	input := "File: bar\n```\nbaz\ntrailing text\n"
	lex := NewLexer(input)
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with 2 children: a
	// Text node and an EOF node.  The Text node should have the
	// content "File: bar\n```\nbaz\ntrailing text\n". The parser will
	// checkpoint the lexer, then start a File node when it sees the
	// File: line, then it will hit end of input without finding an
	// EOF_ token, at which point it will rollback the lexer and
	// re-parse the File: line and everything after it as a Text node.
	// In order for this to work, the parser must checkpoint the lexer
	// before it starts parsing any new node, and it must rollback the
	// lexer if it encounters an error while parsing a node.  The
	// parser must have some sense of node priority, trying more
	// complex nodes like File before simpler nodes like Text.
	rootChildren := pt(t, ast, "Root", "", 2)
	pt(t, rootChildren[0], "Text", input, 0)
	pt(t, rootChildren[1], "EOF", "", 0)
}

// TestParseNoEOF tests the parser's behavior when the input contains a file block without an EOF marker.
func XXXTestParseNoEOF(t *testing.T) {
	lex := NewLexer("File: foo\n```\nbar\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with a single Text child.
	children := ast.Children
	if len(children) != 1 {
		Pl(ast.AsJSON())
		t.Fatalf("expected 1 child node, got %d", len(children))
	}
	Tassert(t, children[0].Type == "Text", "expected child to be of type %q, got %q", "Text", children[0].Type)
	Tassert(t, children[0].Content == "File: foo\n```\nbar\n", "expected child content to be %q, got %q", "File: foo\n```\nbar\n", children[0].Content)
}

/*
func TestParseIncorrectEOFMarker(t *testing.T) {
	// Expected behavior is to ignore the incorrect EOF marker,
	// including it as if it were part of the file content.

	buf, err := ioutil.ReadFile("input_incorrect_eof.md")
	Ck(err)
	Pf("buf: %q\n", buf)

	lex := NewLexer(string(buf))
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	children := ast.Children
	Tassert(t, len(children) == 2, "expected %d children nodes, got %d", 2, len(children))
	Tassert(t, children[0].Type() == "Text", "expected first child to be of type %q, got %q", "Text", children[0].Type())
	Tassert(t, children[1].Type() == "Text", "expected second child to be of type %q, got %q", "Text", children[1].Type())

	// get the input file lines
	lines := bytesToLines(buf)

	// the first child should contain the first line of the input file
	Tassert(t, children[0].Content() == lines[0], "expected first child content to be %q, got %q", lines[0], children[0].Content())

	// the second child should contain the rest of the input file
	restOfInput := strings.Join(lines[1:], "")
	Tassert(t, children[1].Content() == restOfInput, "expected second child content to be %q, got %q", restOfInput, children[1].Content())

}
*/

/*
func XXXTestParseFunctional(t *testing.T) {
	// Functional test reading input from file
	fn := "input.md" // Assuming the input file is in the test directory with the name 'input.md'.
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	Pf("buf: %q\n", buf)

	lex := NewLexer(string(buf))
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// split buf into lines
	lines := bytesToLines(buf)

	// ensure the root node is a Root type
	Tassert(t, ast.Type() == "Root", "ast type expected %q, got %q", "Root", ast.Type())

	// ensure the root node has 3 children
	children := ast.Children
	Tassert(t, len(children) == 3, "root node expected %d children, got %d", 3, len(children))

	// ensure the first child is a Text type
	Tassert(t, children[0].Type() == "Text", "first child expected %q, got %q", "Text", children[0].Type())

	// ensure the second child is a File type
	Tassert(t, children[1].Type() == "File", "second child expected %q, got %q", "File", children[1].Type())

	// ensure the third child is a Text type
	Tassert(t, children[2].Type() == "Text", "third child expected %q, got %q", "Text", children[2].Type())

	// ensure the first child content matches the first 2 lines of the input file
	firstTwoLines := lines[0] + lines[1]
	firstContent := children[0].Content()
	Tassert(t, firstContent == firstTwoLines, "first child content: expected %q, got %q", firstTwoLines, firstContent)

	// ensure the second child content matches lines 4-6 of the input file
	secondContent := children[1].Content()
	expectedContent := strings.Join(lines[3:6], "")
	Pf("expectedContent: %q\n", expectedContent)
	Tassert(t, secondContent == expectedContent, "second child content: expected %q, got %q", expectedContent, secondContent)

	// ensure the third child content matches the last line of the input file
	lastLine := lines[len(lines)-1]
	thirdContent := children[2].Content()
	Tassert(t, thirdContent == lastLine, "third child content: expected %q, got %q", lastLine, thirdContent)
}

// TestParseEmbeddedFileBlocks tests the parser's ability to handle file blocks embedded within other file blocks.
func XXXTestParseEmbeddedFileBlocks(t *testing.T) {
	buf, err := ioutil.ReadFile("input_embedded_files.md")
	Ck(err)
	Pf("buf: %q\n", buf)

	lex := NewLexer(string(buf))
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to parse the input correctly, resulting in
	// a root node with 1 File child.
	children := ast.Children
	Tassert(t, len(children) == 1, "expected %d children nodes, got %d", 1, len(children))

	// the child should be of type File
	Tassert(t, children[0].Type() == "File", "expected child to be of type %q, got %q", "File", children[0].Type())

	// the child should contain the entire input file without the
	// first and last lines, which are the File and EOF markers for the
	// outer file block
	lines := bytesToLines(buf)
	expectedContent := strings.Join(lines[1:len(lines)-1], "")
	Tassert(t, children[0].Content() == expectedContent, "expected child content to be %q, got %q", expectedContent, children[0].Content())
}
*/

/*
1. **TestParseMultipleFiles**: Verify the parser correctly handles input containing multiple file blocks.
2. **TestParseNoEOF**: Test parsing input containing a file block without an EOF marker.
4. **TestParseSpecialCharactersInContent**: Check how the parser deals with special characters or escape sequences within the text or file content.
5. **TestParseWhitespaceHandling**: Verify the parser's behavior with unusual whitespace patterns, such as leading/trailing whitespaces in file names, file content, or around EOF markers.
6. **TestParseInvalidUTF8**: Determine how the parser reacts to invalid UTF-8 sequences within the input.
7. **TestParseLongFileContent**: Test the parser's ability to handle very long file contents to verify if there are any issues with buffer sizes or memory management.
8. **TestParseSingleLineFileBlock**: Ensure that the parser correctly handles a file block defined in a single line.
9. **TestParseFileNameCollisions**: Test how the parser behaves when two file blocks have the same name but different EOF markers or content.
10. **TestParseEmptyFileContent**: Verify the behavior when a file block has no content between the start and EOF markers.
11. **TestParseCommentLines**: Include tests for parsing input with lines that should be ignored, such as comments or annotations within the text.
12. **TestParseUnexpectedEOFLocation**: Test cases where the EOF marker appears in unexpected locations, such as before the file content or at the very beginning/end of the input.
13. **TestParseRobustnessAgainstMalformedInput**: Check the parser's robustness against various forms of malformed input, including incomplete file blocks, missing file names, and abrupt endings.
14. **TestParseConcurrency**: If applicable, test the parser's behavior and correctness under concurrent execution to ensure thread safety if the Parse function is anticipated to be called from multiple goroutines.
15. **TestParseErrorHandling**: Include tests that verify the parser returns meaningful error messages or codes for various error conditions, ensuring that clients can respond appropriately to different failure modes.
*/
