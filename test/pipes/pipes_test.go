package pipes_test

import (
	"fmt"
	"testing"

	"github.com/antonmedv/expr"
	"github.com/stretchr/testify/require"
)

func TestPipes(t *testing.T) {
	env := map[string]interface{}{
		"sprintf": fmt.Sprintf,
	}

	tests := []struct {
		input string
		want  interface{}
	}{
		{
			`-1 | abs()`,
			1,
		},
		{
			`"%s bar %d" | sprintf("foo", -42 | abs())`,
			"foo bar 42",
		},
		{
			`[] | first() ?? "foo"`,
			"foo",
		},
		{
			`"a" | upper() + "B" | lower()`,
			"ab",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			program, err := expr.Compile(test.input, expr.Env(env), expr.ExperimentalPipes())
			require.NoError(t, err)

			out, err := expr.Run(program, env)
			require.NoError(t, err)
			require.Equal(t, test.want, out)
		})
	}
}

func TestPipes_map_filter(t *testing.T) {
	program, err := expr.Compile(`1..9 | map(# + 1) | filter(# % 2 == 0)`, expr.ExperimentalPipes())
	require.NoError(t, err)

	out, err := expr.Run(program, nil)
	require.NoError(t, err)
	require.Equal(t, []interface{}{2, 4, 6, 8, 10}, out)
}
