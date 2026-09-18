package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

func buildTree() (*cli.Command, *bytes.Buffer) {
	out := &bytes.Buffer{}
	leaf := &cli.Command{
		Name:  "leaf",
		Flags: []cli.Flag{&cli.StringFlag{Name: "leaf-local", Usage: "local flag on leaf"}},
		Action: func(context.Context, *cli.Command) error {
			return nil
		},
	}
	mid := &cli.Command{
		Name: "mid",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "mid-persistent", Usage: "persistent flag on mid", Local: false},
			&cli.StringFlag{Name: "mid-local", Usage: "local flag on mid", Local: true},
			&cli.StringFlag{Name: "mid-hidden", Usage: "hidden persistent on mid", Hidden: true},
			&cli.StringFlag{Name: "shared", Usage: "shared from mid"},
		},
		Commands: []*cli.Command{leaf},
	}
	root := &cli.Command{
		Name: "root",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "root-persistent", Usage: "persistent flag on root", Local: false},
			&cli.StringFlag{Name: "root-local", Usage: "local flag on root", Local: true},
			&cli.StringFlag{Name: "shared", Usage: "shared from root"},
		},
		Commands: []*cli.Command{mid},
		Writer:   out,
	}
	root.ErrWriter = out
	return root, out
}

func runHelp(t *testing.T, args ...string) string {
	t.Helper()
	root, out := buildTree()
	if err := root.Run(context.Background(), append([]string{"root"}, args...)); err != nil {
		t.Fatalf("Run(%q): %v", args, err)
	}
	return out.String()
}

// section returns the text of the help section that starts with header,
// up to the next blank-line-separated header.
func section(help, header string) string {
	idx := strings.Index(help, header+"\n")
	if idx == -1 {
		return ""
	}
	rest := help[idx+len(header)+1:]
	if end := strings.Index(rest, "\n\n"); end != -1 {
		rest = rest[:end]
	}
	return rest
}

func TestLeafHelpShowsIntermediatePersistentFlag(t *testing.T) {
	help := runHelp(t, "mid", "leaf", "--help")
	if !strings.Contains(help, "--mid-persistent") || !strings.Contains(help, "persistent flag on mid") {
		t.Fatalf("leaf help does not list --mid-persistent:\n%s", help)
	}
	global := section(help, "GLOBAL OPTIONS:")
	if !strings.Contains(global, "--mid-persistent") {
		t.Fatalf("--mid-persistent is not in the inherited section:\n%s", help)
	}
	if !strings.Contains(global, "--root-persistent") {
		t.Fatalf("--root-persistent missing from the inherited section:\n%s", help)
	}
}

func TestLeafHelpOmitsLocalAndHiddenAncestorFlags(t *testing.T) {
	help := runHelp(t, "mid", "leaf", "--help")
	for _, bad := range []string{"--mid-local", "--root-local", "--mid-hidden"} {
		if strings.Contains(help, bad) {
			t.Fatalf("leaf help must not list %s:\n%s", bad, help)
		}
	}
}

func TestLeafHelpKeepsOwnFlagsInOptions(t *testing.T) {
	help := runHelp(t, "mid", "leaf", "--help")
	opts := section(help, "OPTIONS:")
	if !strings.Contains(opts, "--leaf-local") {
		t.Fatalf("--leaf-local missing from OPTIONS:\n%s", help)
	}
	if strings.Contains(opts, "--mid-persistent") {
		t.Fatalf("inherited flag listed under the command's own OPTIONS:\n%s", help)
	}
}

func TestLeafHelpShowsNearestShadowingDefinitionOnly(t *testing.T) {
	help := runHelp(t, "mid", "leaf", "--help")
	if !strings.Contains(help, "shared from mid") {
		t.Fatalf("nearest definition of --shared missing:\n%s", help)
	}
	if strings.Contains(help, "shared from root") {
		t.Fatalf("shadowed ancestor definition of --shared must not be listed:\n%s", help)
	}
	if strings.Count(help, "--shared") != 1 {
		t.Fatalf("--shared listed %d times, want 1:\n%s", strings.Count(help, "--shared"), help)
	}
}

func TestMidHelpShowsOnlyRootPersistentAsInherited(t *testing.T) {
	help := runHelp(t, "mid", "--help")
	global := section(help, "GLOBAL OPTIONS:")
	if !strings.Contains(global, "--root-persistent") {
		t.Fatalf("--root-persistent missing from mid's inherited section:\n%s", help)
	}
	for _, own := range []string{"--mid-persistent", "--mid-local", "--shared"} {
		if strings.Contains(global, own) {
			t.Fatalf("mid's own flag %s listed as inherited:\n%s", own, help)
		}
	}
	opts := section(help, "OPTIONS:")
	if !strings.Contains(opts, "--mid-persistent") || !strings.Contains(opts, "--mid-local") {
		t.Fatalf("mid's own flags missing from OPTIONS:\n%s", help)
	}
	if strings.Contains(help, "shared from root") {
		t.Fatalf("mid shadows --shared; root's definition must not be listed:\n%s", help)
	}
}

func TestRootHelpUnchanged(t *testing.T) {
	help := runHelp(t, "--help")
	for _, want := range []string{"--root-persistent", "--root-local", "shared from root"} {
		if !strings.Contains(help, want) {
			t.Fatalf("root help missing %q:\n%s", want, help)
		}
	}
	for _, bad := range []string{"--mid-persistent", "--mid-local", "shared from mid", "--leaf-local"} {
		if strings.Contains(help, bad) {
			t.Fatalf("root help must not list descendant flag %q:\n%s", bad, help)
		}
	}
	if strings.Count(help, "--shared") != 1 {
		t.Fatalf("--shared listed %d times at root, want 1:\n%s", strings.Count(help, "--shared"), help)
	}
}

func TestDeeperLineageShowsAllIntermediatePersistentFlags(t *testing.T) {
	out := &bytes.Buffer{}
	d := &cli.Command{Name: "d", Action: func(context.Context, *cli.Command) error { return nil }}
	c := &cli.Command{Name: "c", Flags: []cli.Flag{&cli.BoolFlag{Name: "c-flag", Usage: "from c"}}, Commands: []*cli.Command{d}}
	b := &cli.Command{Name: "b", Flags: []cli.Flag{&cli.IntFlag{Name: "b-flag", Usage: "from b"}}, Commands: []*cli.Command{c}}
	a := &cli.Command{Name: "a", Flags: []cli.Flag{&cli.StringFlag{Name: "a-flag", Usage: "from a"}}, Commands: []*cli.Command{b}, Writer: out}
	if err := a.Run(context.Background(), []string{"a", "b", "c", "d", "--help"}); err != nil {
		t.Fatal(err)
	}
	global := section(out.String(), "GLOBAL OPTIONS:")
	for _, want := range []string{"--a-flag", "--b-flag", "--c-flag"} {
		if !strings.Contains(global, want) {
			t.Fatalf("%s missing from inherited section:\n%s", want, out.String())
		}
	}
}

func TestLeafOwnFlagShadowsAncestorPersistentFlag(t *testing.T) {
	out := &bytes.Buffer{}
	leaf := &cli.Command{
		Name:   "leaf",
		Flags:  []cli.Flag{&cli.StringFlag{Name: "mode", Usage: "mode from leaf"}},
		Action: func(context.Context, *cli.Command) error { return nil },
	}
	mid := &cli.Command{Name: "mid", Flags: []cli.Flag{&cli.StringFlag{Name: "mode", Usage: "mode from mid"}}, Commands: []*cli.Command{leaf}}
	root := &cli.Command{Name: "root", Commands: []*cli.Command{mid}, Writer: out}
	if err := root.Run(context.Background(), []string{"root", "mid", "leaf", "--help"}); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	if !strings.Contains(help, "mode from leaf") {
		t.Fatalf("leaf's own --mode missing:\n%s", help)
	}
	if strings.Contains(help, "mode from mid") {
		t.Fatalf("shadowed ancestor --mode must not be listed:\n%s", help)
	}
}

func TestInheritedFlagStillParses(t *testing.T) {
	var got string
	root, _ := buildTree()
	root.Commands[0].Commands[0].Action = func(_ context.Context, c *cli.Command) error {
		got = c.String("mid-persistent")
		return nil
	}
	if err := root.Run(context.Background(), []string{"root", "mid", "leaf", "--mid-persistent", "X"}); err != nil {
		t.Fatal(err)
	}
	if got != "X" {
		t.Fatalf("mid-persistent = %q, want X", got)
	}
}
