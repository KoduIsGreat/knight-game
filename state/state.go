package state

type State interface {
	Update(dt float64)
	Physics(dt float64)
}

type MachineImpl struct {
	current State
	states  map[int]State
}

func NewMachine() *MachineImpl {
	return &MachineImpl{}
}
