// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of opt library.
//
// opt is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// opt is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with opt. If not, see <http://www.gnu.org/licenses/>.

// Package opt implements command line options parsing.
//
// Short and long option formats are supported: -o and --option. Multiple
// short options can be grouped with one dash sign: -abc. Option can has
// non-optional argument. Short option parameter can be passed as a part
// of option string or with a next coming string. To pass value `foo` for
// option `o` next two syntaxes can be used: -ofoo, -o foo. Long option
// can be separated from its value with space or equals sign: --option foo,
// --option=foo.
// Command-line options and arguments can be passed in any order.
// `--` indicates and of the passed options, and rest is treated as
// arguments only, even if it is started from dash.
//
// Usage:
//
// 1. Define supported options list using Desc structure to describe
// every option.
// 	descs := []*opt.Desc{
// 		{"a", "add", opt.ArgNone,
// 			"", "add new item"},
// 		{"d", "delete", opt.ArgNone,
// 			"", "delete item"},
// 		{"h", "help", opt.ArgNone,
// 			"", "display help information and exit"},
// 		{"p", "path", opt.ArgString,
// 			"path", "path to store output files to"},
// 	}
//
// 2. Parse command-line arguments.
// 	opts, args, err := opt.Parse(os.Args[1:], descs)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "%s\n", err)
// 		os.Exit(1)
// 	}
//
// 3. Use options and arguments.
// 	if opts.Bool("help") {
// 		fmt.Println("Options:")
// 		fmt.Print(opt.Usage(descs))
// 	}
//
// 	path := opts.StringOr("path", "")
//
// 	if opts.Bool("add") {
// 		fmt.Printf("Adding new item into '%s'...\n", path)
// 	}
// 	if opts.Bool("delete") {
// 		fmt.Printf("Deleting new item from '%s'...\n", path)
// 	}
//
// 	fmt.Printf("arguments: %s\n", args)
package opt

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var errInvalidOptFormat = errors.New("invalid option format")

// Argument type type.
type ArgType int

const (
	ArgNone ArgType = iota
	ArgFloat
	ArgInt
	ArgString
)

// Desc describes available option.
// Short or Long option can be empty, in case option doen't have
// short or long version respectively.
type Desc struct {
	// Short, one letter, name of the option. Short options
	// starts with single dash sign.
	Short string
	// Long option name. Long options starts with two dash signs.
	Long string
	// Arg is an option type description.
	Arg ArgType
	// Name of the argument for string, int and float options.
	// This name is used for usage information generation.
	ArgName string
	// Option's description. Used for usage information generation.
	Description string
}

type descSlice []*Desc

func (d descSlice) Len() int {
	return len(d)
}

func (d descSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d descSlice) Less(i, j int) bool {
	return d[i].Short+d[i].Long < d[j].Short+d[j].Long
}

// Option parsed from command line arguments.
type Option struct {
	// Option description.
	Desc *Desc
	// Arguments passed with command line arguments.
	Args []interface{}
}

// Options is a list of all parsed options.
type Options []*Option

// Bool returns true if option named `name` was passed with args.
func (o Options) Bool(name string) bool {
	return o.option(name) != nil
}

// Int returns integer option's argument by its short or long name.
// If option was not defined second parameter is false.
func (o Options) Int(name string) (int, bool) {
	v := o.arg(name)
	if v == nil {
		return 0, false
	}
	if _, ok := v.(int); !ok {
		panic("not an integer option")
	}

	return v.(int), true
}

// Ints returns a list of arguments for int option by its short or long name.
func (o Options) Ints(name string) []int {
	v := o.args(name)
	s := make([]int, len(v))

	for i, vv := range v {
		if ii, ok := vv.(int); !ok {
			panic("not an integer option")
		} else {
			s[i] = ii
		}
	}

	return s
}

// IntOr returns argument of int option or default value if option was not
// defined by command line arguments list.
func (o Options) IntOr(name string, value int) int {
	i, ok := o.Int(name)
	if !ok {
		return value
	} else {
		return i
	}
}

// Float returns float64 option's argument by its short or long name.
// If option was not defined second parameter is false.
func (o Options) Float(name string) (float64, bool) {
	v := o.arg(name)
	if v == nil {
		return 0, false
	}
	if _, ok := v.(float64); !ok {
		panic("not a float64 option")
	}

	return v.(float64), true
}

// Floats returns a list of arguments for float64 option
// by its short or long name.
func (o Options) Floats(name string) []float64 {
	v := o.args(name)
	s := make([]float64, len(v))

	for i, vv := range v {
		if ii, ok := vv.(float64); !ok {
			panic("not a float option")
		} else {
			s[i] = ii
		}
	}

	return s
}

// FloatOr returns argument of float option or default value
// if option was not defined by command line arguments list.
func (o Options) FloatOr(name string, value float64) float64 {
	i, ok := o.Float(name)
	if !ok {
		return value
	} else {
		return i
	}
}

// String returns string option's argument by its short or long name.
// If option was not defined second parameter is false.
func (o Options) String(name string) (string, bool) {
	v := o.arg(name)
	if v == nil {
		return "", false
	}
	if _, ok := v.(string); !ok {
		panic("not a string option")
	}

	return v.(string), true
}

// Strings returns a list of arguments for string option
// by its short or long name.
func (o Options) Strings(name string) []string {
	v := o.args(name)
	s := make([]string, len(v))

	for i, vv := range v {
		if ii, ok := vv.(string); !ok {
			panic("not a string option")
		} else {
			s[i] = ii
		}
	}

	return s
}

// StringOr returns argument of string option or default value
// if option was not defined by command line arguments list.
func (o Options) StringOr(name string, value string) string {
	i, ok := o.String(name)
	if !ok {
		return value
	} else {
		return i
	}
}

// option returns option by short or long name.
func (o Options) option(name string) *Option {
	for _, op := range o {
		if op.Desc.Short == name || op.Desc.Long == name {
			return op
		}
	}

	return nil
}

// args returns Args by option short or long name.
func (o Options) args(name string) []interface{} {
	op := o.option(name)

	if op == nil {
		return nil
	}

	return op.Args
}

// arg returns last argument for option with given name.
func (o Options) arg(name string) interface{} {
	args := o.args(name)

	if len(args) == 0 {
		return nil
	} else {
		return args[len(args)-1]
	}
}

// Usage returns option description lines ready to be printed to stdout.
//
// Return string example:
//   -a, --add     add new item
//   -d, --delete  delete item
func Usage(descs []*Desc) string {
	t := make([]*Desc, len(descs))
	copy(t, descs)
	descs = t
	sort.Sort(descSlice(descs))

	var lines []string
	var max int

	for _, d := range descs {
		s := "  "

		if d.Short != "" {
			s += "-" + d.Short
		} else {
			s += "  "
		}
		if d.Long != "" {
			if d.Short != "" {
				s += ", "
			} else {
				s += "  "
			}

			s += "--" + d.Long
		}
		if d.Arg != ArgNone {
			s += " <" + d.ArgName + ">"
		}

		lines = append(lines, s)

		if len(s) > max {
			max = len(s)
		}
	}

	var b bytes.Buffer

	for i, d := range descs {
		l := lines[i]

		b.WriteString(l)
		b.WriteString(strings.Repeat(" ", max-len(l)+2))
		b.WriteString(d.Description)
		b.WriteString("\n")
	}

	return b.String()
}

// Parse parses given command line arguments. Available application arguments
// are defined by `descs` argument.
// Returns a list of parsed options and a list of free arguments.
func Parse(args []string, descs []*Desc) (Options, []string, error) {
	var opts Options
	var params []string

	for i := 0; i < len(args); {
		arg := args[i]

		// Flags end indicator -- rest are treated as arguments.
		if arg == "--" {
			params = append(params, args[i+1:]...)
			i += len(args) - i + 1
		} else if strings.HasPrefix(arg, "-") {
			o, n, err := parseDashed(descs, args[i:])
			if err != nil {
				return nil, nil, err
			}
			opts = append(opts, o...)
			i += n
		} else {
			params = append(params, arg)
			i++
		}
	}

	return join(opts), params, nil
}

// parseDashed parses one command line argument which starts with `-`.
// As a result more that one Option can be parsed, for instance in next case:
// `-abc`. In such case three options will be produced: `a`, `b` and `c`.
func parseDashed(descs []*Desc, args []string) ([]*Option, int, error) {
	var opts []*Option
	a := args[0]
	n := 1

	switch dashesNum(a) {
	case 0:
		// Shoud not happen.
		return nil, 0, errors.New("not an option")
	case 1:
		// Parse short option.
		runes := strings.Split(a[1:], "")

		if len(runes) == 0 {
			return nil, 0, errInvalidOptFormat
		}
		for i := 0; i < len(runes); i++ {
			r := runes[i]

			// for i, r := range runes {
			d := findDesc(descs, string(r))
			if d == nil {
				return nil, 0, fmt.Errorf("unrecognized option '%s'", r)
			}
			o := &Option{Desc: d, Args: nil}
			var a string

			if d.Arg != ArgNone {
				// If short parameter is last next command-line
				// argument is a option's argument. Otherwise,
				// rest of this argument is an option's argument.
				if i == len(runes)-1 {
					if len(args) < 2 {
						return nil, 0, fmt.Errorf("option '%s' requires an argument", r)
					}
					a = args[1]
					n++
				} else {
					a = strings.Join(runes[i+1:], "")
				}
				v, err := parseArg(string(r), d.Arg, a)
				if err != nil {
					return nil, 0, err
				}
				o.Args = []interface{}{v}
				opts = append(opts, o)

				break
			} else {
				opts = append(opts, o)
			}
		}
	case 2:
		// Parse long option.
		var vals []interface{}
		parts := strings.SplitN(a[2:], "=", 2)
		name := parts[0]
		d := findDesc(descs, name)

		if d == nil {
			return nil, 0, fmt.Errorf("unrecognized option '%s'", name)
		}
		if d.Arg != ArgNone {
			if d == nil {
				return nil, 0, fmt.Errorf("unrecognized option '%s'", name)
			}

			var val string
			if len(parts) > 1 {
				val = parts[1]
			} else {
				if len(args) < 2 {
					return nil, 0, fmt.Errorf("option '%s' requires an argument", name)
				}
				val = args[1]
				n++
			}

			v, err := parseArg(name, d.Arg, val)
			if err != nil {
				return nil, 0, err
			}
			vals = []interface{}{v}
		} else if len(parts) > 1 {
			return nil, 0, fmt.Errorf("option '%s' doesn't allow an argument", name)
		} else {
			name = a[2:]
		}

		opts = append(opts, &Option{Desc: d, Args: vals})
	default:
		// More than one dashes is invalid format.
		return nil, 0, errInvalidOptFormat
	}

	return opts, n, nil
}

// parseArg converts raw argument string to the type defined
// by option's descriptor.
func parseArg(name string, tp ArgType, value string) (interface{}, error) {
	switch tp {
	case ArgNone:
		return nil, nil
	case ArgFloat:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid argument for option '%s'", name)
		}
		return f, nil
	case ArgInt:
		i, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid argument for option '%s'", name)
		}
		return i, nil
	case ArgString:
		return value, nil
	default:
		return nil, errors.New("argument not allowed")
	}
}

// join joins two or more options with the same name.
func join(opts []*Option) []*Option {
	var uopts []*Option

	for _, o := range opts {
		var u *Option

		for _, uu := range uopts {
			if uu.Desc == o.Desc {
				u = uu
			}
		}
		if u != nil {
			if u.Desc.Arg != ArgNone {
				u.Args = append(u.Args, o.Args...)
			}
		} else {
			uopts = append(uopts, o)
		}
	}

	return uopts
}

// findDesc searches Desc by options' short or long name.
func findDesc(descs []*Desc, name string) *Desc {
	for _, d := range descs {
		if d.Short == name || d.Long == name {
			return d
		}
	}

	return nil
}

// dashesNum returns number of `-` at the beginning
// of the given string.
func dashesNum(s string) int {
	n := 0

	for _, r := range s {
		if r == '-' {
			n++
		} else {
			break
		}
	}

	return n
}
