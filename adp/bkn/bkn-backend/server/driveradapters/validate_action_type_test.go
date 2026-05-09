// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func Test_ValidateActionType(t *testing.T) {
	Convey("Test ValidateActionType\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid action type\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with strictMode false: empty ObjectTypeID and invalid condition not validated\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at-relaxed",
					ATName:       "action_relaxed",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
					Condition: &interfaces.ActionCondCfg{
						Field: "field1",
						// Operation omitted — invalid under strict validation
					},
				},
			}
			err := ValidateActionType(ctx, at, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid ID\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "_invalid_id",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty name\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionType_NullParameter_Name)
		})

		Convey("Failed with invalid action source type\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
					ActionSource: interfaces.ActionSource{
						Type: "invalid_type",
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with tool type having mcp data\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ActionSource: interfaces.ActionSource{
						Type:  interfaces.ACTION_SOURCE_TYPE_TOOL,
						McpID: "mcp1",
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with tool type having tool_name\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ActionSource: interfaces.ActionSource{
						Type:     interfaces.ACTION_SOURCE_TYPE_TOOL,
						ToolName: "tool1",
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with tool type without mcp data\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ActionSource: interfaces.ActionSource{
						Type: interfaces.ACTION_SOURCE_TYPE_TOOL,
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with mcp type without tool data\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ActionSource: interfaces.ActionSource{
						Type: interfaces.ACTION_SOURCE_TYPE_MCP,
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with empty parameter name\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Parameters: []interfaces.Parameter{
						{
							Name: "",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with mcp type having box_id\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionSource: interfaces.ActionSource{
						Type:  interfaces.ACTION_SOURCE_TYPE_MCP,
						BoxID: "box1",
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with mcp type having tool_id\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ActionSource: interfaces.ActionSource{
						Type:   interfaces.ACTION_SOURCE_TYPE_MCP,
						ToolID: "tool1",
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with strictMode false: non-empty ActionSource.Type and empty or mixed binding fields allowed\n", func() {
			atTool := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at-draft-tool",
					ATName:       "action_draft",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ActionSource: interfaces.ActionSource{
						Type:     interfaces.ACTION_SOURCE_TYPE_TOOL,
						BoxID:    "",
						ToolID:   "",
						McpID:    "",
						ToolName: "",
					},
				},
			}
			So(ValidateActionType(ctx, atTool, false), ShouldBeNil)

			atToolMixed := *atTool
			atToolMixed.ActionSource.McpID = "mcp1"
			atToolMixed.ActionSource.ToolName = "tn"
			So(ValidateActionType(ctx, &atToolMixed, false), ShouldBeNil)

			atMcp := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at-draft-mcp",
					ATName:       "action_draft_mcp",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ActionSource: interfaces.ActionSource{
						Type:     interfaces.ACTION_SOURCE_TYPE_MCP,
						McpID:    "",
						ToolName: "",
						BoxID:    "",
						ToolID:   "",
					},
				},
			}
			So(ValidateActionType(ctx, atMcp, false), ShouldBeNil)

			atMcpMixed := *atMcp
			atMcpMixed.ActionSource.BoxID = "box1"
			atMcpMixed.ActionSource.ToolID = "tid1"
			So(ValidateActionType(ctx, &atMcpMixed, false), ShouldBeNil)
		})

		Convey("Success with valid condition\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    cond.OperationEq,
						ValueOptCfg: cond.ValueOptCfg{
							Value: "value1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		// Convey("Failed with condition missing ObjectTypeID\n", func() {
		// 	at := &interfaces.ActionType{
		// 		ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
		// 			ATID:   "at1",
		// 			ATName: "action1",
		// 			Condition: &interfaces.ActionCondCfg{
		// 				Field:     "field1",
		// 				Operation: cond.OperationEq,
		// 				ValueOptCfg: cond.ValueOptCfg{
		// 					Value: "value1",
		// 				},
		// 			},
		// 		},
		// 	}
		// 	err := ValidateActionType(ctx, at, true)
		// 	So(err, ShouldNotBeNil)
		// })

		Convey("Failed with condition missing Operation\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						ValueOptCfg: cond.ValueOptCfg{
							Value: "value1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid operation\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    "invalid_op",
						ValueOptCfg: cond.ValueOptCfg{
							Value: "value1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with and operation missing field\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Operation:    cond.OperationAnd,
						SubConds: []*interfaces.ActionCondCfg{
							{
								ObjectTypeID: "ot1",
								Field:        "field1",
								Operation:    cond.OperationEq,
								ValueOptCfg: cond.ValueOptCfg{
									Value: "value1",
								},
							},
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil) // and/or operation doesn't require field
		})

		Convey("Failed with eq operation having array value\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    "eq",
						ValueOptCfg: cond.ValueOptCfg{
							Value: []any{"value1", "value2"},
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with in operation missing array value\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    cond.OperationIn,
						ValueOptCfg: cond.ValueOptCfg{
							Value: "value1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with in operation having empty array\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    cond.OperationIn,
						ValueOptCfg: cond.ValueOptCfg{
							Value: []any{},
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with range operation missing array value\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    cond.OperationRange,
						ValueOptCfg: cond.ValueOptCfg{
							Value: "value1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with range operation having wrong array length\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    cond.OperationRange,
						ValueOptCfg: cond.ValueOptCfg{
							Value: []any{"value1"},
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with valid nested condition\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Operation:    cond.OperationAnd,
						SubConds: []*interfaces.ActionCondCfg{
							{
								ObjectTypeID: "ot1",
								Field:        "field1",
								Operation:    cond.OperationEq,
								ValueOptCfg: cond.ValueOptCfg{
									Value: "value1",
								},
							},
							{
								ObjectTypeID: "ot1",
								Field:        "field2",
								Operation:    cond.OperationIn,
								ValueOptCfg: cond.ValueOptCfg{
									Value: []any{"value2", "value3"},
								},
							},
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with empty ObjectTypeID\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with empty ObjectTypeID but having condition\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    cond.OperationEq,
						ValueOptCfg: cond.ValueOptCfg{
							Value: "value1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionType_InvalidParameter)
		})

		Convey("Success with empty ObjectTypeID and condition when strictMode false\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1",
						Field:        "field1",
						Operation:    cond.OperationEq,
						ValueOptCfg: cond.ValueOptCfg{
							Value: "value1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed with empty ObjectTypeID but parameter using property (strict)\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_PROPERTY,
							Value:     "prop1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionType_InvalidParameter)
		})

		Convey("Success with empty ObjectTypeID and parameter using property when strictMode false\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_PROPERTY,
							Value:     "prop1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, false)
			So(err, ShouldBeNil)
		})

		Convey("Success with empty ObjectTypeID and parameter using const\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_CONST,
							Value:     "const_value",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with empty ObjectTypeID and parameter using input\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_INPUT,
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with ObjectTypeID and parameter using property\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_PROPERTY,
							Value:     "prop1",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with ObjectTypeID and parameter using const\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_CONST,
							Value:     "const_value",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with ObjectTypeID and parameter using input\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_INPUT,
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with ObjectTypeID and multiple parameters with different ValueFrom\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
					ObjectTypeID: "ot1",
					Parameters: []interfaces.Parameter{
						{
							Name:      "param1",
							ValueFrom: interfaces.VALUE_FROM_PROPERTY,
							Value:     "prop1",
						},
						{
							Name:      "param2",
							ValueFrom: interfaces.VALUE_FROM_CONST,
							Value:     "const_value",
						},
						{
							Name:      "param3",
							ValueFrom: interfaces.VALUE_FROM_INPUT,
						},
					},
				},
			}
			err := ValidateActionType(ctx, at, true)
			So(err, ShouldBeNil)
		})
	})
}
