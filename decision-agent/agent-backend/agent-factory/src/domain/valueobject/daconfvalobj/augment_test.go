package daconfvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAugment_ValObjCheck(t *testing.T) {
	t.Parallel()

	validKg := KgSource{
		KgID:   "kg1",
		Fields: []string{"field1", "field2"},
	}
	enableTrue := true
	enableFalse := false

	tests := []struct {
		name    string
		augment *Augment
		wantErr bool
	}{
		{
			name: "enable为true且有有效DataSource",
			augment: &Augment{
				Enable: &enableTrue,
				DataSource: &AugmentDataSource{
					Kg: []KgSource{validKg},
				},
			},
			wantErr: false,
		},
		{
			name: "enable为false",
			augment: &Augment{
				Enable:     &enableFalse,
				DataSource: nil,
			},
			wantErr: false,
		},
		{
			name: "enable为nil",
			augment: &Augment{
				Enable:     nil,
				DataSource: nil,
			},
			wantErr: true,
		},
		{
			name: "enable为true但DataSource为nil",
			augment: &Augment{
				Enable:     &enableTrue,
				DataSource: nil,
			},
			wantErr: true,
		},
		{
			name: "enable为true但DataSource的Kg为空",
			augment: &Augment{
				Enable: &enableTrue,
				DataSource: &AugmentDataSource{
					Kg: []KgSource{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.augment.ValObjCheck()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAugment_DataSource_ValObjCheck(t *testing.T) {
	t.Parallel()

	validKg := KgSource{
		KgID:   "kg1",
		Fields: []string{"field1", "field2"},
	}
	invalidKg := KgSource{
		KgID: "kg-invalid",
	}

	tests := []struct {
		name    string
		ads     *AugmentDataSource
		wantErr bool
	}{
		{
			name: "有效的Kg配置",
			ads: &AugmentDataSource{
				Kg: []KgSource{validKg},
			},
			wantErr: false,
		},
		{
			name:    "空的Kg列表",
			ads:     &AugmentDataSource{},
			wantErr: true,
		},
		{
			name: "Kg包含无效项",
			ads: &AugmentDataSource{
				Kg: []KgSource{invalidKg},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.ads.ValObjCheck()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValObjCheck() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValObjCheck() unexpected error: %v", err)
				}
			}
		})
	}
}
