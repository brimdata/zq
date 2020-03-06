package slice

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/brimsec/zq/cmd/pcap/root"
	"github.com/brimsec/zq/pcap/pcapio"
	"github.com/mccanne/charm"
)

var Ts = &charm.Spec{
	Name:  "ts",
	Usage: "ts [options] ts",
	Short: "print timestamps of a pcap",
	Long: `
The ts command prints the time stamps of each packet in the input pcap in
fractional seconds.  This is useful for testing.
`,
	New: New,
}

func init() {
	root.Pcap.Add(Ts)
}

type Command struct {
	inputFile  string
	outputFile string
	*root.Command
}

func New(parent charm.Command, f *flag.FlagSet) (charm.Command, error) {
	c := &Command{Command: parent.(*root.Command)}
	f.StringVar(&c.inputFile, "r", "-", "file to read from or stdin if -")
	f.StringVar(&c.outputFile, "w", "-", "file to write to or stdout if -")
	return c, nil
}

func (c *Command) Run(args []string) error {
	if len(args) != 0 {
		return errors.New("pcap ts takes no arguments")
	}
	in := os.Stdin
	if c.inputFile != "-" {
		var err error
		in, err = os.Open(c.inputFile)
		if err != nil {
			return err
		}
		defer in.Close()
	}
	// XXX assumes legacy pcap format
	// TBD: use generic packet reader here once we have the interface
	// and logic to chooose between NG and legacy
	reader, err := pcapio.NewPcapReader(in)
	if err != nil {
		return err
	}
	out := os.Stdout
	if c.outputFile != "-" {
		var err error
		out, err = os.OpenFile(c.outputFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer out.Close()
	}
	// skip header
	_, _, err = reader.Read()
	if err != nil {
		return err
	}
	for {
		block, info, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if block == nil {
			break
		}
		fmt.Fprintln(out, info.Ts.StringFloat())
	}
	return nil
}
