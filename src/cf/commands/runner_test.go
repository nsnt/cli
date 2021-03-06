package commands_test

import (
	. "cf/commands"
	"cf/requirements"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	"testing"
)

type TestCommandFactory struct {
	Cmd     Command
	CmdName string
}

func (f *TestCommandFactory) GetByCmdName(cmdName string) (cmd Command, err error) {
	f.CmdName = cmdName
	cmd = f.Cmd
	return
}

type TestCommand struct {
	Reqs       []requirements.Requirement
	WasRunWith *cli.Context
}

func (cmd *TestCommand) GetRequirements(factory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = cmd.Reqs
	return
}

func (cmd *TestCommand) Run(c *cli.Context) {
	cmd.WasRunWith = c
}

type TestRequirement struct {
	Passes      bool
	WasExecuted bool
}

func (r *TestRequirement) Execute() (success bool) {
	r.WasExecuted = true

	if !r.Passes {
		return false
	}

	return true
}

func TestRun(t *testing.T) {
	passingReq := TestRequirement{Passes: true}
	failingReq := TestRequirement{Passes: false}
	lastReq := TestRequirement{Passes: true}

	cmd := TestCommand{
		Reqs: []requirements.Requirement{&passingReq, &failingReq, &lastReq},
	}

	cmdFactory := &TestCommandFactory{Cmd: &cmd}
	runner := NewRunner(cmdFactory, nil)

	ctxt := testcmd.NewContext("login", []string{})

	err := runner.RunCmdByName("some-cmd", ctxt)

	assert.Equal(t, cmdFactory.CmdName, "some-cmd")

	assert.True(t, passingReq.WasExecuted, ctxt)
	assert.True(t, failingReq.WasExecuted, ctxt)

	assert.False(t, lastReq.WasExecuted)
	assert.Nil(t, cmd.WasRunWith)

	assert.Error(t, err)
}
