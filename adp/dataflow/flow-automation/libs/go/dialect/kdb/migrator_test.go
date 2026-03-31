package kdb

import "testing"

func Test_parseDefaultValueValue(t *testing.T) {
	type args struct {
		defaultValue string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "it should works with number without colons",
			args: args{defaultValue: "0"},
			want: "0",
		},
		{
			name: "it should works with number and two colons",
			args: args{defaultValue: "0::int8"},
			want: "0",
		},
		{
			name: "it should works with number and three colons",
			args: args{defaultValue: "0:::int8"},
			want: "0",
		},
		{
			name: "it should works with empty string without colons",
			args: args{defaultValue: "''"},
			want: "",
		},
		{
			name: "it should works with empty string with two colons",
			args: args{defaultValue: "''::character varying"},
			want: "",
		},
		{
			name: "it should works with empty string with three colons",
			args: args{defaultValue: "'':::character varying"},
			want: "",
		},
		{
			name: "it should works with string without colons",
			args: args{defaultValue: "'field'"},
			want: "field",
		},
		{
			name: "it should works with string with two colons",
			args: args{defaultValue: "'field'::character varying"},
			want: "field",
		},
		{
			name: "it should works with string with three colons",
			args: args{defaultValue: "'field':::character varying"},
			want: "field",
		},
		{
			name: "it should works with value with two colons",
			args: args{defaultValue: "field"},
			want: "field",
		},
		{
			name: "it should works with value without colons",
			args: args{defaultValue: "field::character varying"},
			want: "field",
		},
		{
			name: "it should works with value with three colons",
			args: args{defaultValue: "field:::character varying"},
			want: "field",
		},
		{
			name: "it should works with function without colons",
			args: args{defaultValue: "now()"},
			want: "now()",
		},
		{
			name: "it should works with function with two colons",
			args: args{defaultValue: "now()::timestamp without time zone"},
			want: "now()",
		},
		{
			name: "it should works with json without colons",
			args: args{defaultValue: "{}"},
			want: "{}",
		},
		{
			name: "it should works with json with two colons",
			args: args{defaultValue: "{}::jsonb"},
			want: "{}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDefaultValueValue(tt.args.defaultValue); got != tt.want {
				t.Errorf("parseDefaultValueValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
