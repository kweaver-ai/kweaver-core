package dm

import (
	"testing"
	"bytes"
)

func TestDialector_QuoteTo(t *testing.T) {
	testdatas := []struct {
		raw    string
		expect string
	}{
		{"datadase.tableUser", "\"datadase\".\"tableUser\""},
		{"datadase.table`User", "\"datadase\".\"table`User\""},
		{"`a`.`b`", "\"`a`\".\"`b`\""},
		{"`a`.b`", "\"`a`\".\"b`\""},
		{"a.`b`", "\"a\".\"`b`\""},
		{"`a`.b`c", "\"`a`\".\"b`c\""},
		{"`a`.`b`c`", "\"`a`\".\"`b`c`\""},
		{"`a`.b", "\"`a`\".\"b\""},
		{"`ab`", "\"`ab`\""},
		{"`a``b`", "\"`a``b`\""},
		{"`a```b`", "\"`a```b`\""},
		{"a`b", "\"a`b\""},
		{"ab", "\"ab\""},
		{"`a.b`", "\"`a\".\"b`\""},
		{"a.b", "\"a\".\"b\""},
	}

	dailor := Open("")
	for _, item := range testdatas {
		buf := &bytes.Buffer{}
		dailor.QuoteTo(buf, item.raw)
		if buf.String() != item.expect {
			t.Fatalf("quote %q fail, got %q, expect %q", item.raw, buf.String(), item.expect)
		}
	}
}

// BenchmarkDialector_QuoteTo
// Result:
// goos: darwin
// goarch: amd64
// pkg: gorm.io/driver/mysql
// cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
// BenchmarkDialector_QuoteTo               9184232               113.2 ns/op
// BenchmarkDialector_QuoteTo-2             9782818               112.3 ns/op
// BenchmarkDialector_QuoteTo-4            10726722               109.0 ns/op
// BenchmarkDialector_QuoteTo-8             9656778               113.1 ns/op
// BenchmarkDialector_QuoteTo-12           10729615               112.7 ns/op
func BenchmarkDialector_QuoteTo(b *testing.B) {
	dailor := Open("")
	buf := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dailor.QuoteTo(buf, "datadase.table`User")
		buf.Reset()
	}
}


