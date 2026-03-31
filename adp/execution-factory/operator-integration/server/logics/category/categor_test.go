package category

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestGetCategoryName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDBTx := mocks.NewMockDBTx(ctrl)
	mockDBCategory := mocks.NewMockDBCategory(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	manager := &categoryManager{
		logger:     logger.DefaultLogger(),
		DBTx:       mockDBTx,
		DBCategory: mockDBCategory,
		Validator:  mockValidator,
		Cache:      mockCache,
	}
	Convey("TestGetCategoryName:获取分类名称", t, func() {
		Convey("分类名称为空", func() {
			name := manager.GetCategoryName(context.TODO(), interfaces.BizCategory(""))
			So(name, ShouldEqual, "")
		})
		Convey("内置分类: 其他", func() {
			ctx := context.Background()
			ctx = common.SetLanguageByCtx(ctx, common.SimplifiedChinese)
			name := manager.GetCategoryName(ctx, interfaces.CategoryTypeOther)
			So(name, ShouldEqual, "未分类")
			ctx = common.SetLanguageByCtx(ctx, common.AmericanEnglish)
			name = manager.GetCategoryName(ctx, interfaces.CategoryTypeOther)
			So(name, ShouldEqual, "Uncategorized")
		})
		Convey("内置分类: 系统", func() {
			ctx := context.Background()
			ctx = common.SetLanguageByCtx(ctx, common.SimplifiedChinese)
			name := manager.GetCategoryName(ctx, interfaces.CategoryTypeSystem)
			So(name, ShouldEqual, "系统工具")
			ctx = common.SetLanguageByCtx(ctx, common.AmericanEnglish)
			name = manager.GetCategoryName(ctx, interfaces.CategoryTypeSystem)
			So(name, ShouldEqual, "System")
		})
		Convey("自定义分类", func() {
			mockName := "自定义分类"
			Convey("缓存命中", func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(mockName, true).Times(1)
				name := manager.GetCategoryName(context.TODO(), interfaces.BizCategory(mockName))
				So(name, ShouldEqual, mockName)
			})
			Convey("从数据库中查询失败（db）", func() {
				mockCache.EXPECT().Get(gomock.Any()).Return("", false).Times(1)
				mockDBCategory.EXPECT().SelectListByCategoryID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("mock SelectListByCategoryID error")).Times(1)
				name := manager.GetCategoryName(context.TODO(), interfaces.BizCategory(mockName))
				So(name, ShouldEqual, "")
			})
			Convey("数据库中不存在", func() {
				mockCache.EXPECT().Get(gomock.Any()).Return("", false).Times(1)
				mockDBCategory.EXPECT().SelectListByCategoryID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil).Times(1)
				name := manager.GetCategoryName(context.TODO(), interfaces.BizCategory(mockName))
				So(name, ShouldEqual, "")
			})
			Convey("数据库中存在，返回并保存到缓存", func() {
				mockCache.EXPECT().Get(gomock.Any()).Return("", false).Times(1)
				mockCache.EXPECT().Set(gomock.Any(), gomock.Any()).Times(1)
				mockDBCategory.EXPECT().SelectListByCategoryID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&model.CategoryDB{
						CategoryName: mockName,
					}, nil).Times(1)
				name := manager.GetCategoryName(context.TODO(), interfaces.BizCategory(mockName))
				So(name, ShouldEqual, mockName)
			})
		})
	})
}

func TestCheckCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDBTx := mocks.NewMockDBTx(ctrl)
	mockDBCategory := mocks.NewMockDBCategory(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	manager := &categoryManager{
		logger:     logger.DefaultLogger(),
		DBTx:       mockDBTx,
		DBCategory: mockDBCategory,
		Validator:  mockValidator,
		Cache:      mockCache,
	}
	Convey("TestCheckCategory:检查分类是否存在", t, func() {
		Convey("内置分类", func() {
			So(manager.CheckCategory(interfaces.CategoryTypeOther), ShouldBeTrue)
			So(manager.CheckCategory(interfaces.CategoryTypeSystem), ShouldBeTrue)
		})
		Convey("自定义分类", func() {
			Convey("获取自定义分类报错（db）", func() {
				mockDBCategory.EXPECT().SelectListByCategoryID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("mock SelectListByCategoryID error")).Times(1)
				isExist := manager.CheckCategory(interfaces.BizCategory(""))
				So(isExist, ShouldBeFalse)
			})
			Convey("自定义分类为空", func() {
				mockDBCategory.EXPECT().SelectListByCategoryID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil).Times(1)
				isExist := manager.CheckCategory(interfaces.BizCategory(""))
				So(isExist, ShouldBeFalse)
			})
			Convey("自定义分类不为空", func() {
				mockDBCategory.EXPECT().SelectListByCategoryID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&model.CategoryDB{}, nil).Times(1)
				isExist := manager.CheckCategory(interfaces.BizCategory(""))
				So(isExist, ShouldBeTrue)
			})
		})
	})
}
