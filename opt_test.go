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

package opt

import (
	"reflect"
	"testing"
)

func TestDoubleDash(t *testing.T) {
	testArgs := []string{"-a", "A", "--", "-b", "C"}
	descs := []*Desc{
		{"a", "", ArgNone, "", ""},
	}
	expOpts := []*Option{{descs[0], nil}}
	expArgs := []string{"A", "-b", "C"}

	opts, args, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}
	assertOptions(t, expOpts, opts)
	assertArgs(t, expArgs, args)
}

func TestShort(t *testing.T) {
	testArgs := []string{"-a", "-b", "-cd", "-e"}
	descs := []*Desc{
		{"a", "", ArgNone, "", ""},
		{"b", "", ArgNone, "", ""},
		{"c", "", ArgNone, "", ""},
		{"d", "", ArgNone, "", ""},
		{"e", "", ArgNone, "", ""},
	}
	expected := []*Option{
		{descs[0], nil},
		{descs[1], nil},
		{descs[2], nil},
		{descs[3], nil},
		{descs[4], nil},
	}

	opts, args, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}
	assertOptions(t, expected, opts)
	assertArgs(t, nil, args)
}

func TestShortArgument(t *testing.T) {
	testArgs := []string{"-a", "A", "-a", "AA", "-c", "C", "-bc", "CC"}
	descs := []*Desc{
		{"a", "", ArgString, "", ""},
		{"b", "", ArgNone, "", ""},
		{"c", "", ArgString, "", ""},
	}
	expected := []*Option{
		{descs[0], []interface{}{"A", "AA"}},
		{descs[1], nil},
		{descs[2], []interface{}{"C", "CC"}},
	}

	opts, args, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}
	assertOptions(t, expected, opts)
	assertArgs(t, nil, args)
}

func TestLong(t *testing.T) {
	testArgs := []string{"--a-opt", "--b-opt", "--c-opt"}
	descs := []*Desc{
		{"", "a-opt", ArgNone, "", ""},
		{"", "b-opt", ArgNone, "", ""},
		{"", "c-opt", ArgNone, "", ""},
	}
	expected := []*Option{
		{descs[0], nil},
		{descs[1], nil},
		{descs[2], nil},
	}

	opts, args, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}
	assertOptions(t, expected, opts)
	assertArgs(t, nil, args)
}

func TestLongArgument(t *testing.T) {
	testArgs := []string{"--a-opt", "A", "--b-opt", "B", "--b-opt=BB"}
	descs := []*Desc{
		{"", "a-opt", ArgString, "", ""},
		{"", "b-opt", ArgString, "", ""},
	}
	expected := []*Option{
		{descs[0], []interface{}{"A"}},
		{descs[1], []interface{}{"B", "BB"}},
	}

	opts, args, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}
	assertOptions(t, expected, opts)
	assertArgs(t, nil, args)
}

func TestBool(t *testing.T) {
	testArgs := []string{"-ab", "-c"}
	descs := []*Desc{
		{"a", "", ArgNone, "", ""},
		{"b", "", ArgNone, "", ""},
		{"c", "", ArgNone, "", ""},
	}

	opts, _, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}

	if !opts.Bool("a") {
		t.Fatal()
	}
	if !opts.Bool("b") {
		t.Fatal()
	}
	if !opts.Bool("c") {
		t.Fatal()
	}
	if opts.Bool("d") {
		t.Fatal()
	}
}

func TestString(t *testing.T) {
	testArgs := []string{"--a-opt", "A", "-bB"}
	descs := []*Desc{
		{"a", "a-opt", ArgString, "", ""},
		{"b", "b-opt", ArgString, "", ""},
	}

	opts, _, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}

	a, ok := opts.String("a")
	if a != "A" || !ok {
		t.Fatal()
	}
	a, ok = opts.String("a-opt")
	if a != "A" || !ok {
		t.Fatal()
	}

	b, ok := opts.String("b")
	if b != "B" || !ok {
		t.Fatal()
	}
	b, ok = opts.String("b-opt")
	if b != "B" || !ok {
		t.Fatal()
	}
}

func TestTypeInt(t *testing.T) {
	testArgs := []string{"-a", "1", "--aaa", "2", "-b", "3"}
	descs := []*Desc{
		{"a", "aaa", ArgInt, "", ""},
		{"b", "bbb", ArgInt, "", ""},
	}
	expected := []*Option{
		{descs[0], []interface{}{1, 2}},
		{descs[1], []interface{}{3}},
	}

	opts, args, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}

	assertOptions(t, expected, opts)
	assertArgs(t, nil, args)

	i, ok := opts.Int("a")
	if i != 2 || !ok {
		t.Fatal()
	}
	is := opts.Ints("a")
	if len(is) != 2 || is[0] != 1 || is[1] != 2 {
		t.Fatal()
	}
	i = opts.IntOr("a", -1)
	if i != 2 {
		t.Fatal()
	}
	i = opts.IntOr("c", -1)
	if i != -1 {
		t.Fatal()
	}
}

func TestTypeFloat(t *testing.T) {
	testArgs := []string{"-a", "1.23", "--aaa", "2.34", "-b", "3"}
	descs := []*Desc{
		{"a", "aaa", ArgFloat, "", ""},
		{"b", "bbb", ArgFloat, "", ""},
	}
	expected := []*Option{
		{descs[0], []interface{}{1.23, 2.34}},
		{descs[1], []interface{}{3.0}},
	}

	opts, args, err := Parse(testArgs, descs)
	if err != nil {
		t.Fatal(err)
	}

	assertOptions(t, expected, opts)
	assertArgs(t, nil, args)

	f, ok := opts.Float("a")
	if f != 2.34 || !ok {
		t.Fatal()
	}
	fs := opts.Floats("a")
	if len(fs) != 2 || fs[0] != 1.23 || fs[1] != 2.34 {
		t.Fatal()
	}
	f = opts.FloatOr("a", 0.12)
	if f != 2.34 {
		t.Fatal()
	}
	f = opts.FloatOr("c", 0.12)
	if f != 0.12 {
		t.Fatal()
	}
}

func TestUsage(t *testing.T) {
	expected := `  -a, --add          add new item
      --delete       delete item
  -h                 display help information and exit
  -p, --path <path>  path to store output files to
`

	descs := []*Desc{
		{"", "delete", ArgNone,
			"", "delete item"},
		{"a", "add", ArgNone,
			"", "add new item"},
		{"p", "path", ArgString,
			"path", "path to store output files to"},
		{"h", "", ArgNone,
			"", "display help information and exit"},
	}

	if Usage(descs) != expected {
		t.Fatal()
	}
}

func assertOptions(t *testing.T, expected, actual []*Option) {
	if len(expected) != len(actual) {
		t.Fatalf("%d options expected but %d found",
			len(expected), len(actual))
	}

exp:
	for _, e := range expected {
		var a *Option
		for _, a = range actual {
			if reflect.DeepEqual(e, a) {
				continue exp
			}
		}

		t.Fatalf("%s option expected but %s found", e, a)
	}
}

func assertArgs(t *testing.T, expected, actual []string) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("%v arguments expected but %v found", expected, actual)
	}
}
