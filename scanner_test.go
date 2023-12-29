package json

import (
	"io"
	"testing"
)

func TestScannerNext(t *testing.T) {
	tests := []struct {
		in     string
		tokens []string
	}{
		{in: `""`, tokens: []string{`""`}},
		{in: `"a"`, tokens: []string{`"a"`}},
		{in: ` "a" `, tokens: []string{`"a"`}},
		{in: `"\""`, tokens: []string{`"\""`}},
		{in: `1`, tokens: []string{`1`}},
		{in: `-1234567.8e+90`, tokens: []string{`-1234567.8e+90`}},
		{in: `{}`, tokens: []string{`{`, `}`}},
		{in: `[]`, tokens: []string{`[`, `]`}},
		{in: `[{}, {}]`, tokens: []string{`[`, `{`, `}`, `,`, `{`, `}`, `]`}},
		{in: `{"a": 0}`, tokens: []string{`{`, `"a"`, `:`, `0`, `}`}},
		{in: `{"a": []}`, tokens: []string{`{`, `"a"`, `:`, `[`, `]`, `}`}},
		{in: `[10]`, tokens: []string{`[`, `10`, `]`}},
		{in: `[{"a": 1,"b": 123.456, "c": null, "d": [1, -2, "three", true, false, ""]}]`,
			tokens: []string{`[`,
				`{`,
				`"a"`, `:`, `1`, `,`,
				`"b"`, `:`, `123.456`, `,`,
				`"c"`, `:`, `null`, `,`,
				`"d"`, `:`, `[`,
				`1`, `,`, `-2`, `,`, `"three"`, `,`, `true`, `,`, `false`, `,`, `""`,
				`]`,
				`}`,
				`]`,
			},
		},
		{in: `{"x": "va\\\\ue", "y": "value y"}`, tokens: []string{
			`{`, `"x"`, `:`, `"va\\\\ue"`, `,`, `"y"`, `:`, `"value y"`, `}`,
		}},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			scanner := NewScanner([]byte(tc.in))
			for n, want := range tc.tokens {
				got := scanner.Next()
				if string(got) != want {
					t.Fatalf("%v: expected: %v, got: %v", n+1, want, string(got))
				}
			}
			last := scanner.Next()
			if len(last) > 0 {
				t.Fatalf("expected: %q, got: %q", "", string(last))
			}
			//if err := scanner.Error(); err != io.EOF {
			//	t.Fatalf("expected: %v, got: %v", io.EOF, err)
			//}
		})
	}
}

func TestParseString(t *testing.T) {
	testParseString(t, `""`, `""`)
	testParseString(t, `"" `, `""`)
	testParseString(t, `"\""`, `"\""`)
	testParseString(t, `"\\\\\\\\\6"`, `"\\\\\\\\\6"`)
	testParseString(t, `"\6"`, `"\6"`)
}

func testParseString(t *testing.T, json, want string) {
	t.Helper()
	scanner := NewScanner([]byte(json))
	got := scanner.Next()
	if string(got) != want {
		t.Fatalf("expected: %q, got: %q", want, got)
	}
}

func TestParseNumber(t *testing.T) {
	testParseNumber(t, `1`)
	// testParseNumber(t, `0000001`)
	testParseNumber(t, `12.0004`)
	testParseNumber(t, `1.7734`)
	testParseNumber(t, `15`)
	testParseNumber(t, `-42`)
	testParseNumber(t, `-1.7734`)
	testParseNumber(t, `1.0e+28`)
	testParseNumber(t, `-1.0e+28`)
	testParseNumber(t, `1.0e-28`)
	testParseNumber(t, `-1.0e-28`)
	testParseNumber(t, `-18.3872`)
	testParseNumber(t, `-2.1`)
	testParseNumber(t, `-1234567.891011121314`)
}

func testParseNumber(t *testing.T, tc string) {
	t.Helper()
	scanner := NewScanner([]byte(tc))
	got := scanner.Next()
	if string(got) != tc {
		t.Fatalf("expected: %q, got: %q", tc, got)
	}
}

func BenchmarkParseNumber(b *testing.B) {
	tests := []string{
		`1`,
		`12.0004`,
		`1.7734`,
		`15`,
		`-42`,
		`-1.7734`,
		`1.0e+28`,
		`-1.0e+28`,
		`1.0e-28`,
		`-1.0e-28`,
		`-18.3872`,
		`-2.1`,
		`-1234567.891011121314`,
	}

	for _, tc := range tests {
		b.Run(tc, func(b *testing.B) {
			data := []byte(tc)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				scanner := &Scanner{
					data: []byte(tc),
				}
				n := scanner.parseNumber(scanner.data[scanner.offset])
				if n != len(tc) {
					b.Fatalf("expected: %v, got: %v", len(tc), n)
				}
			}
		})
	}
}

func TestScanner(t *testing.T) {
	testScanner(t, 1)
	testScanner(t, 8)
	testScanner(t, 64)
	testScanner(t, 256)
	testScanner(t, 1<<10)
	testScanner(t, 8<<10)
	testScanner(t, 1<<20)
}

func testScanner(t *testing.T, sz int) {
	t.Helper()
	for _, tc := range inputs {
		r := fixture(t, tc.path)
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read fixture: %v", err)
		}
		t.Run(tc.path, func(t *testing.T) {
			sc := &Scanner{data: data}
			n := 0
			for len(sc.Next()) > 0 {
				n++
			}
			if n != tc.alltokens {
				t.Fatalf("expected %v tokens, got %v", tc.alltokens, n)
			}
		})
	}
}

func BenchmarkScanner_skipArray(b *testing.B) {
	input := []byte(`[{"some": "value", "props": [1, 2, 3]}, {"some": "value2", "props": [1, 2, 3]}, {"some": "value3", "props": [1, 2, 3]}]
		"c": [1, 2, true]
	}`)
	s := Scanner{
		offset: 1,
		data:   input,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.offset = 1
		s.skipArray()
	}
}
