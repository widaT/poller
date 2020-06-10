package pollopt

type PollOpt uint8

const (
	Edge    PollOpt = 0b0001
	Level   PollOpt = 0b0010
	Oneshot PollOpt = 0b0100
	ALL     PollOpt = Edge | Level | Oneshot
)

func (p PollOpt) IsEdge() bool {
	return p.contains(Edge)
}
func (p PollOpt) IsLevel() bool {
	return p.contains(Level)
}
func (p PollOpt) IsOneshot() bool {
	return p.contains(Oneshot)
}

func (p PollOpt) contains(other PollOpt) bool {
	return (p & other) == other
}
