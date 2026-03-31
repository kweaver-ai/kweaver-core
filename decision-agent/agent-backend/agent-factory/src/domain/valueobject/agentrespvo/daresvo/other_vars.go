package daresvo

func (r *DataAgentRes) GetOtherVarsMap() (m map[string]interface{}, err error) {
	m, err = r.otherVarsHelper.GetOtherVarsMap()

	return
}
