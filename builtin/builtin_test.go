package builtin_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/checker"
	"github.com/antonmedv/expr/conf"
	"github.com/antonmedv/expr/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltin(t *testing.T) {
	var tests = []struct {
		input string
		want  interface{}
	}{
		{`len(1..10)`, 10},
		{`len({foo: 1, bar: 2})`, 2},
		{`len("hello")`, 5},
		{`abs(-5)`, 5},
		{`abs(.5)`, .5},
		{`abs(-.5)`, .5},
		{`int(5.5)`, 5},
		{`int(5)`, 5},
		{`int("5")`, 5},
		{`float(5)`, 5.0},
		{`float(5.5)`, 5.5},
		{`float("5.5")`, 5.5},
		{`string(5)`, "5"},
		{`string(5.5)`, "5.5"},
		{`string("5.5")`, "5.5"},
		{`trim("  foo  ")`, "foo"},
		{`trim("__foo___", "_")`, "foo"},
		{`trimPrefix("prefix_foo", "prefix_")`, "foo"},
		{`trimSuffix("foo_suffix", "_suffix")`, "foo"},
		{`upper("foo")`, "FOO"},
		{`lower("FOO")`, "foo"},
		{`split("foo,bar,baz", ",")`, []string{"foo", "bar", "baz"}},
		{`splitN("foo,bar,baz", ",", 2)`, []string{"foo", "bar,baz"}},
		{`splitAfter("foo,bar,baz", ",")`, []string{"foo,", "bar,", "baz"}},
		{`splitAfterN("foo,bar,baz", ",", 2)`, []string{"foo,", "bar,baz"}},
		{`replace("foo,bar,baz", ",", ";")`, "foo;bar;baz"},
		{`replace("foo,bar,baz,goo", ",", ";", 2)`, "foo;bar;baz,goo"},
		{`repeat("foo", 3)`, "foofoofoo"},
		{`join(ArrayOfString, ",")`, "foo,bar,baz"},
		{`join(ArrayOfString)`, "foobarbaz"},
		{`join(["foo", "bar", "baz"], ",")`, "foo,bar,baz"},
		{`join(["foo", "bar", "baz"])`, "foobarbaz"},
		{`indexOf("foo,bar,baz", ",")`, 3},
		{`lastIndexOf("foo,bar,baz", ",")`, 7},
		{`hasPrefix("foo,bar,baz", "foo")`, true},
		{`hasSuffix("foo,bar,baz", "baz")`, true},
		{`max(1, 2, 3)`, 3},
		{`max(1.5, 2.5, 3.5)`, 3.5},
		{`min(1, 2, 3)`, 1},
		{`min(1.5, 2.5, 3.5)`, 1.5},
		{`toJSON({foo: 1, bar: 2})`, "{\n  \"bar\": 2,\n  \"foo\": 1\n}"},
		{`fromJSON("[1, 2, 3]")`, []interface{}{1.0, 2.0, 3.0}},
		{`toBase64("hello")`, "aGVsbG8="},
		{`fromBase64("aGVsbG8=")`, "hello"},
		{`now().Format("2006-01-02T15:04:05Z")`, time.Now().Format("2006-01-02T15:04:05Z")},
		{`duration("1h")`, time.Hour},
		{`date("2006-01-02T15:04:05Z")`, time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
		{`date("2006.01.02", "2006.01.02")`, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)},
		{`first(ArrayOfString)`, "foo"},
		{`first(ArrayOfInt)`, 1},
		{`first(ArrayOfAny)`, 1},
		{`first([])`, nil},
		{`last(ArrayOfString)`, "baz"},
		{`last(ArrayOfInt)`, 3},
		{`last(ArrayOfAny)`, true},
		{`last([])`, nil},
		{`get(ArrayOfString, 1)`, "bar"},
		{`get(ArrayOfString, 99)`, nil},
		{`get(ArrayOfInt, 1)`, 2},
		{`get(ArrayOfInt, -1)`, 3},
		{`get(ArrayOfAny, 1)`, "2"},
		{`get({foo: 1, bar: 2}, "foo")`, 1},
		{`get({foo: 1, bar: 2}, "unknown")`, nil},
	}

	env := map[string]interface{}{
		"ArrayOfString": []string{"foo", "bar", "baz"},
		"ArrayOfInt":    []int{1, 2, 3},
		"ArrayOfAny":    []interface{}{1, "2", true},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			program, err := expr.Compile(test.input, expr.Env(env))
			require.NoError(t, err)

			out, err := expr.Run(program, env)
			require.NoError(t, err)
			assert.Equal(t, test.want, out)
		})
	}
}

func TestBuiltin_errors(t *testing.T) {
	var errorTests = []struct {
		input string
		err   string
	}{
		{`len()`, `invalid number of arguments (expected 1, got 0)`},
		{`len(1)`, `invalid argument for len (type int)`},
		{`abs()`, `invalid number of arguments (expected 1, got 0)`},
		{`abs(1, 2)`, `invalid number of arguments (expected 1, got 2)`},
		{`abs("foo")`, `invalid argument for abs (type string)`},
		{`int()`, `invalid number of arguments (expected 1, got 0)`},
		{`int(1, 2)`, `invalid number of arguments (expected 1, got 2)`},
		{`float()`, `invalid number of arguments (expected 1, got 0)`},
		{`float(1, 2)`, `invalid number of arguments (expected 1, got 2)`},
		{`string(1, 2)`, `too many arguments to call string`},
		{`trim()`, `not enough arguments to call trim`},
		{`max()`, `not enough arguments to call max`},
		{`max(1, "2")`, `invalid argument for max (type string)`},
		{`min()`, `not enough arguments to call min`},
		{`min(1, "2")`, `invalid argument for min (type string)`},
		{`duration("error")`, `invalid duration`},
		{`date("error")`, `invalid date`},
		{`get()`, `invalid number of arguments (expected 2, got 0)`},
		{`get(1, 2)`, `type int does not support indexing`},
	}
	for _, test := range errorTests {
		t.Run(test.input, func(t *testing.T) {
			_, err := expr.Eval(test.input, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), test.err)
		})
	}
}

func TestBuiltin_types(t *testing.T) {
	env := map[string]interface{}{
		"num":           42,
		"str":           "foo",
		"ArrayOfString": []string{"foo", "bar", "baz"},
		"ArrayOfInt":    []int{1, 2, 3},
	}

	tests := []struct {
		input string
		want  reflect.Kind
	}{
		{`get(ArrayOfString, 0)`, reflect.String},
		{`get(ArrayOfInt, 0)`, reflect.Int},
		{`first(ArrayOfString)`, reflect.String},
		{`first(ArrayOfInt)`, reflect.Int},
		{`last(ArrayOfString)`, reflect.String},
		{`last(ArrayOfInt)`, reflect.Int},
		{`get($env, 'str')`, reflect.String},
		{`get($env, 'num')`, reflect.Int},
		{`get($env, 'ArrayOfString')`, reflect.Slice},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			tree, err := parser.Parse(test.input)
			require.NoError(t, err)

			rtype, err := checker.Check(tree, conf.New(env))
			require.NoError(t, err)
			require.True(t, rtype.Kind() == test.want, fmt.Sprintf("expected %s, got %s", test.want, rtype.Kind()))
		})
	}
}
