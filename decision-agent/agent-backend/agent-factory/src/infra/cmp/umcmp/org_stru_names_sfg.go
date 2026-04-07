package umcmp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"golang.org/x/sync/singleflight"
)

var _getOsnNamesSFG singleflight.Group

// GetOsnNamesSFG 获取组织架构对象的names
//
//nolint:funlen
func (u *Um) GetOsnNamesSFG(ctx context.Context, args *umarg.GetOsnArgDto) (ret *umtypes.OsnInfoMapS, err error) {
	// 1. 生成sfg key
	key, err := args.ToSfgKey()
	if err != nil {
		return
	}

	// 2. 使用sfg
	val, err, _ := _getOsnNamesSFG.Do(key, func() (interface{}, error) {
		_ret, _err := u.GetOsnNames(ctx, args)
		tRes := &sfgTmpResGetOsnNamesRet{
			Ret: _ret,
		}

		return tRes, _err
	})

	if err != nil {
		return
	}

	// 3. val转换为pos
	sfgRes := &sfgTmpResGetOsnNamesRet{}

	err = sfgRes.LoadFromValByCopyWithJSON(val)
	if err != nil {
		return
	}

	ret = sfgRes.Ret

	return
}

type sfgTmpResGetOsnNamesRet struct {
	Ret *umtypes.OsnInfoMapS `json:"ret"`
}

func (r *sfgTmpResGetOsnNamesRet) LoadFromValByCopyWithJSON(val interface{}) (err error) {
	v, ok := val.(*sfgTmpResGetOsnNamesRet)
	if !ok {
		panic("[sfgTmpResGetOsnNamesRet]: 类型转换失败")
	}

	err = cutil.CopyStructUseJSON(r, v)

	return
}
