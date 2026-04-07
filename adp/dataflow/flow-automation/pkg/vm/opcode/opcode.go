package opcode

type OpCode uint

const (
	Push       OpCode = 1
	Pop        OpCode = 2
	Load       OpCode = 3
	Save       OpCode = 4
	Del        OpCode = 5
	Jmp        OpCode = 6
	Jnz        OpCode = 7
	Call       OpCode = 8
	Dict       OpCode = 9
	List       OpCode = 10
	LoadGlobal OpCode = 11
	SaveGlobal OpCode = 12
	Return     OpCode = 13
	Mark       OpCode = 14
	LoopTrace  OpCode = 15
)

const (
	MARK_BEFORE_ASSIGN = "BEFORE_ASSIGN"
	MARK_BEFORE_RETURN = "BEFORE_RETURN"
	MARK_BRANCH_START  = "BRANCH_START"
	MARK_BRANCH_SKIP   = "BRANCH_SKIP"
	MARK_LOOP_START    = "LOOP_START"
	MARK_LOOP_END      = "LOOP_END"
)

type Instruction struct {
	OpCode OpCode `json:"o,omitempty" bson:"o,omitempty"`
	Name   string `json:"n,omitempty" bson:"n,omitempty"`
	Pos    int    `json:"p,omitempty" bson:"p,omitempty"`
	Size   int    `json:"s,omitempty" bson:"s,omitempty"`
	Value  any    `json:"v,omitempty" bson:"v,omitempty"`
}
