package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/scgolang/sc"
)

func main() {
	controller := newController()

	if len(os.Args) < 2 {
		controller.usage()
		os.Exit(1)
	}
	controller.command = os.Args[1]

	command := controller.flagSets[controller.command]

	if command == nil {
		controller.usage()
		os.Exit(1)
	}
	// parse cli flags for the command
	if err := command.Parse(os.Args[2:]); err != nil {
		if err == flag.ErrHelp {
			controller.usage()
			os.Exit(0)
		} else {
			controller.die(err)
		}
	}
	// run the command
	if err := controller.run(); err != nil {
		controller.die(err)
	}
}

// controller controls the behavior of the app
type controller struct {
	command  string
	output   *string
	flagSets map[string]*flag.FlagSet
}

func newController() *controller {
	c := &controller{}
	c.flagSets = make(map[string]*flag.FlagSet)
	c.flagSets["format"] = flag.NewFlagSet("format", flag.ExitOnError)
	c.flagSets["diff"] = flag.NewFlagSet("diff", flag.ExitOnError)
	c.output = c.flagSets["format"].String("output", "json", "output format")
	return c
}

// die prints an error message and kills the process
func (c *controller) die(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}

// diff runs the diff command
func (c *controller) diff() error {
	fset := c.flagSets["diff"]

	if expected, got := 2, len(fset.Args()); expected != got {
		return errors.Errorf("expected %d args, got %d", expected, got)
	}
	f1, err := os.Open(fset.Arg(0))
	if err != nil {
		return err
	}
	f2, err := os.Open(fset.Arg(1))
	if err != nil {
		return err
	}
	s1, err := sc.ReadSynthdef(f1)
	if err != nil {
		return err
	}
	s2, err := sc.ReadSynthdef(f2)
	if err != nil {
		return err
	}
	diffs, err := differ{}.do(s1, s2)
	if err != nil {
		return err
	}
	fmt.Printf("%-50s%-50s\n", fset.Arg(0), fset.Arg(1))
	for _, diff := range diffs {
		fmt.Printf("%-50s%-50s\n", diff[0], diff[1])
	}
	return nil
}

// format runs the format command
func (c *controller) format() error {
	r, err := os.Open(c.flagSets["format"].Arg(0))
	if err != nil {
		return err
	}
	d, err := sc.ReadSynthdef(r)
	if err != nil {
		return err
	}
	switch *c.output {
	case "json":
		return d.WriteJSON(os.Stdout)
	case "dot":
		return d.WriteGraph(os.Stdout)
	case "xml":
		return d.WriteXML(os.Stdout)
	default:
		return c.writeTree(os.Stdout, d)
	}
}

// run runs a command
func (c *controller) run() error {
	switch c.command {
	case "format":
		return c.format()
	case "diff":
		return c.diff()
	}
	return nil
}

// usage prints a usage message on stderr
func (c *controller) usage() {
	w, prog := os.Stderr, os.Args[0]
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "%s COMMAND [OPTIONS]\n\n", prog)
	fmt.Fprintf(w, "Commands:\n")
	for cmd, _ := range c.flagSets {
		fmt.Fprintf(w, "%s\n", cmd)
	}
}

func (c *controller) writeTree(w io.Writer, d *sc.Synthdef) error {
	buf := &bytes.Buffer{}

	if err := d.WriteJSON(buf); err != nil {
		return err
	}
	s := synthdef{}

	if err := json.NewDecoder(buf).Decode(&s); err != nil {
		return err
	}
	return nil
}
