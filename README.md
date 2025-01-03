### Description
Package opt implements command line options parsing, which can be used
instead of flag package from Go's standard library. It is inspired by
GNU getopt package, and try to be close to it (but optional arguments
not supported for now).

Short and long option formats are supported: `-o` and `--option`. Multiple
short options can be grouped with one dash sign: `-abc`. Option can has
non-optional argument. Short option parameter can be passed as a part
of option string or with a next coming string. To pass value `foo` for
option `o` next two syntaxes can be used: `-ofoo`, `-o foo`. Long option
can be separated from its value with space or equals sign: `--option foo`,
`--option=foo`.
Command-line options and arguments can be passed in any order.
`--` indicates and of the passed options, and rest is treated as
arguments only, even if it is started from dash.

### Installation
opt is available using the standard `go get` command.

Install by running:

    go get github.com/vchimishuk/opt

Run tests by running:

    go test github.com/vchimishuk/opt

### Usage

Define supported options list using Desc structure to describe
every option.
```go
descs := []*opt.Desc{
	{"a", "add", opt.ArgNone,
		"", "add new item"},
	{"d", "delete", opt.ArgNone,
		"", "delete item"},
	{"h", "help", opt.ArgNone,
		"", "display help information and exit"},
	{"p", "path", opt.ArgString,
		"path", "path to store output files to"},
}
```

Parse command-line arguments.
```go
opts, args, err := opt.Parse(os.Args[1:], descs)
if err != nil {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
```

Use options and arguments.
```go
if opts.Has("help") {
	fmt.Println("Options:")
	fmt.Print(opt.Usage(descs))
}

path := "default"
if opts.Has("path") {
	path = opts.String("path")
}
// Alternative way is to use the next line.
// path := opts.StringOr("path", "default")

if opts.Has("add") {
	fmt.Printf("Adding new item into '%s'...\n", path)
}
if opts.Has("delete") {
	fmt.Printf("Deleting new item from '%s'...\n", path)
}

fmt.Printf("arguments: %s\n", args)
```
