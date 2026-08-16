# pisano

Designs made by reducing an integer sequence modulo m — the Pisano period made
visible. Two renderers, one core.

Built after Jacob Yatsko's *[A New Way to Look at Fibonacci
Numbers](https://www.youtube.com/watch?v=o1eLKODSCqw)*, including the questions
the video raises and leaves open.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-1-40-dark.png">
  <img alt="Fibonacci circular designs, moduli 1 to 40" src="docs/img/fib-1-40-light.png">
</picture>

Cobra CLI and a Bubble Tea v2 viewer; nothing else. Vendored, so it builds offline.

```
go build .        # -> ./pisano
go run . gallery
```

## What it draws

**Circular designs.** Space m points evenly around a circle, label them with the
possible remainders, and join them in the order the remainders appear. Output is
SVG, one figure or a whole contact sheet.

```
go run . circle --mod 1-40 -o sheet.svg
go run . circle --mod 8,13,21,34,55,89 --cols 3 --cell 260 -o families.svg
```

**Turtle paths.** Read each term as an instruction: odd turns left and steps
forward, even turns right and steps forward, zero does neither. Terminal text by
default, SVG or HTML with `--out`.

```
go run . turtle --mod 10                      # box-drawing characters
go run . turtle --mod 0                       # no modulus at all — the plus sign
go run . turtle --mod 1-40 -o paths.svg       # vector
go run . turtle --mod 1-40 -o paths.html      # page of inline SVG
go run . turtle --mod 8,21,55 --split web/     # one file each, for a site
```

**Watch it draw.** `tui` opens the designs in the alternate screen and animates
them. This is the one for paths that never close — see below.

```
go run . tui                       # fib mod 25, an open path, scrolling
go run . tui --mod 11              # a closed path, recolouring each lap
go run . tui --circle --mod 10     # trace the chords in sequence order
```

**Everything at once.** `gallery` builds a single self-contained page holding
every figure the video walks through, in the order it introduces them, and
writes the same figures out as standalone SVG sheets.

```
go run . gallery                    # out/index.html + out/svg/*.svg
```

**The arithmetic itself**, if you just want the numbers:

```
$ go run . period --mod 10
fib mod 10: period 60, 4 zero(s)
  terms:  [0 1 1 2 3 5 8 3 1 4 5 9 4 3 7 0 7 7 4 1 5 6 1 7 8 5 3 8 1 9 ...]
```

## Output formats

`--out` picks by extension: `.svg` writes one sheet, `.html` writes a page of
inline SVG, `-` writes SVG to stdout. `--split DIR` writes one file per figure,
which is what you want if the designs are going onto a site individually.

The terminal renderer is still there and still pipes to `ansifilter` if you want
the box-drawing look on a page:

```
go run . turtle --mod 10,25 | ansifilter -H -F "DejaVu Sans Mono" -s 10 > paths.html
```

That works, and it is how `tinygo-stuff/schematic.html` was made. It is the
wrong choice here, though. It gives you a `<pre>` of box-drawing glyphs, so the
figure depends on a font being present and correctly metric'd — which is why
that file carries 100 KB of base64 woff2 — it will not scale, and adjacent rows
show hairline gaps at most sizes. The native SVG has none of those problems and
is smaller. Use ansifilter when you specifically want the terminal's look;
use `--out` when you want the drawing.

## The viewer

The walk never finishes.

An **open path** drifts forever, so the camera follows its head and the oldest of
the drawing is dropped as it scrolls off the edge. The figure stays at its true
size rather than shrinking to fit — box-drawing characters at one lattice step
per row, joining up cleanly, however long it runs.

A **closed path** is simply walked again from where it left off. It retraces
exactly the same figure, so the geometry cannot change — but each circuit is
drawn in the next colour, and the loop sweeps round instead of sitting there
done. Colour is the only thing that can change on a closed figure, which is
precisely why it is worth changing.

```
go run . tui --mod 25            # open: scrolls
go run . tui --mod 11            # closed: recolours in place
go run . tui --trail comet       # a short tail chasing the head
```

| key | |
|---|---|
| `space` | play / pause |
| `← →` `h l` | modulus down / up (`H L` by ten) |
| `↑ ↓` `k j` | speed double / halve |
| `o` | jump to the next modulus whose path never closes |
| `t` | trail: whole circuit, long, short, comet |
| `0` | no modulus at all |
| `s` | next sequence · `[ ]` multiplier |
| `v` | turtle path or circular design |
| `m` | box-drawing characters or braille dots |
| `f` | camera: auto, fit, follow |
| `c` `r` `?` `q` | colour · restart · help · quit |

**Camera** defaults to auto, because the right answer is not a matter of taste:
an open path has a head that runs away from you and the view has to keep up,
while a closed one wants to sit still and be looked at. Auto picks `fit` for a
closed path and `scroll` for an open one.

Three ways to keep up, in increasing order of how much they leave alone:

| `--cam` | |
|---|---|
| `follow` | pins the head to the centre, so the whole drawing slides under it every frame |
| `scroll` | holds still until the head reaches a margin of the edge, then moves the least it can |
| `page` | same, but shoves the view most of a screen at a time — still for longer, then one jump |

Follow is exact and unwatchable: nothing on screen ever holds still, and the
head's wandering *within* a motif shakes the entire drawing. Under scroll that
wandering costs nothing — only the net drift moves the view, and only along the
axis it drifts on. Consecutive frames of the same walk:

```
follow                          scroll
     ┌─┬─┼─┼─┐                       ┌─┬─┼─┼─┐        cam -29,-6
     └─┼─┼─┼─┼─┐                     └─┼─┼─┼─┼─┐
     ┌─┤ ├─┤ └─┤                 ┌─╴ ┌─┤ ├─┤ └─┤
   ┌─┴─┘ └─┘ ┌─┼─┼─              └─┐ ┌─┼─┼─┘ └─┘

  ┌─┬─┼─┼─┐                          ┌─┬─┼─┼─┐        cam -29,-6
  └─┼─┼─┼─┼─┐                        └─┼─┼─┼─┼─┐
  ┌─┤ ├─┤ └─┤                    ┌─┐ ┌─┤ ├─┤ └─┤
┌─┴─┘ └─┘ ┌─┼─┼─                 ├─┤ ┌─┼─┼─┘ └─┘
   ^ the motif jumps                 ^ same place, new drawing only
```

**Trail** is how much stays on screen. `whole` keeps exactly one circuit for a
closed path — so the figure stands complete and still while the colour moves
through it — and as much as memory allows for an open one. Shorter settings
leave a comet tail behind the head. Trimming runs in batches: each trim shifts
the slice down, so dropping one point per frame would copy the whole trail for
the sake of one element.

**Two renderers.** Box drawing is the default and is the figure at its true
size: one lattice step is one row or two columns, which is why the characters
join. It cannot zoom, so the camera is the only thing that moves and anything
outside the window is clipped — which is exactly what makes an endless path
watchable, since the drawing scrolls rather than shrinking.

Braille packs a 2x4 dot grid into every cell (`U+2800`), trading those joins for
an arbitrary scale, so it can keep shrinking as an open path drifts. It is also
the only one that can draw the circular designs at all, since those are chords
at arbitrary angles — press `v` to watch one trace itself chord by chord, in the
order the remainders actually appear, which the finished figure discards.

A braille dot is half a cell wide and a quarter tall, and a terminal cell is
roughly twice as tall as it is wide, so the dots come out very nearly square and
one scale serves both axes.

`--max-points` (60000) caps memory. At that size a redraw costs about 3.6 ms in
box mode and 7.8 ms in braille, against a 40 ms frame.

### Don't loop it from the shell

`--cycle 5s` steps through the moduli on its own. Use it rather than a shell
loop around the whole program:

```
go run . tui --cycle 5s --mod 1          # yes
seq 0 1000 | while read n; do timeout 5 ./pisano tui --mod $n; done   # no
```

A full-screen program takes the terminal over and puts it in **raw mode**, where
Ctrl-C is no longer a signal — it is a keystroke the program has to choose to
honour. In a loop that means each Ctrl-C ends one iteration and the shell
immediately starts the next, so a thousand of them takes a thousand interrupts
to escape. `tui` therefore refuses to start unless stdin and stdout are both
terminals.

### Why Bubble Tea v2

Bubble Tea v1 shipped a package `init()` that called `lipgloss.HasDarkBackground()`,
which asks the terminal for its background colour (`OSC 11`) and waits up to
`termenv.OSCTimeout` — **five seconds** — for a reply. Because it was package
initialisation it ran before `main()`, so it stalled *every* command in the
binary, `period` and `gallery` included, on any terminal that does not answer.

v2 deletes the workaround along with its cause: no `init()` anywhere in the
module, and no dependency on lipgloss or termenv at all. Background colour is an
ordinary opt-in command whose answer arrives as a message, so nothing blocks and
a terminal that never replies simply never sends one.

Measured on a terminal that does not answer:

| | v1.3.10 | v2.0.8 |
|---|---|---|
| `pisano period --mod 10` | 5.01 s | 0.01 s |
| `pisano --help` | 5.01 s | 0.01 s |
| `pisano gallery` | 5.04 s | 0.03 s |
| `pisano tui` | 12 bytes, drew nothing | draws immediately |
| terminal after `SIGTERM` | left raw: `ICANON=off ECHO=off` | fully restored |

The port also dropped lipgloss and termenv from the tree. The chrome is styled
with raw SGR codes now, which is what the canvas renderers were already doing.

## Sequences

`--seq fib` (default), `lucas`, `tri`, `nat`, `prime`. Nothing about either
construction is particular to Fibonacci — it is the ingredient the recipe is
usually made with, not the recipe.

`nat` is the plain number line, the reference case both constructions are
introduced against: its circular designs are regular polygons and its turtle
path is a staircase, because it alternates odd and even forever.

The primes are the instructive case: their residues never repeat, so no period
exists and the drawing is an open path rather than a closed figure. The code
does not assume periodicity; a sequence declares whether it has a finite state,
and one that does not never reports a period.

`--mul n` multiplies the Fibonacci sequence through by n. The plain designs then
reappear at every nth modulus with new ones interleaved between them, which is
the video's "design m/k" labelling: m is the modulus, n the multiplier. It is
not "mod m/n" — a modulus is always an integer.

## The open questions

The video notices two patterns it cannot account for. Both are testable, so
`sweep` tests them.

**Does the zero count predict symmetry?** Every Fibonacci Pisano period contains
one, two or four zeros. The video observes that four zeros always seem to give a
symmetrical design and one zero always an asymmetrical one, and says it would be
nice to know why.

```
$ go run . sweep --max 3000
ZEROS     SYMMETRIC    ASYMMETRIC
-----     ---------    ----------
1                 0           335
2               324          1981
4               358             0

no counterexamples: 4 zeros => symmetrical and 1 zero => asymmetrical
held for every modulus 3..3000
```

Not a proof. But a single counterexample would have settled it the other way,
and none turns up, which is worth knowing before anyone spends time proving it.

**What decides whether a turtle path closes?** The video sees that some paths
loop and others march off to infinity, and says it never looked into why. This
one does have an answer, and it needs no search.

One pass of the period is a rigid motion of the plane: turn through some whole
number of right angles, then translate. Repeating the period repeats that
motion, so the net turn decides everything.

| net turn | result |
|---|---|
| 90° or 270° | rotation about a fixed point — closes after 4 passes |
| 180° | rotation — closes after 2 passes |
| 0° | pure translation — closes only if the drift is also zero, otherwise open forever |

`Classify` reads this off a single pass without walking the path, and the test
suite checks its verdict against actually walking it for every modulus to 200.

The video also notes that several moduli produce identical turtle designs.
`sweep --dupes` canonicalises each path up to position and orientation and groups
them:

```
$ go run . sweep --max 60 --dupes
43 distinct figures across 60 moduli
  4, 8
  6, 12, 16, 18, 24, 36, 48
  ...
```

## In the browser

The same designs, driven from a shell, compiled to WebAssembly:
**[0magnet.github.io/pisano/term/](https://0magnet.github.io/pisano/term/)**
(TinyGo; the [standard Go build](https://0magnet.github.io/pisano/term/go/) is
linked from the page header)

```
pisano:~$ pisano tui                                 the viewer, full screen
pisano:~$ pisano turtle --mod 25                     box drawing, in the shell
pisano:~$ pisano sweep --max 300                     the analysis
pisano:~$ pisano circle --mod 1-40 -o sheet.svg && download sheet.svg
```

The terminal and the shell are **[websh](https://github.com/0magnet/websh)** —
xterm-go, a real Bash interpreter, and a virtual filesystem, all already
compiled to wasm. This repo adds one applet to it. So the pipes, globs, history
and tab completion are the shell's, `download` is websh's, and the figures are
this package's, unchanged.

The viewer running there is the *same* `tui.Model` the desktop binary runs. What
differs is only what feeds it:

```
                 Resize / Advance / Key / Frame
                              |
        +---------------------+---------------------+
        |                                           |
   Bubble Tea                                  websh applet
   (native, pkg/tui/bubbletea.go)              (wasm, web/applet.go)
   key messages, a Tick command                raw bytes, a time.Ticker
```

Four calls, one model, one set of tests covering both. `pkg/tui` has no
dependency on Bubble Tea except in that one build-tagged file — which is what
lets it compile for js/wasm at all, since Bubble Tea has no port there.

### On a desktop

**[0magnet.github.io/pisano/desk/](https://0magnet.github.io/pisano/desk/)** puts
the same shell in a window, on [desk](https://github.com/0magnet/desk), so a
design can be generated in one window and opened in another:

```
pisano circle --mod 8,13,21,34 --cols 4 -o sheet.svg && view sheet.svg
```

There is no message passing behind that. Both windows are panes in one binary
over one filesystem, so writing the file *is* the hand-off and `view` only has
to name it. The dependency points pisano → desk; desk knows nothing about
pisano, which is why the command lives in `web/app` and both binaries register
it.

Building it:

```
web/build.sh          # both     -> docs/term and docs/term/go
web/build.sh tinygo   # TinyGo only                   3.5 MB
web/build.sh go       # standard Go only               13 MB
```

Both are carried and the page header links between them. TinyGo is the default
because it is a quarter the size and is fetched before anything appears; the
standard Go build is there because TinyGo occasionally miscompiles something,
and having the other one a click away is how you find out that is what
happened.

`web/` is a separate module. Keeping it out of the root one means the CLI's
dependencies stay at cobra and Bubble Tea, rather than dragging in a terminal
emulator and a shell interpreter for a build almost nobody makes.

## The figures

Every sheet below is produced by `pisano gallery`, which also writes them as
standalone SVG and builds [`docs/index.html`](docs/index.html) holding the lot.
The images follow your system's light or dark setting.

### The number line, moduli 1–9

The reference case: reduced mod m the naturals cycle through every remainder in order, so the chords always close into a regular polygon.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/nat-1-9-dark.png">
  <img alt="The number line, moduli 1–9" src="docs/img/nat-1-9-light.png">
</picture>

<sub>[SVG](docs/svg/nat-1-9.svg)</sub>

### Fibonacci, moduli 1–9

The same construction on Fibonacci remainders. Mod 5 uses every possible chord; mod 6 keeps a mirror axis; mod 8 has none.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-1-9-dark.png">
  <img alt="Fibonacci, moduli 1–9" src="docs/img/fib-1-9-light.png">
</picture>

<sub>[SVG](docs/svg/fib-1-9.svg)</sub>

### Fibonacci, modulus 10

A period of 60 — far more chords than its neighbours, and still symmetrical.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-10-dark.png">
  <img alt="Fibonacci, modulus 10" src="docs/img/fib-10-light.png">
</picture>

<sub>[SVG](docs/svg/fib-10.svg)</sub>

### Fibonacci moduli 8, 21, 55

Near-duplicates of one another. All three moduli are Fibonacci numbers, all three periods hold two zeros, none is symmetrical.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-family-a-dark.png">
  <img alt="Fibonacci moduli 8, 21, 55" src="docs/img/fib-family-a-light.png">
</picture>

<sub>[SVG](docs/svg/fib-family-a.svg)</sub>

### Fibonacci moduli 13, 34, 89

The other family, alternating with the first down the Fibonacci numbers. Four zeros each, symmetrical each.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-family-b-dark.png">
  <img alt="Fibonacci moduli 13, 34, 89" src="docs/img/fib-family-b-light.png">
</picture>

<sub>[SVG](docs/svg/fib-family-b.svg)</sub>

### Lucas moduli 7, 18, 47

The same splitting into two families happens to the Lucas numbers, at moduli that are themselves Lucas numbers.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/lucas-family-a-dark.png">
  <img alt="Lucas moduli 7, 18, 47" src="docs/img/lucas-family-a-light.png">
</picture>

<sub>[SVG](docs/svg/lucas-family-a.svg)</sub>

### Lucas moduli 11, 29, 76

The second Lucas family.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/lucas-family-b-dark.png">
  <img alt="Lucas moduli 11, 29, 76" src="docs/img/lucas-family-b-light.png">
</picture>

<sub>[SVG](docs/svg/lucas-family-b.svg)</sub>

### Five sequences, all at modulus 10

Fibonacci, Lucas, triangular, the naturals and the primes. Fibonacci is the ingredient, not the recipe — and the primes never repeat, so there is no loop to close.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/sequences-10-dark.png">
  <img alt="Five sequences, all at modulus 10" src="docs/img/sequences-10-light.png">
</picture>

<sub>[SVG](docs/svg/sequences-10.svg)</sub>

### Fibonacci × 2, moduli 1–24

Multiplying the sequence through by 2 makes the plain designs reappear at every 2nd modulus, with new ones interleaved.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-x2-1-24-dark.png">
  <img alt="Fibonacci × 2, moduli 1–24" src="docs/img/fib-x2-1-24-light.png">
</picture>

<sub>[SVG](docs/svg/fib-x2-1-24.svg)</sub>

### Fibonacci × 3, moduli 1–24

Every 3rd modulus, and so on — the "design m/k" labelling.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-x3-1-24-dark.png">
  <img alt="Fibonacci × 3, moduli 1–24" src="docs/img/fib-x3-1-24-light.png">
</picture>

<sub>[SVG](docs/svg/fib-x3-1-24.svg)</sub>

### Fibonacci × 4, moduli 1–24

Every 4th. The designs of ×2 show up at every other one of these.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-x4-1-24-dark.png">
  <img alt="Fibonacci × 4, moduli 1–24" src="docs/img/fib-x4-1-24-light.png">
</picture>

<sub>[SVG](docs/svg/fib-x4-1-24.svg)</sub>

### Fibonacci, moduli 1–40

The contact sheet. Length alone does not decide whether a figure comes out symmetrical — the zero count does.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-1-40-dark.png">
  <img alt="Fibonacci, moduli 1–40" src="docs/img/fib-1-40-light.png">
</picture>

<sub>[SVG](docs/svg/fib-1-40.svg)</sub>

### Fibonacci, moduli 41–80

Further out, where the near-duplicate families keep recurring.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/fib-41-80-dark.png">
  <img alt="Fibonacci, moduli 41–80" src="docs/img/fib-41-80-light.png">
</picture>

<sub>[SVG](docs/svg/fib-41-80.svg)</sub>

### Turtle paths with no modulus at all

The naturals make a staircase. Fibonacci runs odd, odd, even — left, left, right — and closes into something that looks like a plus sign.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/turtle-unreduced-dark.png">
  <img alt="Turtle paths with no modulus at all" src="docs/img/turtle-unreduced-light.png">
</picture>

<sub>[SVG](docs/svg/turtle-unreduced.svg)</sub>

### Fibonacci turtle paths, moduli 1–40

Now the remainders drive the turns. Some close; others repeat a motif off to infinity, drawn here for three passes.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/turtle-fib-1-40-dark.png">
  <img alt="Fibonacci turtle paths, moduli 1–40" src="docs/img/turtle-fib-1-40-light.png">
</picture>

<sub>[SVG](docs/svg/turtle-fib-1-40.svg)</sub>

### Fibonacci turtle paths, moduli 41–80

The same, further out.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/turtle-fib-41-80-dark.png">
  <img alt="Fibonacci turtle paths, moduli 41–80" src="docs/img/turtle-fib-41-80-light.png">
</picture>

<sub>[SVG](docs/svg/turtle-fib-41-80.svg)</sub>

### Turtle paths shared between moduli

6, 12, 16, 18, 24, 36 and 48 all draw one figure; 14, 28, 32, 42 and 46 draw another.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/turtle-twins-dark.png">
  <img alt="Turtle paths shared between moduli" src="docs/img/turtle-twins-light.png">
</picture>

<sub>[SVG](docs/svg/turtle-twins.svg)</sub>

### Lucas turtle paths, moduli 1–24

The same walk over a different sequence.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/turtle-lucas-1-24-dark.png">
  <img alt="Lucas turtle paths, moduli 1–24" src="docs/img/turtle-lucas-1-24-light.png">
</picture>

<sub>[SVG](docs/svg/turtle-lucas-1-24.svg)</sub>

### Triangular turtle paths, moduli 1–24

And over the triangular numbers, whose rule for making terms has nothing to do with the other two.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/img/turtle-tri-1-24-dark.png">
  <img alt="Triangular turtle paths, moduli 1–24" src="docs/img/turtle-tri-1-24-light.png">
</picture>

<sub>[SVG](docs/svg/turtle-tri-1-24.svg)</sub>

## Layout

A thin `package main` at the repo root, the CLI as
a cobra tree under `cmd/`, and the actual work in `pkg/`.

```
pisano.go                     package main — wires up styling, calls Execute
cmd/pisano/commands/          the cobra tree: root, period, circle, turtle,
                              sweep, gallery
pkg/flags/                    coloredcobra help templates and styling
pkg/pisano/                   the library — no CLI, no cobra, no bubbletea
pkg/tui/                      the bubbletea viewer
```

`RootCmd` is exported, so the whole tree can be grafted onto another binary's
root command the way skywire composes its subcommands.

## Design

`Period` is the only thing the renderers see, so a drawing cannot disagree with
the arithmetic it came from. Everything else is a consequence:

- `pisano.go` — sequences and period detection. Cycles are found by generator
  state, not by assuming a recurrence, which is what lets the primes honestly
  report no period instead of a wrong one.
- `unreduced.go` — reading a sequence with no modulus. This is *not* the same as
  reducing mod 2: mod 2 sends every even term to the residue 0, which the turtle
  reads as "stay", while unreduced only a genuine zero stays. Fibonacci
  unreduced gives the left, left, right cycle behind the plus-sign figure; mod 2
  gives something else. Parity is tracked separately from value so that
  saturating the values (a Fibonacci number outruns int64 in ninety terms) never
  costs accuracy — the only question ever asked of the value is whether it is
  zero.
- `design.go` — chords and the symmetry tests.
- `sheet.go` — tiles, sheets and the HTML page. Both renderers emit a `Tile`,
  which is what lets circular designs and turtle paths share a sheet and go to
  SVG or HTML without either renderer knowing which.
- `svg.go`, `turtlesvg.go` — vector renderers.
- `canvas.go`, `braille.go`, `turtle.go` — the walk and the two terminal
  renderers.

### Why SVG for the circles and both for the turtle

The circular designs are chords at arbitrary angles. A character grid cannot
hold them, and a plotting library cannot draw them at all: plotting draws y as a
function of x, so a line can never double back.

Turtle paths are axis-aligned unit steps on a lattice, which is exactly what box
drawing characters are for — so they get a terminal renderer as well. The two
draw the same walk from the same points. The terminal one is bound to a
character grid, so a horizontal unit has to span two cells; the vector one
scales the lattice to whatever box it is given.

### canvas.go

Adapted from the wire canvas in `tinygo-stuff/schematic`, with three changes:

- **A cell stores which of its four edges carry track, not a finished glyph.**
  The original resolved a corner at the moment it drew one, which cannot express
  a path crossing its own earlier self — and these paths do that constantly. All
  sixteen edge combinations have a glyph, so T-junctions and crossings come out
  right where the original's `elbow` had no name for them and fell back to a
  placeholder.
- **The origin can be negative.** A turtle wanders wherever it likes, so the
  canvas is told its bounds up front and translates into them.
- **A horizontal unit spans two columns**, because terminal cells are about
  twice as tall as they are wide and a one-for-one grid comes out squashed.

None of the original's terminal-cursor interrogation came along. That existed
because asciigraph emits uneven-width ANSI output and the drawing had to ask the
terminal where the cursor had ended up; a canvas that knows its own cells has
nothing to ask.
