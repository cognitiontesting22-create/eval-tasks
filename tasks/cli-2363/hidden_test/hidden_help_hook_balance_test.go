package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

type hookLog struct {
	events []string
	out    *bytes.Buffer
}

func (l *hookLog) before(name string) cli.BeforeFunc {
	return func(ctx context.Context, _ *cli.Command) (context.Context, error) {
		l.events = append(l.events, name+".Before")
		return ctx, nil
	}
}

func (l *hookLog) after(name string) cli.AfterFunc {
	return func(context.Context, *cli.Command) error {
		l.events = append(l.events, name+".After")
		return nil
	}
}

func (l *hookLog) action(name string, err error) cli.ActionFunc {
	return func(context.Context, *cli.Command) error {
		l.events = append(l.events, name+".Action")
		return err
	}
}

// buildApp builds app -> sub -> nested, each with Before/After/Action hooks.
func buildApp(subActionErr error) (*cli.Command, *hookLog) {
	l := &hookLog{out: &bytes.Buffer{}}
	nested := &cli.Command{
		Name:   "nested",
		Flags:  []cli.Flag{&cli.StringFlag{Name: "n"}},
		Before: l.before("nested"),
		After:  l.after("nested"),
		Action: l.action("nested", nil),
	}
	sub := &cli.Command{
		Name:     "sub",
		Flags:    []cli.Flag{&cli.StringFlag{Name: "s"}},
		Before:   l.before("sub"),
		After:    l.after("sub"),
		Action:   l.action("sub", subActionErr),
		Commands: []*cli.Command{nested},
	}
	app := &cli.Command{
		Name:      "app",
		Before:    l.before("app"),
		After:     l.after("app"),
		Commands:  []*cli.Command{sub},
		Writer:    l.out,
		ErrWriter: l.out,
	}
	return app, l
}

func run(t *testing.T, args ...string) *hookLog {
	t.Helper()
	app, l := buildApp(nil)
	if err := app.Run(context.Background(), append([]string{"app"}, args...)); err != nil {
		t.Fatalf("Run(%q): %v", args, err)
	}
	return l
}

func count(events []string, suffix string) int {
	n := 0
	for _, e := range events {
		if strings.HasSuffix(e, suffix) {
			n++
		}
	}
	return n
}

func assertBalanced(t *testing.T, l *hookLog) {
	t.Helper()
	for _, name := range []string{"app", "sub", "nested"} {
		b, a := count(l.events, name+".Before"), count(l.events, name+".After")
		if b != a {
			t.Fatalf("%s: Before called %d times but After called %d times: %v", name, b, a, l.events)
		}
	}
}

func TestSubcommandHelpFlagDoesNotFireParentAfter(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		l := run(t, "sub", flag)
		if count(l.events, ".After") != 0 || count(l.events, ".Action") != 0 {
			t.Fatalf("%s: unexpected hooks ran: %v", flag, l.events)
		}
		assertBalanced(t, l)
		if !strings.Contains(l.out.String(), "USAGE") || !strings.Contains(l.out.String(), "sub") {
			t.Fatalf("help for sub was not printed:\n%s", l.out.String())
		}
	}
}

func TestNestedHelpFlagDoesNotFireAnyAncestorAfter(t *testing.T) {
	l := run(t, "sub", "nested", "--help")
	if count(l.events, ".After") != 0 {
		t.Fatalf("After hooks ran without Before: %v", l.events)
	}
	assertBalanced(t, l)
	if !strings.Contains(l.out.String(), "nested") {
		t.Fatalf("help for nested was not printed:\n%s", l.out.String())
	}
}

func TestHelpFlagBeforeTrailingArgs(t *testing.T) {
	l := run(t, "sub", "--help", "nested")
	if count(l.events, ".After") != 0 {
		t.Fatalf("After hooks ran without Before: %v", l.events)
	}
	assertBalanced(t, l)
}

func TestRootHelpFlagRunsNoHooks(t *testing.T) {
	l := run(t, "--help")
	if len(l.events) != 0 {
		t.Fatalf("unexpected hooks: %v", l.events)
	}
}

func TestNormalSubcommandRunKeepsHookOrder(t *testing.T) {
	l := run(t, "sub")
	want := "app.Before sub.Before sub.Action sub.After app.After"
	if got := strings.Join(l.events, " "); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalNestedRunKeepsHookOrder(t *testing.T) {
	l := run(t, "sub", "nested", "--n", "x")
	want := "app.Before sub.Before nested.Before nested.Action nested.After sub.After app.After"
	if got := strings.Join(l.events, " "); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestHelpCommandStaysBalanced(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"help", "sub"}, {"sub", "help"}, {"sub", "help", "nested"}} {
		l := run(t, args...)
		assertBalanced(t, l)
		if count(l.events, ".Action") != 0 {
			t.Fatalf("%v: command action ran while showing help: %v", args, l.events)
		}
	}
}

func TestAfterStillRunsWhenActionFails(t *testing.T) {
	app, l := buildApp(errors.New("boom"))
	err := app.Run(context.Background(), []string{"app", "sub"})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected action error, got %v", err)
	}
	want := "app.Before sub.Before sub.Action sub.After app.After"
	if got := strings.Join(l.events, " "); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAfterRunsWithoutBeforeDefined(t *testing.T) {
	var afterCalls int
	out := &bytes.Buffer{}
	app := &cli.Command{
		Name:   "app",
		Writer: out,
		After: func(context.Context, *cli.Command) error {
			afterCalls++
			return nil
		},
		Commands: []*cli.Command{{
			Name:   "sub",
			Action: func(context.Context, *cli.Command) error { return nil },
		}},
	}
	if err := app.Run(context.Background(), []string{"app", "sub"}); err != nil {
		t.Fatal(err)
	}
	if afterCalls != 1 {
		t.Fatalf("After should run once on a normal subcommand run, ran %d times", afterCalls)
	}
	afterCalls = 0
	if err := app.Run(context.Background(), []string{"app", "sub", "--help"}); err != nil {
		t.Fatal(err)
	}
	if afterCalls != 0 {
		t.Fatalf("After ran %d times while showing subcommand help", afterCalls)
	}
}
