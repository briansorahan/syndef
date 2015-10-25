package main

import (
	"flag"
	"fmt"
	"github.com/scgolang/sc"
	"os"
)

type syndef struct {
	command  string
	output   *string
	flagSets map[string]*flag.FlagSet
}

func newSyndef() *syndef {
	s := new(syndef)
	s.flagSets = make(map[string]*flag.FlagSet)
	s.flagSets["format"] = flag.NewFlagSet("format", flag.ExitOnError)
	s.flagSets["diff"] = flag.NewFlagSet("diff", flag.ExitOnError)
	s.output = s.flagSets["format"].String("output", "json", "output format")
	return s
}

// command returns the FlagSet for a particular command,
// or nil if this was an invalid command
func (self *syndef) flagsFor(command string) *flag.FlagSet {
	if flagSet, hasCommand := self.flagSets[command]; hasCommand {
		self.command = command
		return flagSet
	} else {
		return nil
	}
}

// usage prints a usage message on stderr
func (self *syndef) usage() {
	w, prog := os.Stderr, os.Args[0]
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "%s COMMAND [OPTIONS]\n\n", prog)
	fmt.Fprintf(w, "Commands:\n")
	for cmd, _ := range self.flagSets {
		fmt.Fprintf(w, "%s\n", cmd)
	}
}

// die prints an error message and kills the process
func (self *syndef) die(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}

// format runs the format command
func (self *syndef) format() error {
	r, err := os.Open(self.flagSets["format"].Arg(0))
	if err != nil {
		return err
	}
	d, err := sc.ReadSynthdef(r)
	if err != nil {
		return err
	}
	switch *self.output {
	case "json":
		err = d.WriteJSON(os.Stdout)
	case "dot":
		err = d.WriteGraph(os.Stdout)
	default:
		err = d.WriteXML(os.Stdout)
	}
	return err
}

// diff runs the diff command
func (self *syndef) diff() error {
	return nil
}

// run runs a command
func (self *syndef) run() error {
	switch self.command {
	case "format":
		return self.format()
	case "diff":
		return self.diff()
	}
	return nil
}

func main() {
	syndef := newSyndef()
	if len(os.Args) < 2 {
		syndef.usage()
		os.Exit(1)
	}
	// determine if it is a valid command
	var command *flag.FlagSet
	if command = syndef.flagsFor(os.Args[1]); command == nil {
		syndef.usage()
		os.Exit(1)
	}
	// parse cli flags for the command
	err := command.Parse(os.Args[2:])
	if err != nil {
		if err == flag.ErrHelp {
			syndef.usage()
			os.Exit(0)
		} else {
			syndef.die(err)
		}
	}
	// run the command
	err = syndef.run()
	if err != nil {
		syndef.die(err)
	}
}
