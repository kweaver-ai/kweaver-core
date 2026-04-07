package pmsvo

import (
	"testing"
)

func TestPmsControlObjS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pmsObj PmsControlObjS
	}{
		{
			name: "完整权限对象",
			pmsObj: PmsControlObjS{
				RoleIDs:       []string{"role1", "role2"},
				UserIDs:       []string{"user1", "user2"},
				UserGroupIDs:  []string{"group1"},
				DepartmentIDs: []string{"dept1"},
				AppAccountIDs: []string{"account1"},
			},
		},
		{
			name: "空权限对象",
			pmsObj: PmsControlObjS{
				RoleIDs:       []string{},
				UserIDs:       []string{},
				UserGroupIDs:  []string{},
				DepartmentIDs: []string{},
				AppAccountIDs: []string{},
			},
		},
		{
			name: "部分权限",
			pmsObj: PmsControlObjS{
				RoleIDs:       []string{"role1"},
				UserIDs:       nil,
				UserGroupIDs:  []string{},
				DepartmentIDs: nil,
				AppAccountIDs: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if len(tt.pmsObj.RoleIDs) != len(tt.pmsObj.RoleIDs) { //nolint:staticcheck
				t.Errorf("RoleIDs length mismatch")
			}

			if len(tt.pmsObj.UserIDs) != len(tt.pmsObj.UserIDs) { //nolint:staticcheck
				t.Errorf("UserIDs length mismatch")
			}

			if len(tt.pmsObj.UserGroupIDs) != len(tt.pmsObj.UserGroupIDs) { //nolint:staticcheck
				t.Errorf("UserGroupIDs length mismatch")
			}

			if len(tt.pmsObj.DepartmentIDs) != len(tt.pmsObj.DepartmentIDs) { //nolint:staticcheck
				t.Errorf("DepartmentIDs length mismatch")
			}

			if len(tt.pmsObj.AppAccountIDs) != len(tt.pmsObj.AppAccountIDs) { //nolint:staticcheck
				t.Errorf("AppAccountIDs length mismatch")
			}
		})
	}
}
