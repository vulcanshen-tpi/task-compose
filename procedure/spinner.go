package procedure

import (
	"github.com/chelnak/ysmrr"
	"github.com/vulcanshen-tpi/task-compose/app"
)

type SpinnerAgent struct {
	sm       ysmrr.SpinnerManager
	spinners map[string]*ysmrr.Spinner
}

var (
	TaskSpinner SpinnerAgent
)

func InitializeSpinnerAgent() {
	if !app.DetachMode {
		return
	}
	var spinnerManager = ysmrr.NewSpinnerManager()
	TaskSpinner = SpinnerAgent{
		sm:       spinnerManager,
		spinners: make(map[string]*ysmrr.Spinner),
	}
}

func StartSpinnerAgent() {
	if !app.DetachMode {
		return
	}
	if TaskSpinner.sm != nil {
		TaskSpinner.sm.Start()
	}
}

func StopSpinnerAgent() {
	if !app.DetachMode {
		return
	}
	if TaskSpinner.sm != nil {
		TaskSpinner.sm.Stop()
	}
}

func (sa *SpinnerAgent) RegisterSpinner(name string, prefix string, defaultMessage string) {
	if !app.DetachMode {
		return
	}
	var spinner = sa.sm.AddSpinner(name)
	spinner.UpdatePrefix(prefix)
	spinner.UpdateMessage(defaultMessage)
	sa.spinners[name] = spinner
}

func (sa *SpinnerAgent) GetSpinner(name string) (*ysmrr.Spinner, bool) {
	if !app.DetachMode {
		return nil, false
	}
	var spinner, ok = sa.spinners[name]
	return spinner, ok
}
