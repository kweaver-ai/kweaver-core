package skill

import (
	"context"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestOSSGatewaySkillAssetStoreUploadUsesCurrentStorageIDEvenWhenClientStartsNotReady(t *testing.T) {
	Convey("Upload relies on CurrentStorageID instead of failing early on IsReady", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockOSSGatewayBackendClient(ctrl)
		store := &ossGatewaySkillAssetStore{
			client:      mockClient,
			SkillPrefix: "skill/",
		}

		mockClient.EXPECT().CurrentStorageID(gomock.Any()).Return("storage-1", nil)
		mockClient.EXPECT().UploadFile(gomock.Any(), &interfaces.OssObject{
			StorageID:  "storage-1",
			StorageKey: "skill/skill-1/v1/refs/guide.md",
		}, []byte("content")).Return(nil)

		object, checksum, err := store.Upload(context.Background(), "skill-1", "v1", "refs/guide.md", []byte("content"))
		So(err, ShouldBeNil)
		So(object, ShouldResemble, &interfaces.OssObject{
			StorageID:  "storage-1",
			StorageKey: "skill/skill-1/v1/refs/guide.md",
		})
		So(checksum, ShouldEqual, checksumSHA256([]byte("content")))
	})
}
