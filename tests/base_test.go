package tests

import "github.com/stretchr/testify/assert"

func (cli *CLISuite) TestCmdMockerRun() {
	t := cli.T()
	t.Skip()
	_, err := cli.ExecuteCommand(
		"eip",
		"-c", "configs/config.json",
	)
	assert.NoError(t, err)
}
