package interf

import "context"

type IReqCheck interface {
	ReqCheck() (err error)
	ReqCheckWithCtx(ctx context.Context) (err error)
}

type IValObjCheck interface {
	ValObjCheck() (err error)
	ValObjCheckWithCtx(ctx context.Context) (err error)
}

type IEnumCheck interface {
	EnumCheck() (err error)
	EnumCheckWithCtx(ctx context.Context) (err error)
}
