package mod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store"
)

func UploadDagInstanceEvents(ctx context.Context, dagIns *entity.DagInstance) error {

	config := common.NewConfig()
	og := drivenadapters.NewOssGateWay()
	eventRepo := rds.GetDagInstanceEventRepository()
	extDataRepo := rds.GetDagInstanceExtDataDao()

	if config.Server.DagInstanceEventArchivePolicy == common.DagInstanceEventArchivePolicyNever {
		return nil
	}

	ossId, err := og.GetAvaildOSS(ctx)

	if err != nil {
		return err
	}

	ossKey := fmt.Sprintf("%s/dag_instance/%s/events_%s.jsonl", config.Server.StoragePrefix, dagIns.ID, time.Now().Format("20060102150405"))

	patch := &entity.DagInstance{
		BaseInfo:         dagIns.BaseInfo,
		EventPersistence: entity.DagInstanceEventPersistenceOss,
		EventOssPath:     fmt.Sprintf("%s/%s", ossId, ossKey),
	}

	var size int64 = 0
	pr, pw := io.Pipe()
	go func() {
		var err error

		defer func() {
			if err != nil {
				pw.CloseWithError(err)
			} else {
				pw.Close()
			}
		}()

		batchSize := 50
		opts := &rds.DagInstanceEventListOptions{
			DagInstanceID: dagIns.ID,
			Offset:        0,
			Limit:         batchSize,
			Visibilities:  []rds.DagInstanceEventVisibility{rds.DagInstanceEventVisibilityPublic},
		}

		for {
			var rdsEvents []*rds.DagInstanceEvent
			if rdsEvents, err = eventRepo.List(context.Background(), opts); err != nil {
				return
			}

			if len(rdsEvents) == 0 {
				return
			}

			for _, ev := range rdsEvents {
				var event *entity.DagInstanceEvent
				if event, err = entity.FromRdsEvent(ctx, ev); err != nil {
					return
				}

				b, _ := json.Marshal(event)

				if _, err = pw.Write(b); err != nil {
					return
				}

				if _, err = pw.Write([]byte("\n\n")); err != nil {
					return
				}
			}

			opts.Offset += batchSize
		}
	}()

	if err := og.SimpleUpload(ctx, ossId, ossKey, true, pr); err != nil {
		log.Warnf("[UploadDagInstanceEvents] SimpleUpload err %s", err.Error())
		return err
	}

	if err := extDataRepo.InsertMany(ctx, []*rds.DagInstanceExtData{
		{
			ID:        store.NextStringID(),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			DagID:     dagIns.DagID,
			DagInsID:  dagIns.ID,
			Field:     "events",
			OssID:     ossId,
			OssKey:    ossKey,
			Size:      size,
			Removed:   false,
		},
	}); err != nil {
		log.Warnf("[UploadDagInstanceEvents] InsertMany err %s", err.Error())
		return err
	}

	if err = GetStore().PatchDagIns(ctx, patch); err != nil {
		log.Warnf("[UploadDagInstanceEvents] PatchDagIns err %s", err.Error())
		return err
	}

	if err := DeleteDagInstanceEvents(ctx, dagIns); err != nil {
		log.Warnf("[UploadDagInstanceEvents] DeleteDagInstanceEvents err %s", err.Error())
	}

	return nil
}

func DeleteDagInstanceEvents(ctx context.Context, dagIns *entity.DagInstance) error {
	og := drivenadapters.NewOssGateWay()
	eventRepo := rds.GetDagInstanceEventRepository()
	inline := false
	events, err := eventRepo.List(ctx, &rds.DagInstanceEventListOptions{
		DagInstanceID: dagIns.ID,
		Inline:        &inline,
		Fields:        []rds.DagInstanceEventField{rds.DagInstanceEventFieldData},
	})

	if err != nil {
		return err
	}

	for _, event := range events {
		parts := strings.SplitN(event.Data, "/", 2)
		ossID, ossKey := parts[0], parts[1]
		oerr := og.DeleteFile(ctx, ossID, ossKey, true)
		if oerr != nil {
			log.Warnf("[DagInstance.DeleteEvents] failed to delete event data, ins %s, event %d, err %v", dagIns.ID, event.ID, err)
		}
	}

	err = eventRepo.DeleteByInstanceIDs(ctx, []string{dagIns.ID})
	if err != nil {
		log.Warnf("[DagInstance.DeleteEvents] failed to delete events for instance ID %s: %v", dagIns.ID, err)
	}
	return err
}

func GetDagInstanceEvents(ctx context.Context, opts *rds.DagInstanceEventListOptions) ([]*rds.DagInstanceEvent, error) {
	eventRepo := rds.GetDagInstanceEventRepository()
	events, err := eventRepo.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	return events, err
}
