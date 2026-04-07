package daresvo

import "github.com/pkg/errors"

func (r *DataAgentRes) Effective() (effective bool) {
	return r.GetFinalAnswer() != nil
}

func (r *DataAgentRes) GetFinalAnswer() (answer interface{}) {
	answer = r.finalAnswerVarHelper.getSingleFieldVal()
	return
}

func (r *DataAgentRes) GetFinalAnswerJSON() (byt []byte, err error) {
	byt, err = r.finalAnswerVarHelper.GetSingleFieldJSON()
	if err != nil {
		err = errors.Wrapf(err, "GetFinalAnswerJSON error:%v,res: %s\n", err, string(byt))
	}

	return
}
