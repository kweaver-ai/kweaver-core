package cenvhelper

type RunScenario string

const (
	RunScenario_Aaron_Local_Dev RunScenario = "aaron_local_dev"
)

func (rs RunScenario) String() string {
	return string(rs)
}
