package inspect

import (
	"errors"
	"flag"
	"os"

	"github.com/brimsec/zq/cli"
	"github.com/brimsec/zq/cmd/zst/root"
	"github.com/brimsec/zq/zbuf"
	"github.com/brimsec/zq/zio"
	"github.com/brimsec/zq/zng/resolver"
	"github.com/brimsec/zq/zst"
	"github.com/mccanne/charm"
	"golang.org/x/crypto/ssh/terminal"
)

var Read = &charm.Spec{
	Name:  "read",
	Usage: "read [flags] path",
	Short: "read a zst file and output as zng",
	Long: `
The read command reads columnar zst from
a zst storage objects (local files or s3 objects) and outputs
the reconstructed zng row data in the format of choice.

This command is most useful for test, debug, and demo as you can also
read zst objects with zq.
`,
	New: newCommand,
}

func init() {
	root.Zst.Add(Read)
}

type Command struct {
	*root.Command
	writerFlags zio.WriterFlags
	output      cli.OutputFlags
}

func newCommand(parent charm.Command, f *flag.FlagSet) (charm.Command, error) {
	c := &Command{Command: parent.(*root.Command)}
	c.writerFlags.SetFlags(f)
	c.output.SetFlags(f)
	return c, nil
}

func isTerminal(f *os.File) bool {
	return terminal.IsTerminal(int(f.Fd()))
}

func (c *Command) Run(args []string) error {
	defer c.Cleanup()
	if ok, err := c.Init(); !ok {
		return err
	}
	if len(args) != 1 {
		return errors.New("zst read: must be run with a single path argument")
	}
	path := args[0]
	reader, err := zst.NewReaderFromPath(resolver.NewContext(), path)
	if err != nil {
		return err
	}
	defer reader.Close()
	writerOpts := c.writerFlags.Options()
	if err := c.output.Init(&writerOpts); err != nil {
		return err
	}
	writer, err := c.output.Open(writerOpts)
	if err != nil {
		return err
	}
	if err := zbuf.Copy(writer, reader); err != nil {
		writer.Close()
		return err
	}
	return writer.Close()
}
