package state

import (
	"sync"

	"github.com/grbit/post_bot/model"
)

type stateMap struct {
	States map[int64]*model.State
	sync.Mutex
}

var globalStates stateMap

func init() {
	globalStates.States = make(map[int64]*model.State)
}

func Get(chatID int64) *model.State {
	globalStates.Lock()
	defer globalStates.Unlock()

	if s, ok := globalStates.States[chatID]; ok {
		return s
	}

	s := model.State{SearchCmdWasPrevious: false}
	globalStates.States[chatID] = &s

	return &s
}
