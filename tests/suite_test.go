package tests

import (
	"bytes"
	"github.com/ucloud/migrate/cmd/migrate/root"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestCLISuiteRun(t *testing.T) {
	suite.Run(t, new(CLISuite))
}

type CLISuite struct {
	suite.Suite

	root    *cobra.Command
	workDir string
}

func (cli *CLISuite) SetupSuite() {
	t := cli.T()
	workdir, err := os.Getwd()
	assert.NoError(t, err)
	cli.workDir = workdir

	err = os.Chdir("..")
	assert.NoError(t, err)
	cli.root = root.NewCmd()
}

func (cli *CLISuite) TearDownSuite() {
	t := cli.T()
	err := os.Chdir(cli.workDir)
	assert.NoError(t, err)
}

// ExecuteCommand is used for test purpose.
func (cli *CLISuite) ExecuteCommand(args ...string) (output []byte, err error) {
	buf := new(bytes.Buffer)
	cli.root.SetOut(buf)
	cli.root.SetArgs(args)
	err = cli.root.Execute()
	return buf.Bytes(), err
}
