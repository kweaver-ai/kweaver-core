package toolbox

import (
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterToolboxResourceIDs(t *testing.T) {
	Convey("业务域返回全量哨兵时应跳过业务域过滤", t, func() {
		resourceToBdMap := map[string]string{interfaces.ResourceIDAll: ""}

		Convey("有限权限时直接使用 auth 结果", func() {
			resourceIDs := filterToolboxResourceIDs(resourceToBdMap, []string{"box-1", "box-2"})
			So(resourceIDs, ShouldResemble, []string{"box-1", "box-2"})
		})

		Convey("全量权限时返回全量哨兵", func() {
			resourceIDs := filterToolboxResourceIDs(resourceToBdMap, []string{interfaces.ResourceIDAll})
			So(resourceIDs, ShouldResemble, []string{interfaces.ResourceIDAll})
		})
	})
}
