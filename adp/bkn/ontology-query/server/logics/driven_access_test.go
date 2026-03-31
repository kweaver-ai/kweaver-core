// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"testing"

	omock "ontology-query/interfaces/mock"

	"go.uber.org/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_SetAgentOperatorAccess(t *testing.T) {
	Convey("Test SetAgentOperatorAccess", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		aoa := omock.NewMockAgentOperatorAccess(mockCtrl)

		SetAgentOperatorAccess(aoa)
		So(AOA, ShouldEqual, aoa)
	})
}

func Test_SetModelFactoryAccess(t *testing.T) {
	Convey("Test SetModelFactoryAccess", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mfa := omock.NewMockModelFactoryAccess(mockCtrl)

		SetModelFactoryAccess(mfa)
		So(MFA, ShouldEqual, mfa)
	})
}

func Test_SetOntologyManagerAccess(t *testing.T) {
	Convey("Test SetOntologyManagerAccess", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		oma := omock.NewMockOntologyManagerAccess(mockCtrl)

		SetOntologyManagerAccess(oma)
		So(OMA, ShouldEqual, oma)
	})
}

func Test_SetOpenSearchAccess(t *testing.T) {
	Convey("Test SetOpenSearchAccess", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := omock.NewMockOpenSearchAccess(mockCtrl)

		SetOpenSearchAccess(osa)
		So(OSA, ShouldEqual, osa)
	})
}

func Test_SetUniqueryAccess(t *testing.T) {
	Convey("Test SetUniqueryAccess", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ua := omock.NewMockUniqueryAccess(mockCtrl)

		SetUniqueryAccess(ua)
		So(UA, ShouldEqual, ua)
	})
}

func Test_GlobalVariables(t *testing.T) {
	Convey("Test Global Variables", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		Convey("成功 - 设置所有全局变量", func() {
			aoa := omock.NewMockAgentOperatorAccess(mockCtrl)
			mfa := omock.NewMockModelFactoryAccess(mockCtrl)
			oma := omock.NewMockOntologyManagerAccess(mockCtrl)
			osa := omock.NewMockOpenSearchAccess(mockCtrl)
			ua := omock.NewMockUniqueryAccess(mockCtrl)

			SetAgentOperatorAccess(aoa)
			SetModelFactoryAccess(mfa)
			SetOntologyManagerAccess(oma)
			SetOpenSearchAccess(osa)
			SetUniqueryAccess(ua)

			So(AOA, ShouldEqual, aoa)
			So(MFA, ShouldEqual, mfa)
			So(OMA, ShouldEqual, oma)
			So(OSA, ShouldEqual, osa)
			So(UA, ShouldEqual, ua)
		})

		Convey("成功 - 多次设置同一变量", func() {
			aoa1 := omock.NewMockAgentOperatorAccess(mockCtrl)
			aoa2 := omock.NewMockAgentOperatorAccess(mockCtrl)

			SetAgentOperatorAccess(aoa1)
			So(AOA, ShouldEqual, aoa1)

			SetAgentOperatorAccess(aoa2)
			So(AOA, ShouldEqual, aoa2)
		})
	})
}
