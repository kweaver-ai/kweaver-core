// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"bytes"
	"context"
	"sync"

	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
	"bkn-backend/logics/knowledge_network"
)

var (
	bServiceOnce sync.Once
	bService     interfaces.BKNService
)

type bknService struct {
	appSetting *common.AppSetting
	kns        interfaces.KNService
}

// NewBKNService 创建 BKN 服务
func NewBKNService(appSetting *common.AppSetting) interfaces.BKNService {
	bServiceOnce.Do(func() {
		bService = &bknService{
			appSetting: appSetting,
			kns:        knowledge_network.NewKNService(appSetting),
		}
	})
	return bService
}

// ExportToTar 将知识网络导出为 tar 包
func (bs *bknService) ExportToTar(ctx context.Context, knID string, branch string) ([]byte, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "BKN导出为Tar")
	defer span.End()

	logger.Debugf("BKN ExportToTar Start: kn_id=%s", knID)

	kn, err := bs.kns.GetKNByID(ctx, knID, branch, interfaces.Mode_Export)
	if err != nil {
		logger.Errorf("BKN GetKNByID failed: %s", err.Error())
		return nil, err
	}

	bknNetwork := logics.ToBKNNetWork(kn)
	for _, ot := range kn.ObjectTypes {
		bknNetwork.ObjectTypes = append(bknNetwork.ObjectTypes, logics.ToBKNObjectType(ot))
	}
	for _, rt := range kn.RelationTypes {
		bknNetwork.RelationTypes = append(bknNetwork.RelationTypes, logics.ToBKNRelationType(rt))
	}
	for _, act := range kn.ActionTypes {
		bknNetwork.ActionTypes = append(bknNetwork.ActionTypes, logics.ToBKNActionType(act))
	}
	for _, risk := range kn.RiskTypes {
		bknNetwork.RiskTypes = append(bknNetwork.RiskTypes, logics.ToBKNRiskType(risk))
	}
	for _, cg := range kn.ConceptGroups {
		bknNetwork.ConceptGroups = append(bknNetwork.ConceptGroups, logics.ToBKNConceptGroup(cg))
	}

	var buf bytes.Buffer
	err = bknsdk.WriteNetworkToTar(bknNetwork, &buf)
	if err != nil {
		logger.Errorf("BKN ExportToTar failed: %s", err.Error())
		return nil, err
	}
	tarData := buf.Bytes()

	logger.Debugf("BKN ExportToTar Completed: size=%d", len(tarData))
	return tarData, nil
}
