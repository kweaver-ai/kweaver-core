package releasesvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandleCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		categoryIDs  []string
		releaseID    string
		setup        func(*gomock.Controller) (*releaseSvc, *sql.Tx)
		wantErr      bool
		errContains  string
		expectDelete bool
		expectCreate bool
	}{
		{
			name:        "empty category IDs - no operation",
			categoryIDs: []string{},
			releaseID:   "release-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, *sql.Tx) {
				repo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
				tx := &sql.Tx{}

				svc := &releaseSvc{
					SvcBase:                service.NewSvcBase(),
					releaseCategoryRelRepo: repo,
				}
				return svc, tx
			},
			wantErr:      false,
			expectDelete: false,
			expectCreate: false,
		},
		{
			name:        "nil category IDs - no operation",
			categoryIDs: nil,
			releaseID:   "release-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, *sql.Tx) {
				repo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
				tx := &sql.Tx{}

				svc := &releaseSvc{
					SvcBase:                service.NewSvcBase(),
					releaseCategoryRelRepo: repo,
				}
				return svc, tx
			},
			wantErr:      false,
			expectDelete: false,
			expectCreate: false,
		},
		{
			name:        "successful category handling",
			categoryIDs: []string{"cat-1", "cat-2"},
			releaseID:   "release-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, *sql.Tx) {
				repo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
				tx := &sql.Tx{}

				repo.EXPECT().
					DelByReleaseID(gomock.Any(), tx, "release-123").
					Return(nil)

				repo.EXPECT().
					BatchCreate(gomock.Any(), tx, gomock.Any()).
					Return(nil)

				svc := &releaseSvc{
					SvcBase:                service.NewSvcBase(),
					releaseCategoryRelRepo: repo,
				}
				return svc, tx
			},
			wantErr: false,
		},
		{
			name:        "category IDs with whitespace - trimmed",
			categoryIDs: []string{" cat-1 ", "  cat-2  "},
			releaseID:   "release-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, *sql.Tx) {
				repo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
				tx := &sql.Tx{}

				repo.EXPECT().
					DelByReleaseID(gomock.Any(), tx, "release-123").
					Return(nil)

				repo.EXPECT().
					BatchCreate(gomock.Any(), tx, gomock.Any()).
					Return(nil)

				svc := &releaseSvc{
					SvcBase:                service.NewSvcBase(),
					releaseCategoryRelRepo: repo,
				}
				return svc, tx
			},
			wantErr: false,
		},
		{
			name:        "category IDs with empty strings - filtered out",
			categoryIDs: []string{"cat-1", "", "cat-2", ""},
			releaseID:   "release-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, *sql.Tx) {
				repo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
				tx := &sql.Tx{}

				repo.EXPECT().
					DelByReleaseID(gomock.Any(), tx, "release-123").
					Return(nil)

				repo.EXPECT().
					BatchCreate(gomock.Any(), tx, gomock.Any()).
					Return(nil)

				svc := &releaseSvc{
					SvcBase:                service.NewSvcBase(),
					releaseCategoryRelRepo: repo,
				}
				return svc, tx
			},
			wantErr: false,
		},
		{
			name:        "delete category relations fails",
			categoryIDs: []string{"cat-1"},
			releaseID:   "release-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, *sql.Tx) {
				repo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
				tx := &sql.Tx{}

				repo.EXPECT().
					DelByReleaseID(gomock.Any(), tx, "release-123").
					Return(errors.New("database error"))

				svc := &releaseSvc{
					SvcBase:                service.NewSvcBase(),
					releaseCategoryRelRepo: repo,
				}
				return svc, tx
			},
			wantErr:     true,
			errContains: "delete category relations failed",
		},
		{
			name:        "batch create fails",
			categoryIDs: []string{"cat-1"},
			releaseID:   "release-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, *sql.Tx) {
				repo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
				tx := &sql.Tx{}

				repo.EXPECT().
					DelByReleaseID(gomock.Any(), tx, "release-123").
					Return(nil)

				repo.EXPECT().
					BatchCreate(gomock.Any(), tx, gomock.Any()).
					Return(errors.New("batch create failed"))

				svc := &releaseSvc{
					SvcBase:                service.NewSvcBase(),
					releaseCategoryRelRepo: repo,
				}
				return svc, tx
			},
			wantErr:     true,
			errContains: "batch create category relations failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, tx := tt.setup(ctrl)
			ctx := context.Background()
			err := svc.handleCategory(ctx, tt.categoryIDs, tt.releaseID, tx)

			if tt.wantErr {
				if tt.errContains != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errContains)
				} else {
					require.NoError(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
