package cli_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

const shellCompletionMarker = "--generate-shell-completion"

type completionFixture struct {
	root   *cli.Command
	output *bytes.Buffer
	ran    []string
}

func newCompletionFixture() *completionFixture {
	f := &completionFixture{output: &bytes.Buffer{}}
	mk := func(name string) *cli.Command {
		return &cli.Command{
			Name: name,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "color-" + name, Usage: ""},
				&cli.StringFlag{Name: "count-" + name},
				&cli.StringFlag{Name: "config-" + name + "-secret", Hidden: true},
				&cli.BoolFlag{Name: "verbose-" + name},
			},
			Action: func(context.Context, *cli.Command) error {
				f.ran = append(f.ran, name)
				return nil
			},
		}
	}
	root, sub, nested := mk("root"), mk("sub"), mk("nested")
	root.EnableShellCompletion = true
	root.Writer, root.ErrWriter = f.output, f.output
	nested.Commands = []*cli.Command{{Name: "leaf", Action: func(context.Context, *cli.Command) error { return nil }}}
	sub.Commands = []*cli.Command{nested}
	root.Commands = []*cli.Command{sub}
	f.root = root
	return f
}

func runCompletion(t *testing.T, f *completionFixture, args ...string) string {
	t.Helper()
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	full := append([]string{"app"}, args...)
	os.Args = full
	if err := f.root.Run(context.Background(), full); err != nil {
		t.Fatalf("Run(%q) returned error: %v", full, err)
	}
	if len(f.ran) != 0 {
		t.Fatalf("completion for %q executed an action: %v", full, f.ran)
	}
	return f.output.String()
}

func sortedLines(s string) []string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

func assertFlagSuggestions(t *testing.T, got string, scope string) {
	t.Helper()
	want := []string{"--color-" + scope, "--count-" + scope}
	lines := sortedLines(got)
	if len(lines) != len(want) {
		t.Fatalf("got suggestions %q, want %q", lines, want)
	}
	seen := map[string]bool{}
	for _, l := range lines {
		seen[l] = true
	}
	for _, w := range want {
		if !seen[w] {
			t.Fatalf("missing suggestion %q in %q", w, lines)
		}
	}
}

func TestPartialFlagCompletionInChildAfterOnePositional(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "value", "--co", shellCompletionMarker)
	assertFlagSuggestions(t, got, "sub")
}

func TestPartialFlagCompletionInChildAfterTwoPositionals(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "first", "second", "--co", shellCompletionMarker)
	assertFlagSuggestions(t, got, "sub")
}

func TestPartialFlagCompletionInChildWithoutPositional(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "--co", shellCompletionMarker)
	assertFlagSuggestions(t, got, "sub")
}

func TestPartialFlagCompletionInNestedAfterPositional(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "nested", "value", "--co", shellCompletionMarker)
	assertFlagSuggestions(t, got, "nested")
}

func TestPartialFlagCompletionInNestedWithoutPositional(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "nested", "--co", shellCompletionMarker)
	assertFlagSuggestions(t, got, "nested")
}

func TestPartialFlagCompletionAtRoot(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "--co", shellCompletionMarker)
	assertFlagSuggestions(t, got, "root")
}

func TestPartialFlagCompletionAtRootAfterPositional(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "value", "--co", shellCompletionMarker)
	assertFlagSuggestions(t, got, "root")
}

func TestPartialFlagCompletionMatchesMoreSpecificPrefix(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "value", "--cou", shellCompletionMarker)
	lines := sortedLines(got)
	if len(lines) != 1 || lines[0] != "--count-sub" {
		t.Fatalf("got %q, want [--count-sub]", lines)
	}
}

func TestPartialFlagCompletionHidesHiddenFlags(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "value", "--conf", shellCompletionMarker)
	if strings.Contains(got, "secret") {
		t.Fatalf("hidden flag was suggested: %q", got)
	}
}

func TestPartialFlagCompletionAfterDoubleDashProducesNothing(t *testing.T) {
	for _, path := range [][]string{nil, {"sub"}, {"sub", "nested"}} {
		f := newCompletionFixture()
		args := append(append([]string{}, path...), "--", "--co", shellCompletionMarker)
		got := runCompletion(t, f, args...)
		if got != "" {
			t.Fatalf("completion after -- for %q produced %q", path, got)
		}
	}
}

func TestCommandCompletionAfterPositionalStillListsSubcommands(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "nested", shellCompletionMarker)
	if !strings.Contains(got, "leaf") {
		t.Fatalf("expected subcommand suggestion, got %q", got)
	}
}

func TestSingleDashCompletionInChildAfterPositional(t *testing.T) {
	f := newCompletionFixture()
	got := runCompletion(t, f, "sub", "value", "-", shellCompletionMarker)
	lines := sortedLines(got)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"--color-sub", "--count-sub", "--verbose-sub"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %q", want, lines)
		}
	}
	if strings.Contains(joined, "secret") {
		t.Fatalf("hidden flag was suggested: %q", lines)
	}
}

func TestCustomShellCompleteReceivesArgs(t *testing.T) {
	f := newCompletionFixture()
	var seen []string
	f.root.Commands[0].ShellComplete = func(_ context.Context, cmd *cli.Command) {
		seen = append([]string(nil), cmd.Args().Slice()...)
	}
	_ = runCompletion(t, f, "sub", "value", "--co", shellCompletionMarker)
	if len(seen) != 2 || seen[0] != "value" || seen[1] != "--co" {
		t.Fatalf("custom completer args = %q, want [value --co]", seen)
	}
}

func TestPartialFlagWithCompletionDisabledIsAnError(t *testing.T) {
	f := newCompletionFixture()
	f.root.EnableShellCompletion = false
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	full := []string{"app", "sub", "value", "--co", shellCompletionMarker}
	os.Args = full
	if err := f.root.Run(context.Background(), full); err == nil {
		t.Fatal("expected an error for an undefined flag when completion is disabled")
	}
	if strings.Contains(f.output.String(), "--color-sub\n") {
		t.Fatalf("flags were suggested with completion disabled: %q", f.output.String())
	}
}

func TestNormalExecutionUnaffected(t *testing.T) {
	f := newCompletionFixture()
	full := []string{"app", "sub", "--count-sub", "3", "value"}
	if err := f.root.Run(context.Background(), full); err != nil {
		t.Fatal(err)
	}
	if len(f.ran) != 1 || f.ran[0] != "sub" {
		t.Fatalf("expected sub action to run once, got %v", f.ran)
	}
}
