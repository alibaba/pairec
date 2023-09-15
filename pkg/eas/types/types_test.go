package types

import "testing"

type rangeTestCase struct {
	input       string
	expectErr   bool
	expectRange Range
}

var rangeTestCases = []rangeTestCase{
	{
		input:     "(20, 400)",
		expectErr: false,
		expectRange: Range{
			LeftInclude:  false,
			RightInclude: false,
			PositiveInf:  false,
			Begin:        20,
			End:          400,
		},
	},
	{
		input:     "(20, 400]",
		expectErr: false,
		expectRange: Range{
			LeftInclude:  false,
			RightInclude: true,
			PositiveInf:  false,
			Begin:        20,
			End:          400,
		},
	},
	{
		input:     "(20, +inf]",
		expectErr: true,
	},
	{
		input:     "(20, +inf)",
		expectErr: false,
		expectRange: Range{
			LeftInclude:  false,
			RightInclude: false,
			PositiveInf:  true,
			Begin:        20,
		},
	},
}

func Test_ParseRange(t *testing.T) {
	for _, tc := range rangeTestCases {
		t.Logf("input %s", tc.input)
		r, err := ParseRange(tc.input)
		if tc.expectErr && err != nil {
			t.Logf("expected error: %v", err)
			continue
		}
		if tc.expectErr && err == nil {
			t.Fatal("should has error but get nil")
		} else if !tc.expectErr && err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if r != tc.expectRange {
				t.Fatalf("got range %+v, but got %+v", tc.expectRange, r)
			}
		}
	}
}

func TestTags(t *testing.T) {
	t1 := Tags{"a": "b", "c": "d"}
	if !t1.Contains(Tags{"a": "b", "c": "d"}) {
		t.Fatal("should contains")
	}
	if !t1.Contains(Tags{"a": "b"}) {
		t.Fatal("should contains")
	}
	if !t1.Contains(Tags{}) {
		t.Fatal("should contains")
	}
}
