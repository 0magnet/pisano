package flags

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// tree is a command tree shaped like the real one: prose, an example, flags,
// subcommands, and a global flag inherited from the root.
func tree(t *testing.T) *cobra.Command {
	t.Helper()
	root := &cobra.Command{
		Use:           "prog <command>",
		Short:         "does things",
		Long:          "prog — does things\n\nAt some length.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           func(c *cobra.Command, _ []string) { c.Help() }, //nolint:errcheck
	}
	root.PersistentFlags().Bool("verbose", false, "say more")

	sub := &cobra.Command{
		Use:     "walk",
		Aliases: []string{"step"},
		Short:   "walks",
		Long:    "Walks somewhere.",
		Example: "  prog walk --far",
		Run:     func(*cobra.Command, []string) {},
	}
	sub.Flags().Bool("far", false, "go further")
	sub.Flags().IntP("count", "n", 3, "how many")
	root.AddCommand(sub, &cobra.Command{Use: "sit", Short: "sits", Run: func(*cobra.Command, []string) {}})

	InitFlags(root, true)
	return root
}

func help(t *testing.T, args ...string) string {
	t.Helper()
	root := tree(t)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return out.String()
}

// TestHelpHasEverySection is the whole point: cobra's template cannot run under
// TinyGo, so this output is built by hand and something has to check that
// building it by hand did not quietly drop a section.
func TestHelpHasEverySection(t *testing.T) {
	Color = false
	got := help(t, "walk", "--help")
	for _, want := range []string{
		"Walks somewhere.",
		"\nUsage:\n  prog walk [flags]\n",
		"\nAliases:\n  walk, step\n",
		"\nExamples:\n  prog walk --far\n",
		"\nFlags:\n",
		"  -n, --count int   how many (default 3)\n",
		"      --far         go further\n",
		"\nGlobal Flags:\n",
		"      --verbose   say more\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRootHelpListsSubcommands(t *testing.T) {
	Color = false
	got := help(t, "--help")
	for _, want := range []string{
		"prog — does things\n\nAt some length.\n\nUsage:\n  prog <command> [flags]\n",
		"\nAvailable Commands:\n",
		"  sit         sits\n",
		"  walk        walks\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	// --help is hidden, so the root has no visible local flags and no Flags
	// section; the persistent flag it defines is a global one on its children.
	if strings.Contains(got, "-h, --help") {
		t.Errorf("the hidden --help flag was listed:\n%s", got)
	}
}

// TestSectionsAreSeparatedButNotPadded — one blank line between sections, none
// trailing, so a command with nothing to list leaves no gap where it would be.
func TestSectionsAreSeparatedButNotPadded(t *testing.T) {
	Color = false
	got := help(t, "sit", "--help")
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("a doubled blank line:\n%q", got)
	}
	if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
		t.Errorf("should end in exactly one newline:\n%q", got)
	}
	if strings.Contains(got, "Examples:") || strings.Contains(got, "Available Commands:") {
		t.Errorf("sections it has nothing for:\n%s", got)
	}
}

// TestColorIsOptional: escapes belong in a terminal and nowhere else, and the
// listing has to line up either way — which means padding the name, not the
// name plus the escapes around it.
func TestColorIsOptional(t *testing.T) {
	Color = false
	plain := help(t, "--help")
	if strings.Contains(plain, "\x1b[") {
		t.Errorf("escapes with Color off:\n%q", plain)
	}

	Color = true
	defer func() { Color = false }()
	painted := help(t, "--help")
	if !strings.Contains(painted, styleHeading+"Available Commands:"+styleReset) {
		t.Errorf("no styled heading:\n%q", painted)
	}

	column := func(s, name string) int {
		for _, line := range strings.Split(s, "\n") {
			if bare := stripANSI(line); strings.HasPrefix(bare, "  "+name+" ") {
				return strings.Index(bare[2+len(name):], strings.TrimSpace(bare[2+len(name):])) + 2 + len(name)
			}
		}
		return -1
	}
	for _, name := range []string{"walk", "sit"} {
		if a, b := column(plain, name), column(painted, name); a != b || a < 0 {
			t.Errorf("%q description starts at column %d plain and %d painted", name, a, b)
		}
	}
}

// stripANSI removes the escapes, so a line can be measured as it is seen.
func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			for i += 2; i < len(s) && (s[i] == ';' || (s[i] >= '0' && s[i] <= '9')); i++ {
			}
			i++ // the final letter
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// TestHelpUsesNoTemplate is the regression itself. cobra renders help by
// executing a template, a template reaches its data by reflection, and TinyGo
// panics on the first method call — so the help must not go through cobra's
// template at all, on either host.
func TestHelpUsesNoTemplate(t *testing.T) {
	root := tree(t)
	for _, c := range append(root.Commands(), root) {
		var out bytes.Buffer
		c.SetOut(&out)
		// A template would render this marker; the renderer writes text.
		c.Long = "{{.Nonexistent}}"
		c.Help() //nolint:errcheck
		if !strings.Contains(out.String(), "{{.Nonexistent}}") {
			t.Errorf("%s: help went through a template:\n%s", c.Name(), out.String())
		}
	}
}
