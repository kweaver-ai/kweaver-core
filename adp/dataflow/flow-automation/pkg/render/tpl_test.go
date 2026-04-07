package render

import (
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCachedTplGetter_GetTpl(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	type fields struct {
		tplProvider *TplProvider
	}
	type args struct {
		tplTexts []string
	}
	type TestCace struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}
	tests := []TestCace{
		{
			name: "once succ",
			fields: fields{
				tplProvider: NewCachedTplProvider(10),
			},
			args: args{
				tplTexts: []string{
					"aaa",
				},
			},
			wantErr: false,
		},
		{
			name: "once failed",
			fields: fields{
				tplProvider: NewCachedTplProvider(10),
			},
			args: args{
				tplTexts: []string{
					"{{a}}",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.tplProvider
			for i, text := range tt.args.tplTexts {
				_, err := c.GetTpl(text)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetTpl() index:%d error = %v, wantErr %v", i, err, tt.wantErr)
					return
				}
			}

		})
	}
}
