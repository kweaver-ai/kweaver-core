package sqlhelper2

import (
	"reflect"
	"testing"
)

func TestAllFieldsByStruct(t *testing.T) {
	t.Parallel()

	type args struct {
		s    interface{}
		tags []string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				s: struct {
					ID   int    `db:"id"`
					Name string `db:"name"`
				}{},
				tags: []string{"db"},
			},
			want: []string{"id", "name"},
		},
		{
			name: "test2",
			args: args{
				s: &struct {
					ID   int    `db2:"id2"`
					Name string `db2:"name2"`
				}{},
				tags: []string{"db2"},
			},
			want: []string{"id2", "name2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := AllFieldsByStruct(tt.args.s, tt.args.tags...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AllFieldsByStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllFieldsByStruct_Anonymous(t *testing.T) {
	t.Parallel()

	type args struct {
		s    interface{}
		tags []string
	}

	type Anonymous struct {
		AnonymousID   int    `db:"anonymous_id"`
		AnonymousName string `db:"anonymous_name"`
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				s: struct {
					ID   int    `db:"id"`
					Name string `db:"name"`
					Anonymous
				}{},
				tags: []string{"db"},
			},
			want: []string{"id", "name", "anonymous_id", "anonymous_name"},
		},
		{
			name: "test2",
			args: args{
				s: &struct {
					ID   int    `db:"id2"`
					Name string `db:"name2"`
					*Anonymous
				}{},
				tags: []string{"db"},
			},
			want: []string{"id2", "name2", "anonymous_id", "anonymous_name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := AllFieldsByStruct(tt.args.s, tt.args.tags...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AllFieldsByStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenSqlSelectFieldsStr(t *testing.T) {
	t.Parallel()

	type args struct {
		fields    []string
		tableFlag string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				fields:    []string{"id", "name"},
				tableFlag: "",
			},
			want: "id,name",
		},
		{
			name: "test2",
			args: args{
				fields:    []string{"id", "name"},
				tableFlag: "t1",
			},
			want: "t1.id,t1.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := GenSQLSelectFieldsStr(tt.args.fields, tt.args.tableFlag); got != tt.want {
				t.Errorf("GenSQLSelectFieldsStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllFieldsByStruct_AnonymousNonStruct(t *testing.T) {
	t.Parallel()

	// Test anonymous field that is not a struct type
	type NotAStruct string

	tests := []struct {
		name string
		s    interface{}
		tags []string
		want []string
	}{
		{
			name: "anonymous non-struct field (NotAStruct alias)",
			s: struct {
				ID         int `db:"id"`
				NotAStruct     // Anonymous field of non-struct type
			}{},
			tags: []string{"db"},
			want: []string{"id"}, // NotAStrut should be skipped
		},
		{
			name: "regular fields without anonymous",
			s: struct {
				ID     int    `db:"id"`
				Name   string `db:"name"`
				Active bool   `db:"active"`
			}{},
			tags: []string{"db"},
			want: []string{"id", "name", "active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := AllFieldsByStruct(tt.s, tt.tags...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AllFieldsByStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
