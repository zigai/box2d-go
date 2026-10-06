package bench

import (
	"errors"
	"strconv"
	"testing"
)

const defaultBenchmarkSteps = 300

var errInvalidBenchmarkSteps = errors.New("BENCH_STEPS must be an integer in 1..2147483647")

func parseBenchmarkSteps(value string) (int, error) {
	if value == "" {
		return defaultBenchmarkSteps, nil
	}

	steps, err := strconv.ParseUint(value, 10, 31)
	if err != nil || steps == 0 {
		return 0, errInvalidBenchmarkSteps
	}

	return int(steps), nil
}

func TestBenchmarkSteps(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  int
	}{
		{name: "default", want: defaultBenchmarkSteps},
		{name: "single", value: "1", want: 1},
		{name: "custom", value: "500", want: 500},
		{name: "leading-zeros", value: "00500", want: 500},
		{name: "maximum", value: "2147483647", want: 2147483647},
		{name: "zero", value: "0"},
		{name: "padded-zero", value: "000"},
		{name: "negative", value: "-1"},
		{name: "plus-sign", value: "+1"},
		{name: "fractional", value: "1.5"},
		{name: "whitespace", value: " 10"},
		{name: "text", value: "ten"},
		{name: "overflow", value: "2147483648"},
		{name: "huge", value: "999999999999999999999999"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(test *testing.T) {
			actual, err := parseBenchmarkSteps(testCase.value)
			if actual != testCase.want {
				test.Fatalf("got %d steps, want %d", actual, testCase.want)
			}

			if (err == nil) != (testCase.want > 0) {
				test.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
