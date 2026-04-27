package hook

type Hook interface{}

type BeforeAssign interface {
	HookBeforeAssign(id string, target string, value any)
}

type BeforeReturn interface {
	HookBeforeReturn(id string, value any)
}

type LoopStart interface {
	HookLoopStart(id string, value any)
}

type LoopEnd interface {
	HookLoopEnd(id string, value any)
}

type BranchStart interface {
	HookBranchStart(id string)
}

type BranchSkip interface {
	HookBranchSkip(id string)
}

type Resume interface {
	HookResume(name, id string, value any)
}

type ResumeError interface {
	HookResumeError(name, id string, err any)
}

type VMStop interface {
	HookVMStop()
}
