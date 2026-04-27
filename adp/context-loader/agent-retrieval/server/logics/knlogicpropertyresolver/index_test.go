// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knlogicpropertyresolver

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestValidateRequest_Success 测试 validateRequest 成功场景
func TestValidateRequest_Success(t *testing.T) {
	convey.Convey("TestValidateRequest_Success", t, func() {
		service := &knLogicPropertyResolverService{}

		req := &interfaces.ResolveLogicPropertiesRequest{
			KnID:  "kn-001",
			OtID:  "ot-001",
			Query: "测试查询",
			InstanceIdentities: []map[string]interface{}{
				{"id": "obj-001"},
			},
			Properties: []string{"prop1", "prop2"},
		}

		err := service.validateRequest(req)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestValidateRequest_MissingKnID 测试 validateRequest 缺少 KnID
func TestValidateRequest_MissingKnID(t *testing.T) {
	convey.Convey("TestValidateRequest_MissingKnID", t, func() {
		service := &knLogicPropertyResolverService{}

		req := &interfaces.ResolveLogicPropertiesRequest{
			KnID:  "",
			OtID:  "ot-001",
			Query: "测试查询",
			InstanceIdentities: []map[string]interface{}{
				{"id": "obj-001"},
			},
			Properties: []string{"prop1"},
		}

		err := service.validateRequest(req)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "kn_id")
	})
}

// TestValidateRequest_MissingOtID 测试 validateRequest 缺少 OtID
func TestValidateRequest_MissingOtID(t *testing.T) {
	convey.Convey("TestValidateRequest_MissingOtID", t, func() {
		service := &knLogicPropertyResolverService{}

		req := &interfaces.ResolveLogicPropertiesRequest{
			KnID:  "kn-001",
			OtID:  "",
			Query: "测试查询",
			InstanceIdentities: []map[string]interface{}{
				{"id": "obj-001"},
			},
			Properties: []string{"prop1"},
		}

		err := service.validateRequest(req)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "ot_id")
	})
}

// TestValidateRequest_MissingQuery 测试 validateRequest 缺少 Query
func TestValidateRequest_MissingQuery(t *testing.T) {
	convey.Convey("TestValidateRequest_MissingQuery", t, func() {
		service := &knLogicPropertyResolverService{}

		req := &interfaces.ResolveLogicPropertiesRequest{
			KnID:  "kn-001",
			OtID:  "ot-001",
			Query: "",
			InstanceIdentities: []map[string]interface{}{
				{"id": "obj-001"},
			},
			Properties: []string{"prop1"},
		}

		err := service.validateRequest(req)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "query")
	})
}

// TestValidateRequest_EmptyInstanceIdentities 测试 validateRequest 空 InstanceIdentities
func TestValidateRequest_EmptyInstanceIdentities(t *testing.T) {
	convey.Convey("TestValidateRequest_EmptyInstanceIdentities", t, func() {
		service := &knLogicPropertyResolverService{}

		req := &interfaces.ResolveLogicPropertiesRequest{
			KnID:               "kn-001",
			OtID:               "ot-001",
			Query:              "测试查询",
			InstanceIdentities: []map[string]interface{}{},
			Properties:         []string{"prop1"},
		}

		err := service.validateRequest(req)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "_instance_identities")
	})
}

// TestValidateRequest_EmptyProperties 测试 validateRequest 空 Properties
func TestValidateRequest_EmptyProperties(t *testing.T) {
	convey.Convey("TestValidateRequest_EmptyProperties", t, func() {
		service := &knLogicPropertyResolverService{}

		req := &interfaces.ResolveLogicPropertiesRequest{
			KnID:  "kn-001",
			OtID:  "ot-001",
			Query: "测试查询",
			InstanceIdentities: []map[string]interface{}{
				{"id": "obj-001"},
			},
			Properties: []string{},
		}

		err := service.validateRequest(req)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "properties")
	})
}

// TestValidateMetricParams_Success_Instant 测试 validateMetricParams 即时查询成功
func TestValidateMetricParams_Success_Instant(t *testing.T) {
	convey.Convey("TestValidateMetricParams_Success_Instant", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		property := &interfaces.LogicPropertyDef{
			Name: "test_metric",
			Type: interfaces.LogicPropertyTypeMetric,
		}

		params := map[string]any{
			"instant": true,
			"start":   int64(1704067200000), // 2024-01-01
			"end":     int64(1706745600000), // 2024-02-01
		}

		ctx := context.Background()
		err := service.validateMetricParams(ctx, property, params)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestValidateMetricParams_Success_Trend 测试 validateMetricParams 趋势查询成功
func TestValidateMetricParams_Success_Trend(t *testing.T) {
	convey.Convey("TestValidateMetricParams_Success_Trend", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		property := &interfaces.LogicPropertyDef{
			Name: "test_metric",
			Type: interfaces.LogicPropertyTypeMetric,
		}

		params := map[string]any{
			"instant": false,
			"start":   int64(1704067200000),
			"end":     int64(1706745600000),
			"step":    "day",
		}

		ctx := context.Background()
		err := service.validateMetricParams(ctx, property, params)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestValidateMetricParams_MissingStart 测试 validateMetricParams 缺少 start
func TestValidateMetricParams_MissingStart(t *testing.T) {
	convey.Convey("TestValidateMetricParams_MissingStart", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		property := &interfaces.LogicPropertyDef{
			Name: "test_metric",
			Type: interfaces.LogicPropertyTypeMetric,
		}

		params := map[string]any{
			"instant": true,
			"end":     int64(1706745600000),
		}

		ctx := context.Background()
		err := service.validateMetricParams(ctx, property, params)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "start")
	})
}

// TestValidateMetricParams_MissingEnd 测试 validateMetricParams 缺少 end
func TestValidateMetricParams_MissingEnd(t *testing.T) {
	convey.Convey("TestValidateMetricParams_MissingEnd", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		property := &interfaces.LogicPropertyDef{
			Name: "test_metric",
			Type: interfaces.LogicPropertyTypeMetric,
		}

		params := map[string]any{
			"instant": true,
			"start":   int64(1704067200000),
		}

		ctx := context.Background()
		err := service.validateMetricParams(ctx, property, params)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "end")
	})
}

// TestValidateMetricParams_InstantWithStep 测试 instant=true 但有 step 的错误
func TestValidateMetricParams_InstantWithStep(t *testing.T) {
	convey.Convey("TestValidateMetricParams_InstantWithStep", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		property := &interfaces.LogicPropertyDef{
			Name: "test_metric",
			Type: interfaces.LogicPropertyTypeMetric,
		}

		params := map[string]any{
			"instant": true,
			"start":   int64(1704067200000),
			"end":     int64(1706745600000),
			"step":    "day", // instant=true 不应该有 step
		}

		ctx := context.Background()
		err := service.validateMetricParams(ctx, property, params)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "step")
	})
}

// TestValidateMetricParams_TrendWithoutStep 测试 instant=false 但没有 step 的错误
func TestValidateMetricParams_TrendWithoutStep(t *testing.T) {
	convey.Convey("TestValidateMetricParams_TrendWithoutStep", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		property := &interfaces.LogicPropertyDef{
			Name: "test_metric",
			Type: interfaces.LogicPropertyTypeMetric,
		}

		params := map[string]any{
			"instant": false,
			"start":   int64(1704067200000),
			"end":     int64(1706745600000),
			// 缺少 step
		}

		ctx := context.Background()
		err := service.validateMetricParams(ctx, property, params)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "step")
	})
}

// TestValidateMetricParams_InvalidStep 测试无效的 step 值
func TestValidateMetricParams_InvalidStep(t *testing.T) {
	convey.Convey("TestValidateMetricParams_InvalidStep", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		property := &interfaces.LogicPropertyDef{
			Name: "test_metric",
			Type: interfaces.LogicPropertyTypeMetric,
		}

		params := map[string]any{
			"instant": false,
			"start":   int64(1704067200000),
			"end":     int64(1706745600000),
			"step":    "invalid_step",
		}

		ctx := context.Background()
		err := service.validateMetricParams(ctx, property, params)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "invalid step")
	})
}

// TestValidateTimestamp_Int64 测试 int64 类型的时间戳
func TestValidateTimestamp_Int64(t *testing.T) {
	convey.Convey("TestValidateTimestamp_Int64", t, func() {
		service := &knLogicPropertyResolverService{}
		ctx := context.Background()

		// 有效时间戳
		err := service.validateTimestamp(ctx, int64(1704067200000), "start", "test_prop")
		convey.So(err, convey.ShouldBeNil)

		// 无效时间戳（太小）
		err = service.validateTimestamp(ctx, int64(100000000000), "start", "test_prop")
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestValidateTimestamp_Float64 测试 float64 类型的时间戳
func TestValidateTimestamp_Float64(t *testing.T) {
	convey.Convey("TestValidateTimestamp_Float64", t, func() {
		service := &knLogicPropertyResolverService{}
		ctx := context.Background()

		// 有效时间戳
		err := service.validateTimestamp(ctx, float64(1704067200000), "start", "test_prop")
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestValidateTimestamp_InvalidType 测试无效类型的时间戳
func TestValidateTimestamp_InvalidType(t *testing.T) {
	convey.Convey("TestValidateTimestamp_InvalidType", t, func() {
		service := &knLogicPropertyResolverService{}
		ctx := context.Background()

		// 无效类型
		err := service.validateTimestamp(ctx, "not_a_number", "start", "test_prop")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "must be a number")
	})
}

// TestExtractLogicProperties_Success 测试 extractLogicProperties 成功
func TestExtractLogicProperties_Success(t *testing.T) {
	convey.Convey("TestExtractLogicProperties_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		objectType := &interfaces.ObjectType{
			ID: "ot-001",
			LogicProperties: []*interfaces.LogicPropertyDef{
				{Name: "prop1", Type: interfaces.LogicPropertyTypeMetric},
				{Name: "prop2", Type: interfaces.LogicPropertyTypeOperator},
				{Name: "prop3", Type: interfaces.LogicPropertyTypeMetric},
			},
		}

		ctx := context.Background()
		result, err := service.extractLogicProperties(ctx, objectType, []string{"prop1", "prop2"})
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, 2)
		convey.So(result["prop1"], convey.ShouldNotBeNil)
		convey.So(result["prop2"], convey.ShouldNotBeNil)
	})
}

// TestExtractLogicProperties_NoLogicProperties 测试对象类没有逻辑属性
func TestExtractLogicProperties_NoLogicProperties(t *testing.T) {
	convey.Convey("TestExtractLogicProperties_NoLogicProperties", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		objectType := &interfaces.ObjectType{
			ID:              "ot-001",
			LogicProperties: []*interfaces.LogicPropertyDef{},
		}

		ctx := context.Background()
		_, err := service.extractLogicProperties(ctx, objectType, []string{"prop1"})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestExtractLogicProperties_PropertyNotFound 测试请求的属性不存在
func TestExtractLogicProperties_PropertyNotFound(t *testing.T) {
	convey.Convey("TestExtractLogicProperties_PropertyNotFound", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knLogicPropertyResolverService{
			logger: mockLogger,
		}

		objectType := &interfaces.ObjectType{
			ID: "ot-001",
			LogicProperties: []*interfaces.LogicPropertyDef{
				{Name: "prop1", Type: interfaces.LogicPropertyTypeMetric},
			},
		}

		ctx := context.Background()
		_, err := service.extractLogicProperties(ctx, objectType, []string{"prop1", "nonexistent_prop"})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "nonexistent_prop")
	})
}
