package kdb

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestDialector_Translate(t *testing.T) {
	type fields struct {
		Config *Config
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "it should return ErrDuplicatedKey error if the status code is 23505",
			args: args{err: &pgconn.PgError{Code: "23505"}},
			want: gorm.ErrDuplicatedKey,
		},
		{
			name: "it should return ErrForeignKeyViolated error if the status code is 23503",
			args: args{err: &pgconn.PgError{Code: "23503"}},
			want: gorm.ErrForeignKeyViolated,
		},
		{
			name: "it should return gorm.ErrInvalidField error if the status code is 42703",
			args: args{err: &pgconn.PgError{Code: "42703"}},
			want: gorm.ErrInvalidField,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialector := Dialector{
				Config: tt.fields.Config,
			}
			if err := dialector.Translate(tt.args.err); !errors.Is(err, tt.want) {
				t.Errorf("Translate() expected error = %v, got error %v", err, tt.want)
			}
		})
	}
}
