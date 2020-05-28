package post

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/brimsec/zq/cmd/zapi/cmd"
	"github.com/brimsec/zq/cmd/zapi/format"
	"github.com/brimsec/zq/pkg/display"
	"github.com/brimsec/zq/zqd/api"
	"github.com/mccanne/charm"
)

var Post = &charm.Spec{
	Name:  "post",
	Usage: "post [options] path...",
	Short: "post log file(s) to a space",
	New:   NewPost,
}

func init() {
	cmd.CLI.Add(Post)
}

type LogCommand struct {
	*cmd.Command
	force      bool
	bytesRead  int64
	bytesTotal int64
	start      time.Time
	done       bool
}

func NewPost(parent charm.Command, flags *flag.FlagSet) (charm.Command, error) {
	c := &LogCommand{Command: parent.(*cmd.Command)}
	flags.BoolVar(&c.force, "f", false, "create space if specified space does not exist")
	return c, nil
}

// Run lists all spaces in the current zqd host or if a parameter
// is provided (in glob style) lists the info about that space.
func (c *LogCommand) Run(args []string) (err error) {
	client := c.Client()
	if len(args) == 0 {
		return errors.New("path arg(s) required")
	}
	if c.force {
		sp, err := client.SpacePost(c.Context(), api.SpacePostRequest{Name: c.Spacename})
		if err != nil && err != api.ErrSpaceExists {
			return err
		}
		c.Spacename = sp.Name
	}
	var out io.Writer
	var dp *display.Display
	if !c.NoFancy {
		dp = display.New(c, time.Second)
		out = dp.Bypass()
		go dp.Run()
	} else {
		out = os.Stdout
	}
	id, err := c.SpaceID()
	if err != nil {
		return err
	}
	c.start = time.Now()
	stream, err := client.LogPost(c.Context(), id, api.LogPostRequest{Paths: args})
	if err != nil {
		return err
	}
	var payload interface{}
loop:
	for {
		payload, err = stream.Next()
		if err != nil {
			break loop
		}
		if payload == nil {
			break loop
		}
		switch typ := payload.(type) {
		case *api.LogPostWarning:
			fmt.Fprintf(out, "warning: %s\n", typ.Warning)
		case *api.TaskEnd:
			if typ.Error != nil {
				err = typ.Error
			}
			break loop
		case *api.LogPostStatus:
			atomic.StoreInt64(&c.bytesRead, typ.LogReadSize)
			atomic.StoreInt64(&c.bytesTotal, typ.LogTotalSize)
		}
	}
	if dp != nil {
		dp.Close()
	}
	if err != nil && c.Context().Err() != nil {
		fmt.Println("post aborted")
		os.Exit(1)
		return nil
	}
	if err == nil {
		read := atomic.LoadInt64(&c.bytesRead)
		fmt.Printf("posted %s in %v\n", format.Bytes(read), time.Since(c.start))
	}
	return err
}

func (c *LogCommand) Display(w io.Writer) bool {
	total := atomic.LoadInt64(&c.bytesTotal)
	if total == 0 {
		io.WriteString(w, "posting...\n")
		return true
	}
	read := atomic.LoadInt64(&c.bytesRead)
	percent := float64(read) / float64(total) * 100
	fmt.Fprintf(w, "%5.1f%% %s/%s\n", percent, format.Bytes(read), format.Bytes(total))
	return true
}
