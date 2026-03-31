package funcs

import "context"

type Func interface {
	Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error)
}

type CallFrame struct {
	Name  string        `json:"name,omitempty" bson:"name,omitempty"`
	Args  []interface{} `json:"args,omitempty" bson:"args,omitempty"`
	NRets int           `json:"nrets,omitempty" bson:"nrets,omitempty"`
	Label string        `json:"label,omitempty" bson:"label,omitempty"`
	Title string        `json:"title,omitempty" bson:"title,omitempty"`
}

var (
	mathFuncs  = &Math{}
	cmpFuncs   = &compare{}
	logicFuncs = &Logic{}
	hashFuncs  = &Hash{}
	strFuncs   = &str{}
	arrayFuncs = &array{}
)

var BuiltIns = map[string]Func{
	"sub":    mathFuncs,
	"mul":    mathFuncs,
	"add":    mathFuncs,
	"div":    mathFuncs,
	"eq":     cmpFuncs,
	"ne":     cmpFuncs,
	"lt":     cmpFuncs,
	"lte":    cmpFuncs,
	"gt":     cmpFuncs,
	"gte":    cmpFuncs,
	"and":    logicFuncs,
	"or":     logicFuncs,
	"not":    logicFuncs,
	"get":    hashFuncs,
	"set":    hashFuncs,
	"str":    strFuncs,
	"len":    arrayFuncs,
	"array":  arrayFuncs,
	"append": arrayFuncs,
}
