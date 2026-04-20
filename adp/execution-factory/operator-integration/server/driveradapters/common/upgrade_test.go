package common

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	"go.uber.org/mock/gomock"
)

func TestMigrateHistoryDataForSkillVersionValidation(t *testing.T) {
	t.Run("reject current version greater than 0.6.0", func(t *testing.T) {
		handler := &upgradeHandler{}

		_, err := handler.migrateHistoryDataForSkill(context.Background(), &MigrateHistoryDataRequest{
			CurrentVersion: "0.6.1",
			TargetVersion:  "0.7.0",
			PageSize:       10,
		})
		if err == nil {
			t.Fatal("expected version validation error")
		}
		if !strings.Contains(err.Error(), "0.6.0") || !strings.Contains(err.Error(), "0.7.0") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("reject target version less than 0.7.0", func(t *testing.T) {
		handler := &upgradeHandler{}

		_, err := handler.migrateHistoryDataForSkill(context.Background(), &MigrateHistoryDataRequest{
			CurrentVersion: "0.6.0",
			TargetVersion:  "0.6.9",
			PageSize:       10,
		})
		if err == nil {
			t.Fatal("expected version validation error")
		}
		if !strings.Contains(err.Error(), "0.6.0") || !strings.Contains(err.Error(), "0.7.0") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("allow boundary versions 0.6.0 to 0.7.0", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
		mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), nil, gomock.Any()).Return(int64(0), nil)

		handler := &upgradeHandler{
			SkillRepo: mockSkillRepo,
		}

		resp, err := handler.migrateHistoryDataForSkill(context.Background(), &MigrateHistoryDataRequest{
			CurrentVersion: "0.6.0",
			TargetVersion:  "0.7.0",
			PageSize:       10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Total != 0 {
			t.Fatalf("unexpected total: %d", resp.Total)
		}
	})
}

func TestMigrateHistoryDataForSkillAll(t *testing.T) {
	t.Run("all=true migrates all skills regardless of page", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rollbackPatch := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
			return nil
		})
		commitPatch := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
			return nil
		})
		defer func() {
			rollbackPatch.Reset()
			commitPatch.Reset()
		}()

		mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
		mockReleaseRepo := mocks.NewMockISkillReleaseDB(ctrl)
		mockHistoryRepo := &stubSkillReleaseHistoryRepo{}
		mockTx := mocks.NewMockDBTx(ctrl)
		tx := &sql.Tx{}

		skills := []*model.SkillRepositoryDB{
			{SkillID: "skill-1", Version: "v1", UpdateTime: 1, UpdateUser: "u1"},
			{SkillID: "skill-2", Version: "v2", UpdateTime: 2, UpdateUser: "u2"},
		}

		mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), nil, gomock.Any()).Return(int64(2), nil)
		mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), nil, gomock.Any(), nil, nil).
			DoAndReturn(func(ctx context.Context, tx *sql.Tx, filter map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) ([]*model.SkillRepositoryDB, error) {
				if filter["limit"] != 2 || filter["offset"] != 0 {
					t.Fatalf("unexpected filter: %#v", filter)
				}
				return skills, nil
			})

		mockTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil).Times(len(skills))
		for _, skill := range skills {
			mockReleaseRepo.EXPECT().SelectBySkillID(gomock.Any(), tx, skill.SkillID).Return(nil, nil)
			mockReleaseRepo.EXPECT().Insert(gomock.Any(), tx, gomock.Any()).Return(nil)
		}

		handler := &upgradeHandler{
			SkillRepo:               mockSkillRepo,
			SkillReleaseRepo:        mockReleaseRepo,
			SkillReleaseHistoryRepo: mockHistoryRepo,
			DBTx:                    mockTx,
		}

		resp, err := handler.migrateHistoryDataForSkill(context.Background(), &MigrateHistoryDataRequest{
			ALL:            true,
			Page:           3,
			PageSize:       1,
			CurrentVersion: "0.6.0",
			TargetVersion:  "0.7.0",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 2 || len(resp.Items) != 2 {
			t.Fatalf("unexpected response: %+v", resp)
		}
		if resp.Items[0].Id != "skill-1" || resp.Items[1].Id != "skill-2" {
			t.Fatalf("unexpected items: %+v", resp.Items)
		}
	})
}

type stubSkillReleaseHistoryRepo struct{}

func (s *stubSkillReleaseHistoryRepo) Insert(ctx context.Context, tx *sql.Tx, history *model.SkillReleaseHistoryDB) error {
	return nil
}

func (s *stubSkillReleaseHistoryRepo) SelectBySkillID(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillReleaseHistoryDB, error) {
	return nil, nil
}

func (s *stubSkillReleaseHistoryRepo) SelectBySkillIDAndVersion(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error) {
	return nil, nil
}

func (s *stubSkillReleaseHistoryRepo) DeleteByID(ctx context.Context, tx *sql.Tx, id int64) error {
	return nil
}

func (s *stubSkillReleaseHistoryRepo) DeleteBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error {
	return nil
}
