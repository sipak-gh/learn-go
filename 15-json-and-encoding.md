# Chapter 15 — Encoding: JSON and Struct Tags

Programs often talk to other programs, most commonly in JSON. Go's `encoding/json` turns Go values
into JSON and back using your existing struct definitions, with small annotations called **struct
tags** controlling the details. You'll learn encoding and decoding, struct tags, streaming with
`Encoder` and `Decoder`, dynamic JSON, and custom formats, plus a short look at `encoding/csv` and
`encoding/base64`. Then `seek` gets a `--json` option so other tools can read its results.

Previous: [Chapter 14 — Iterators and Range-over-Func](14-iterators.md). Next: [Chapter 16 —
Goroutines and Channels](16-goroutines-and-channels.md).

## Marshal: Go to JSON

`json.Marshal` returns any value's JSON encoding as a `[]byte`. Each exported struct field becomes a
key in a JSON object.

`examples/ch15/marshal/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Book struct {
	Title   string   `json:"title"`
	Author  string   `json:"author,omitempty"` // left out when ""
	Year    int      `json:"year"`
	Tags    []string `json:"tags"`
	ISBN    string   `json:"-"` // never encoded
	Pages   int      // no tag: the field name is used as is
	private string   // unexported: invisible to encoding/json
}

func main() {
	b := Book{Title: "The Go Programming Language", Year: 2015,
		Tags: []string{"go", "programming"}, ISBN: "978-0134190440", Pages: 380, private: "x"}
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(data))

	// A nil slice becomes null; an empty slice becomes [].
	var none []string
	n, _ := json.Marshal(none)
	e, _ := json.Marshal([]string{})
	fmt.Println(string(n), string(e))

	// Maps become objects, with the keys sorted.
	m, _ := json.Marshal(map[string]int{"b": 2, "a": 1, "c": 3})
	fmt.Println(string(m))

	// Indented output, for humans.
	pretty, _ := json.MarshalIndent(map[string]any{"ok": true, "n": []int{1, 2}}, "", "  ")
	fmt.Println(string(pretty))

	// Some values have no JSON form.
	_, err = json.Marshal(make(chan int))
	fmt.Println("error:", err)
}
```

```console
$ go run ./ch15/marshal
{"title":"The Go Programming Language","year":2015,"tags":["go","programming"],"Pages":380}
null []
{"a":1,"b":2,"c":3}
{
  "n": [
    1,
    2
  ],
  "ok": true
}
error: json: unsupported type: chan int
```

How Go types map to JSON:

| Go | JSON |
|---|---|
| `bool` | `true` / `false` |
| integers, floats | number |
| `string` | string (with escapes) |
| struct | object; exported fields only |
| `map[string]T` | object, keys sorted |
| slice, array | array; a nil slice is `null` |
| `[]byte` | base64 string |
| pointer | the value it points to, or `null` for nil |
| `any` (interface) | the value it holds |
| channel, function, complex | error |

### Struct tags

A **struct tag** is a string literal after a field's type, in back quotes because it contains double
quotes. Packages read tags at run time to learn how to treat each field. The `json` key holds the JSON
name, then comma-separated options:

- `json:"title"` uses `title` as the key instead of `Title`.
- `json:"author,omitempty"` omits the field when it's `""`, `0`, `false`, `nil`, or an empty slice or
  map.
- `json:"-"` never encodes or decodes the field.
- No tag: the key is the field name, `Pages`.

Only **exported** fields (capitalized, [Chapter 8](08-packages-and-modules.md)) are encoded.
`encoding/json` is another package, so it can't see `private` at all. A struct with only unexported
fields encodes as `{}`, a common surprise.

`go vet` checks struct tag syntax, catching typos like `json:title` (missing quotes).

## Unmarshal: JSON to Go

`json.Unmarshal(data, &v)` parses JSON into `v`. It needs a **pointer** so it can change your
variable.

`examples/ch15/unmarshal/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Book struct {
	Title string   `json:"title"`
	Year  int      `json:"year"`
	Tags  []string `json:"tags"`
}

func main() {
	data := []byte(`{"TITLE": "Go", "year": 2015, "publisher": "AW"}`)
	var b Book
	if err := json.Unmarshal(data, &b); err != nil { // note the &
		fmt.Println("error:", err)
		return
	}
	// "TITLE" matched title (case-insensitive), "publisher" was ignored,
	// and Tags kept its zero value.
	fmt.Printf("%+v\n", b)

	// Unmarshal only sets the fields present in the input.
	_ = json.Unmarshal([]byte(`{"tags": ["go"]}`), &b)
	fmt.Printf("%+v\n", b)

	var errs = []string{
		`{"year": "2015"}`, // wrong type
		`{"year": 2015,}`,  // trailing comma
		`{"title": "Go"`,   // cut short
	}
	for _, in := range errs {
		var b Book
		err := json.Unmarshal([]byte(in), &b)
		fmt.Println("error:", err)
	}
}
```

```console
$ go run ./ch15/unmarshal
{Title:Go Year:2015 Tags:[]}
{Title:Go Year:2015 Tags:[go]}
error: json: cannot unmarshal string into Go struct field Book.year of type int
error: invalid character '}' looking for beginning of object key string
error: unexpected end of JSON input
```

- Keys match field names or tags **case-insensitively**, preferring an exact match.
- Unknown keys are ignored; missing keys leave the field unchanged. So decoding into an existing value
  *merges* into it, as the second call shows. Use a fresh variable to avoid that.
- Errors say what went wrong and where.

Passing the struct instead of a pointer is a common mistake that `go vet` catches:

```console
$ go vet
./main.go:14:23: call of Unmarshal passes non-pointer as second argument
```

Run anyway, `Unmarshal` returns `json: Unmarshal(non-pointer main.Book)`.

## Dynamic JSON: any and map[string]any

When you don't know the data's shape in advance, decode into `any`. Objects become `map[string]any`,
arrays `[]any`, and the rest `string`, `float64`, `bool`, or `nil`.

`examples/ch15/dynamic/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func main() {
	data := []byte(`{"name": "seek", "stars": 42, "tags": ["cli", "go"], "meta": {"ok": true}, "x": null}`)

	// Decoding into any gives maps, slices, and basic types.
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println("error:", err)
		return
	}
	obj := v.(map[string]any)
	for _, k := range []string{"name", "stars", "tags", "meta", "x"} {
		fmt.Printf("%-5s %-24T %v\n", k, obj[k], obj[k])
	}

	// Every JSON number becomes a float64. To use one as an int,
	// convert it, after a type assertion.
	stars := int(obj["stars"].(float64))
	fmt.Println(stars + 1)

	// Big integers lose precision as float64. A Decoder with UseNumber
	// keeps the original text as a json.Number instead.
	big := `{"id": 12345678901234567890}`
	var m1 map[string]any
	_ = json.Unmarshal([]byte(big), &m1)
	fmt.Println(m1["id"])

	dec := json.NewDecoder(strings.NewReader(big))
	dec.UseNumber()
	var m2 map[string]any
	_ = dec.Decode(&m2)
	fmt.Printf("%T %v\n", m2["id"], m2["id"])
}
```

```console
$ go run ./ch15/dynamic
name  string                   seek
stars float64                  42
tags  []interface {}           [cli go]
meta  map[string]interface {}  map[ok:true]
x     <nil>                    <nil>
43
1.2345678901234567e+19
json.Number 12345678901234567890
```

- Every number is a `float64`, even `42`. A `float64` holds integers exactly only up to 2^53, so large
  IDs get rounded. `UseNumber` keeps them as `json.Number`, a string type with `Int64` and `Float64`
  methods.
- Using the values needs type assertions ([Chapter 10](10-interfaces.md)) at every step: tedious and
  error-prone. When you know the shape, prefer a struct, even a partial one with just the fields you
  need.

## Streams: Encoder and Decoder

`Marshal` and `Unmarshal` work on whole `[]byte` values. `json.NewEncoder(w)` and `json.NewDecoder(r)`
work on an `io.Writer` or `io.Reader` ([Chapter 11](11-io-and-files.md)), one value at a time.

`examples/ch15/stream/main.go`

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Event struct {
	Kind string `json:"kind"`
	Msg  string `json:"msg"`
}

func main() {
	// An Encoder writes one JSON value per Encode call, plus a newline.
	// That's the JSON Lines format.
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(Event{"start", "a <b> & c"})
	enc.SetEscapeHTML(false)
	_ = enc.Encode(Event{"start", "a <b> & c"})
	enc.SetIndent("", "  ")
	_ = enc.Encode(Event{"stop", "bye"})

	// A Decoder reads values one after another, until io.EOF.
	input := `{"kind":"a","msg":"one"}
{"kind":"b","msg":"two"}
{"kind":"c","msg":"three","extra":1}`
	dec := json.NewDecoder(strings.NewReader(input))
	dec.DisallowUnknownFields()
	for {
		var e Event
		err := dec.Decode(&e)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("error:", err)
			break
		}
		fmt.Printf("%+v\n", e)
	}
}
```

```console
$ go run ./ch15/stream
{"kind":"start","msg":"a \u003cb\u003e \u0026 c"}
{"kind":"start","msg":"a <b> & c"}
{
  "kind": "stop",
  "msg": "bye"
}
{Kind:a Msg:one}
{Kind:b Msg:two}
error: json: unknown field "extra"
```

- **JSON Lines** (or NDJSON) is one JSON value per line: easy to produce as you go and to process line
  by line with other tools. `Encode` writes exactly this.
- By default `<`, `>`, and `&` are escaped as `\u003c`, `\u003e`, and `\u0026`. That's still valid
  JSON decoding to the same text, and it protects pages that embed JSON in HTML. Elsewhere,
  `SetEscapeHTML(false)` keeps the text readable.
- `SetIndent` pretty-prints, like `MarshalIndent`.
- `Decode` returns `io.EOF` when input ends cleanly between values. Any other error means bad input.
- `DisallowUnknownFields` makes unknown keys errors, useful for config files where a mistyped key
  shouldn't go unnoticed.

## Custom encodings: Marshaler and Unmarshaler

A type controls its own JSON form by implementing these interfaces:

```go
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}
type Unmarshaler interface {
	UnmarshalJSON([]byte) error
}
```

`encoding/json` checks every value for them, as `fmt` checks for `String()`. For types that are
naturally one string, `encoding.TextMarshaler` (`MarshalText`) and `TextUnmarshaler` (`UnmarshalText`)
are simpler, and they work for map keys too.

`examples/ch15/custom/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Duration is a time.Duration that is written as a string like "1m30s"
// in JSON, instead of a number of nanoseconds.
type Duration struct {
	time.Duration
}

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler. It needs a pointer
// receiver, because it changes d.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = v
	return nil
}

// Level is written as text. TextMarshaler also works for map keys.
type Level int

func (l Level) MarshalText() ([]byte, error) {
	return []byte([]string{"debug", "info", "warn"}[l]), nil
}

func (l *Level) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "debug":
		*l = 0
	case "info":
		*l = 1
	case "warn":
		*l = 2
	default:
		return fmt.Errorf("unknown level %q", text)
	}
	return nil
}

type Config struct {
	Timeout Duration      `json:"timeout"`
	Level   Level         `json:"level"`
	Counts  map[Level]int `json:"counts"`
}

func main() {
	c := Config{Timeout: Duration{90 * time.Second}, Level: 1,
		Counts: map[Level]int{0: 5, 2: 1}}
	data, _ := json.Marshal(c)
	fmt.Println(string(data))

	var back Config
	err := json.Unmarshal([]byte(`{"timeout":"250ms","level":"WARN"}`), &back)
	fmt.Println(back.Timeout, back.Level, err)

	err = json.Unmarshal([]byte(`{"timeout":5}`), &back)
	fmt.Println("error:", err)
	err = json.Unmarshal([]byte(`{"level":"loud"}`), &back)
	fmt.Println("error:", err)
}
```

```console
$ go run ./ch15/custom
{"timeout":"1m30s","level":"info","counts":{"debug":5,"warn":1}}
250ms 2 <nil>
error: duration must be a string: json: cannot unmarshal number into Go value of type string
error: unknown level "loud"
```

- `Duration` embeds `time.Duration` ([Chapter 7](07-structs-methods-pointers.md)), so it has its
  `String` method (`time` is in [Chapter 18](18-context-and-time.md)). `MarshalJSON` makes a Go string
  and lets `json.Marshal` add quotes and escapes; hand-built JSON via string concatenation is easy to
  get wrong.
- `UnmarshalJSON` receives the raw JSON bytes, here `"250ms"` with quotes. Decoding them into a
  `string` first handles quotes and escapes.
- `Unmarshal` methods use pointer receivers. `Marshal` methods use value receivers, so they work on
  values and pointers.
- `back.Level` prints as `2` because `fmt` ignores `MarshalText`. Add a `String` method for that.
- Inside `MarshalJSON`, don't call `json.Marshal(d)` on the *same* type: it would call `MarshalJSON`
  again, forever, until the stack overflows. Convert to another type first, as `d.String()` does.

## Delaying decoding: json.RawMessage

Sometimes one field says how to decode another. `json.RawMessage` is a `[]byte` holding a piece of
JSON undecoded, for decoding later:

`examples/ch15/raw/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
)

// An envelope's Data depends on its Type, so it is kept as raw JSON
// and decoded later, once we know what it is.
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Point struct{ X, Y int }

type Text struct {
	Body string `json:"body"`
}

func main() {
	input := []byte(`[
		{"type": "point", "data": {"X": 1, "Y": 2}},
		{"type": "text",  "data": {"body": "hi"}},
		{"type": "sound", "data": [1, 2, 3]}
	]`)
	var envs []Envelope
	if err := json.Unmarshal(input, &envs); err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, e := range envs {
		switch e.Type {
		case "point":
			var p Point
			_ = json.Unmarshal(e.Data, &p)
			fmt.Printf("point %+v\n", p)
		case "text":
			var t Text
			_ = json.Unmarshal(e.Data, &t)
			fmt.Printf("text %q\n", t.Body)
		default:
			fmt.Printf("unknown %s: %s\n", e.Type, e.Data)
		}
	}
}
```

```console
$ go run ./ch15/raw
point {X:1 Y:2}
text "hi"
unknown sound: [1, 2, 3]
```

`RawMessage` keeps the bytes exactly as they appeared, spaces included. Marshaling writes them back
compacted (spaces removed), but otherwise unchanged.

## Times in JSON

`time.Time` implements the marshaling interfaces using RFC 3339, such as `2025-02-11T09:30:00Z`.

`examples/ch15/timejson/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Post struct {
	Title     string    `json:"title"`
	Published time.Time `json:"published"`
	Edited    time.Time `json:"edited,omitzero"` // left out when zero
	Draft     time.Time `json:"draft,omitempty"` // omitempty can't tell
}

func main() {
	p := Post{
		Title:     "Hello",
		Published: time.Date(2025, 2, 11, 9, 30, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(p)
	fmt.Println(string(data))

	var back Post
	err := json.Unmarshal([]byte(`{"published":"2025-02-11T10:30:00+01:00"}`), &back)
	fmt.Println(back.Published, err)
	fmt.Println(back.Published.Equal(p.Published))

	err = json.Unmarshal([]byte(`{"published":"11/02/2025"}`), &back)
	fmt.Println("error:", err)
}
```

```console
$ go run ./ch15/timejson
{"title":"Hello","published":"2025-02-11T09:30:00Z","draft":"0001-01-01T00:00:00Z"}
2025-02-11 10:30:00 +0100 +0100 <nil>
true
error: parsing time "11/02/2025" as "2006-01-02T15:04:05Z07:00": cannot parse "11/02/2025" as "2006"
```

- `omitempty` ignores structs, and `time.Time` is a struct, so the zero time is written as
  `0001-01-01T00:00:00Z`.
- `omitzero` (Go 1.24) omits a field that is its type's zero value or whose `IsZero()` method returns
  true. Use it for times and other structs.
- The time zone offset is kept. Compare times with `Equal`, not `==`: 10:30 at +01:00 is the same
  moment as 09:30 UTC.

### A note on encoding/json/v2

Go 1.26 has a redesigned `encoding/json/v2`, but only as an experiment built with
`GOEXPERIMENT=jsonv2`. This chapter uses the stable `encoding/json`, which isn't going away.

## Other encodings: CSV and base64

`encoding` has more packages; two are common in command-line tools.

`encoding/csv` reads and writes comma-separated values, including quoted fields:

`examples/ch15/csv/main.go`

```go
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func main() {
	input := `name,lang,note
seek,Go,"search, fast"
ripgrep,Rust,"says ""hi"""
`
	r := csv.NewReader(strings.NewReader(input))
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, rec := range records[1:] { // skip the header
		fmt.Printf("%q\n", rec)
	}

	w := csv.NewWriter(os.Stdout)
	_ = w.Write([]string{"path", "count"})
	_ = w.Write([]string{"a file, with comma.txt", "3"})
	w.Flush() // csv.Writer is buffered
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}
}
```

```console
$ go run ./ch15/csv
["seek" "Go" "search, fast"]
["ripgrep" "Rust" "says \"hi\""]
path,count
"a file, with comma.txt",3
```

Every record is a `[]string`; CSV has no types. Like `bufio.Writer`, `csv.Writer` must be flushed;
check its errors afterwards with `Error`. For big files, loop on `r.Read()` instead of `ReadAll`.

`encoding/base64` turns any bytes into plain ASCII text:

`examples/ch15/base64/main.go`

```go
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Blob struct {
	Name string `json:"name"`
	Data []byte `json:"data"` // []byte is encoded as a base64 string
}

func main() {
	raw := []byte("hi?\x00\xff")
	s := base64.StdEncoding.EncodeToString(raw)
	fmt.Println(s)
	fmt.Println(base64.URLEncoding.EncodeToString(raw)) // safe in URLs
	back, err := base64.StdEncoding.DecodeString(s)
	fmt.Printf("%q %v\n", back, err)

	data, _ := json.Marshal(Blob{"tiny", raw})
	fmt.Println(string(data))
	var b Blob
	_ = json.Unmarshal(data, &b)
	fmt.Printf("%q\n", b.Data)
}
```

```console
$ go run ./ch15/base64
aGk/AP8=
aGk_AP8=
"hi?\x00\xff"
{"name":"tiny","data":"aGk/AP8="}
"hi?\x00\xff"
```

JSON strings must be valid text, so `encoding/json` stores a `[]byte` field as base64 and decodes it
back automatically. `URLEncoding` uses `-` and `_` instead of `+` and `/`.

## Building seek: --json

Text output is for people. Scripts would have to split `seek`'s lines on `:`, which breaks when a
file name or line contains a colon. `--json` prints results as JSON Lines instead:

- one object per match: `{"path":P,"line":N,"text":T}`;
- with `-c`, one object per file: `{"path":P,"count":N}`;
- with `-l`, one object per matching file: `{"path":P}`.

Since [Chapter 10](10-interfaces.md), `main` prints through the `output.Printer` interface. A new
`JSONPrinter` implements it and `main` picks the printer. Nothing else changes.

### internal/search/search.go

Give `Match` struct tags for lower-case JSON keys:

```go
// A Match is one selected line of a file. The struct tags give the
// field names used when a Match is encoded as JSON.
type Match struct {
	Path string `json:"path"` // the file the line came from
	Line int    `json:"line"` // line number, counting from 1
	Text string `json:"text"` // the line itself, without its newline
}
```

Tags change nothing for Go code; they matter only to packages that read them, like `encoding/json`.

### internal/output/json.go (new)

```go
package output

import (
	"encoding/json"
	"io"

	"seek/internal/search"
)

// JSONPrinter prints results as JSON Lines: one JSON object per line.
//
//	{"path":"notes.txt","line":3,"text":"a fox"}  for each match
//	{"path":"notes.txt","count":2}                 with Count
//	{"path":"notes.txt"}                           with FilesOnly
//
// Create one with NewJSONPrinter.
type JSONPrinter struct {
	Count     bool // print only the number of matches per file
	FilesOnly bool // print only the names of files that have matches

	enc *json.Encoder
}

var _ Printer = (*JSONPrinter)(nil)

// NewJSONPrinter returns a JSONPrinter that writes to w.
func NewJSONPrinter(w io.Writer) *JSONPrinter {
	enc := json.NewEncoder(w)
	// By default <, >, and & are escaped (for embedding in HTML). Our
	// output is for programs, so keep the text as it is.
	enc.SetEscapeHTML(false)
	return &JSONPrinter{enc: enc}
}

// fileCount is the JSON form of a per-file count.
type fileCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// PrintFile implements Printer.
func (p *JSONPrinter) PrintFile(path string, matches []search.Match) error {
	path = displayPath(path)
	switch {
	case p.FilesOnly:
		if len(matches) == 0 {
			return nil
		}
		// An anonymous struct is handy for a one-off shape.
		return p.enc.Encode(struct {
			Path string `json:"path"`
		}{path})
	case p.Count:
		return p.enc.Encode(fileCount{Path: path, Count: len(matches)})
	}
	for _, m := range matches {
		m.Path = path // m is a copy, so this doesn't change matches
		if err := p.enc.Encode(m); err != nil {
			return err
		}
	}
	return nil
}
```

- The `Encoder` is created once and kept in an unexported field that outside code can't set, so users
  must call `NewJSONPrinter`, as the doc comment says. The options stay exported fields, like
  `TextPrinter`'s.
- `var _ Printer = (*JSONPrinter)(nil)` makes the compiler check that `*JSONPrinter` implements
  `Printer`, as for `TextPrinter`.
- `displayPath` is `TextPrinter`'s helper in the same package, so standard input shows as
  `(standard input)` in JSON too.
- The `-l` shape is used once, so it's an **anonymous struct**: a `struct {...}` type written inline,
  then a literal `{path}`. Tags work on anonymous structs too.
- In the loop, `m` is a copy of each `Match`, so setting `m.Path` leaves the caller's slice alone.
- `Encode` returns write errors, passed on as `TextPrinter` does.

### cmd/seek/main.go

Choose the printer, replacing the block that creates the `TextPrinter`:

```go
	// p's static type is the Printer interface; its dynamic type is
	// *output.JSONPrinter or *output.TextPrinter.
	var p output.Printer
	if opts.json {
		jp := output.NewJSONPrinter(out)
		jp.Count = opts.count
		jp.FilesOnly = opts.filesOnly
		p = jp
	} else {
		tp := &output.TextPrinter{
			W:           out,
			LineNumbers: opts.lineNumbers,
			Count:       opts.count,
			FilesOnly:   opts.filesOnly,
		}
		if useColor(opts.color, stdout) {
			tp.Highlight = matcher
		}
		p = tp
	}
```

The rest of `run` uses only `p.PrintFile`, so either printer works. JSON output is never colored, and
`-n` doesn't matter since the line number is always included.

### cmd/seek/flags.go

Add the help line before `--stats`:

```go
      --color WHEN          highlight matches: auto, always, or never
                            (default auto, or the value of $SEEK_COLOR)
      --json                print results as JSON Lines, one object per line
      --stats               print statistics to standard error when done
```

the field:

```go
	glob        string    // --glob PATTERN
	json        bool      // --json
	stats       bool      // --stats
```

and the flag:

```go
	boolVar(&opts.hidden, "hidden", "")
	boolVar(&opts.json, "json", "")
	boolVar(&opts.stats, "stats", "")
```

The field `json` shares the package's name, but there's no conflict: `flags.go` doesn't import
`encoding/json`, and the field is always reached through `opts.`.

### Checkpoint

```console
$ go run ./cmd/seek --json fox testdata
{"path":"testdata/poem.txt","line":1,"text":"The quick brown fox"}
{"path":"testdata/poem.txt","line":3,"text":"The dog sleeps; the fox runs."}
{"path":"testdata/src/hello.go","line":4,"text":"\tprintln(\"hello, fox\")"}
$ go run ./cmd/seek --json -c -i fox testdata
{"path":"testdata/notes.md","count":1}
{"path":"testdata/poem.txt","count":3}
{"path":"testdata/src/hello.go","count":1}
$ go run ./cmd/seek --json -l -i fox testdata
{"path":"testdata/notes.md"}
{"path":"testdata/poem.txt"}
{"path":"testdata/src/hello.go"}
$ printf 'a <fox> & a hound\n' | go run ./cmd/seek --json fox -
{"path":"(standard input)","line":1,"text":"a <fox> & a hound"}
$ go run ./cmd/seek --json zebra testdata
exit status 1
```

- The tab and quotes in `hello.go` are escaped as `\t` and `\"`, as JSON strings require; a JSON
  reader turns them back into the original text.
- `SetEscapeHTML(false)` keeps `<fox>` readable. Without that line, the same command prints:

  ```console
  $ printf 'a <fox> & a hound\n' | go run ./cmd/seek --json fox -
  {"path":"(standard input)","line":1,"text":"a \u003cfox\u003e \u0026 a hound"}
  ```

- With no matches, plain `--json` prints nothing. With `-c`, every file gets a line, even with count
  0, as in text output:

  ```console
  $ go run ./cmd/seek --json -c zebra testdata
  {"path":"testdata/notes.md","count":0}
  {"path":"testdata/poem.txt","count":0}
  {"path":"testdata/src/hello.go","count":0}
  exit status 1
  ```

The complete project at this point is in `project/ch15/`.

## Summary

- `json.Marshal(v)` encodes a value; `json.Unmarshal(data, &v)` decodes into a pointer. `go vet`
  catches a missing `&`.
- Only exported fields are encoded. Struct tags such as `json:"name,omitempty"` set the key and
  options; `json:"-"` skips a field. `omitzero` omits zero structs like `time.Time`.
- Decoding matches keys case-insensitively, ignores unknown keys, and leaves missing fields unchanged.
- Decoding into `any` gives `map[string]any`, `[]any`, `string`, `float64`, `bool`, and `nil`. Prefer
  structs when you know the shape. `UseNumber` keeps big numbers exact.
- `json.NewEncoder(w).Encode` writes one value per line (JSON Lines). `json.NewDecoder(r).Decode` reads
  values until `io.EOF`. `SetEscapeHTML(false)`, `SetIndent`, and `DisallowUnknownFields` adjust them.
- Implement `MarshalJSON`/`UnmarshalJSON`, or `MarshalText`/`UnmarshalText`, to control a type's JSON
  form. Unmarshal methods need pointer receivers.
- `json.RawMessage` delays decoding part of the input.
- `time.Time` uses RFC 3339. `[]byte` becomes base64. `encoding/csv` and `encoding/base64` handle the
  other common formats.

## Exercises

1. **Order total.** Decode this JSON array into a `[]Item`; print the item count and total price
   (price × qty). A missing `qty` counts as 0.

   ```json
   [
   	{"name": "pen",    "price": 1.25, "qty": 4},
   	{"name": "paper",  "price": 4.5,  "qty": 1},
   	{"name": "eraser", "price": 0.8}
   ]
   ```

<details><summary>Solution</summary>

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Item struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Qty   int     `json:"qty"`
}

const order = `[
	{"name": "pen",    "price": 1.25, "qty": 4},
	{"name": "paper",  "price": 4.5,  "qty": 1},
	{"name": "eraser", "price": 0.8}
]`

func main() {
	var items []Item
	if err := json.Unmarshal([]byte(order), &items); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	total := 0.0
	for _, it := range items {
		total += it.Price * float64(it.Qty)
	}
	fmt.Printf("%d items, total %.2f\n", len(items), total)
}
```

```console
$ go run .
3 items, total 9.50
```

The eraser has no `qty`, so its `Qty` stays zero. (Real programs should store money as integer
cents, not `float64`.)

</details>

2. **Matching tags.** Write a `User` struct that decodes this JSON, never encodes the password, and
   omits an empty email. Decode it, print the struct, and encode it again.

   ```json
   {"user_name":"ann","is_admin":true,"groups":["dev","ops"],"password":"hunter2"}
   ```

<details><summary>Solution</summary>

```go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name     string   `json:"user_name"`
	Admin    bool     `json:"is_admin"`
	Email    string   `json:"email,omitempty"`
	Groups   []string `json:"groups"`
	Password string   `json:"-"`
}

func main() {
	in := `{"user_name":"ann","is_admin":true,"groups":["dev","ops"],"password":"hunter2"}`
	var u User
	if err := json.Unmarshal([]byte(in), &u); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%+v\n", u)
	out, _ := json.Marshal(u)
	fmt.Println(string(out))
}
```

```console
$ go run .
{Name:ann Admin:true Email: Groups:[dev ops] Password:}
{"user_name":"ann","is_admin":true,"groups":["dev","ops"]}
```

`json:"-"` works both ways, so `Password` isn't even decoded. To read a password but never write it,
you'd need a custom `MarshalJSON`.

</details>

3. **Points as arrays.** Make `Point{X, Y int}` encode as a two-element array `[x,y]` instead of an
   object, and decode from it. Test it inside a `Shape` with a `[]Point`.

<details><summary>Solution</summary>

```go
package main

import (
	"encoding/json"
	"fmt"
)

// A Point is encoded as a two-element array: [x, y].
type Point struct {
	X, Y int
}

func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]int{p.X, p.Y})
}

func (p *Point) UnmarshalJSON(data []byte) error {
	var xy [2]int
	if err := json.Unmarshal(data, &xy); err != nil {
		return fmt.Errorf("point: %w", err)
	}
	p.X, p.Y = xy[0], xy[1]
	return nil
}

type Shape struct {
	Name   string  `json:"name"`
	Points []Point `json:"points"`
}

func main() {
	s := Shape{"line", []Point{{0, 0}, {3, 4}}}
	data, _ := json.Marshal(s)
	fmt.Println(string(data))

	var back Shape
	err := json.Unmarshal([]byte(`{"name":"tri","points":[[0,0],[1,0],[0,1]]}`), &back)
	fmt.Printf("%+v %v\n", back, err)

	err = json.Unmarshal([]byte(`{"points":[{"X":1,"Y":2}]}`), &back)
	fmt.Println("error:", err)
}
```

```console
$ go run .
{"name":"line","points":[[0,0],[3,4]]}
{Name:tri Points:[{X:0 Y:0} {X:1 Y:0} {X:0 Y:1}]} <nil>
error: point: json: cannot unmarshal object into Go value of type [2]int
```

The methods are found inside the slice too, since `encoding/json` checks every value. Marshaling
`[2]int` reuses the package's own array encoding; no hand-built JSON.

</details>

4. **Reading seek's output.** In `seek`, add a command `cmd/jsoncount`. It reads `seek --json` output
   from standard input with a `json.Decoder`, decodes each line into a `search.Match`, and prints
   matches per file, most first. Use `stats.Counter` ([Chapter 13](13-generics.md)). Delete it
   afterwards.

<details><summary>Solution</summary>

```go
// Command jsoncount reads seek --json output and counts matches per file.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"seek/internal/search"
	"seek/internal/stats"
)

func main() {
	var perFile stats.Counter[string]
	dec := json.NewDecoder(os.Stdin)
	for {
		var m search.Match
		err := dec.Decode(&m)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "jsoncount:", err)
			os.Exit(1)
		}
		perFile.Add(m.Path, 1)
	}
	for _, e := range stats.TopN(&perFile, perFile.Len()) {
		fmt.Printf("%s: %d\n", e.Key, e.Count)
	}
}
```

```console
$ go run ./cmd/seek --json -i fox testdata | go run ./cmd/jsoncount
testdata/poem.txt: 3
testdata/notes.md: 1
testdata/src/hello.go: 1
$ echo 'hello' | go run ./cmd/jsoncount
jsoncount: invalid character 'h' looking for beginning of value
exit status 1
```

The struct tags that encode a `Match` also decode it. `cmd/jsoncount` is in the `seek` module, so it
may import the `internal` packages.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-15--encoding-json-and-struct-tags)
