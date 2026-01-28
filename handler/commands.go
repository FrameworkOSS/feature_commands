package handler

import (
	"github.com/FrameworkOSS/portal/features/wires/wire"
	"github.com/FrameworkOSS/portal/portal"
)

type CommandWrapper struct {
	// The underlying command transport.
	Feature portal.Feature //The feature transport which will process this command.
	Command *Command       //The definition of the command to be processed.
	Handler func() error   //For native portal commands to mount their handlers.
}

// Command is an abstraction over the event transport to provide a structured protocol for features to communicate with.
type Command struct {
	id           string              //Unique identifier for this command within the portal.
	name         string              //Display name for command.
	about        string              //Description of what this command does.
	usage        string              //Detailed instructions (like manpages) on how to use this command.
	aliases      []string            //Alternative names to reference this command.
	requiresArgs bool                //Whether or not arguments are required before processing this command.
	preprocess   bool                //Whether or not the arguments specified (if any) must adhere to this command definition's argument mapping.
	args         []*CommandArg       //Arguments to provide as context to this command (where values are provided at runtime).
	subcmds      map[string]*Command //Nestable subcommands to provide structured heirarchy.
}

func NewCommandWrapper(f portal.Feature, c *Command) (*CommandWrapper, error) {
	if f == nil {
		return nil, ErrorFeatureNotFound("")
	}
	nc := new(CommandWrapper)
	nc.Feature = f
	nc.Command = c
	return nc, nil
}

func NewCommand() *Command {
	cmd := new(Command)
	cmd.aliases = make([]string, 0)
	cmd.args = make([]*CommandArg, 0)
	cmd.subcmds = make(map[string]*Command)
	return cmd
}

func NewCommandBytes(p []byte) (*Command, error) {
	if len(p) > 0 {
		w := wire.NewWire(p)
		defer w.Close()

		id := w.GetValues(0)
		name := w.GetValues(1)
		about := w.GetValues(2)
		usage := w.GetValues(3)
		aliases := w.GetValues(4)
		args := w.GetValues(5)
		subcmds := w.GetValues(6)
		requiresArgs := w.GetValues(7)
		requiresDefined := w.GetValues(8)

		cmd := NewCommand()
		cmd.SetID(string(id[0]))
		if len(name) > 0 {
			cmd.SetName(string(name[0]))
		}
		if len(about) > 0 {
			cmd.SetAbout(string(about[0]))
		}
		if len(usage) > 0 {
			cmd.SetUsage(string(usage[0]))
		}
		if len(requiresArgs) > 0 {
			cmd.SetRequiresArguments(true)
		}
		if len(requiresDefined) > 0 {
			cmd.SetRequiresPreprocessing(true)
		}

		for i := 0; i < len(aliases); i++ {
			cmd.AddAliases(string(aliases[i]))
		}

		for i := 0; i < len(args); i++ {
			arg, err := NewCommandArgBytes(args[i])
			if err != nil {
				return nil, err
			}
			if arg != nil {
				cmd.AddArguments(arg)
			}
		}

		for i := 0; i < len(subcmds); i++ {
			subcmd, err := NewCommandBytes(subcmds[i])
			if err != nil {
				return nil, err
			}
			if subcmd != nil {
				cmd.SetSubcommand(subcmd)
			}
		}

		return cmd, nil
	}
	return nil, nil
}

/*	---
	--- GETTERS ---
	---
*/

func (cmd *Command) Bytes() []byte {
	w := wire.NewWire()
	defer w.Close()

	w.AddField(0, []byte(cmd.GetID()))
	if name := cmd.GetName(); name != "" {
		w.AddField(1, []byte(name))
	}
	if about := cmd.GetAbout(); about != "" {
		w.AddField(2, []byte(about))
	}
	if usage := cmd.GetUsage(); usage != "" {
		w.AddField(3, []byte(usage))
	}

	aliases := cmd.GetAliases()
	for i := 0; i < len(aliases); i++ {
		w.AddField(4, []byte(aliases[i]))
	}

	args := cmd.GetArgumentsList()
	for i := 0; i < len(args); i++ {
		w.AddField(5, cmd.GetArgument(args[i]).Bytes())
	}

	subcmds := cmd.GetSubcommands()
	for i := 0; i < len(subcmds); i++ {
		w.AddField(6, cmd.GetSubcommand(subcmds[i]).Bytes())
	}

	if cmd.GetRequiresArguments() {
		w.AddField(7, []byte{1})
	}

	if cmd.GetRequiresPreprocessing() {
		w.AddField(8, []byte{1})
	}

	return w.Bytes()
}

func (cmd *Command) GetID() string {
	return cmd.id
}

func (cmd *Command) GetName() string {
	return cmd.name
}

func (cmd *Command) GetAbout() string {
	return cmd.about
}

func (cmd *Command) GetUsage() string {
	return cmd.usage
}

func (cmd *Command) GetAliases() []string {
	return cmd.aliases
}

func (cmd *Command) IsAlias(name string) bool {
	for _, alias := range cmd.aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func (cmd *Command) GetRequiresArguments() bool {
	return cmd.requiresArgs
}

func (cmd *Command) GetRequiresPreprocessing() bool {
	return cmd.preprocess
}

func (cmd *Command) GetArgument(argID string) *CommandArg {
	for _, arg := range cmd.args {
		if arg.GetID() == argID {
			return arg
		}
	}
	return nil
}

func (cmd *Command) GetArgumentsID(argID string) (args []*CommandArg) {
	args = make([]*CommandArg, 0)
	for _, arg := range cmd.args {
		if arg.GetID() == argID {
			args = append(args, arg)
		}
	}
	return
}

func (cmd *Command) GetArguments() []*CommandArg {
	return cmd.args
}

func (cmd *Command) GetArgumentsList() []string {
	args := make([]string, 0)
	for i := range cmd.args {
		args = append(args, cmd.args[i].GetID())
	}
	return args
}

func (cmd *Command) GetRequiredArguments() []string {
	args := make([]string, 0)
	for _, arg := range cmd.args {
		if arg.IsRequired() {
			args = append(args, arg.GetID())
		}
	}
	return args
}

func (cmd *Command) GetSubcommands() []string {
	subcmds := make([]string, 0)
	for _, subcmd := range cmd.subcmds {
		subcmds = append(subcmds, subcmd.GetID())
	}
	return subcmds
}

func (cmd *Command) GetSubcommand(cmdID string) *Command {
	if val := cmd.subcmds[cmdID]; val != nil {
		return val
	}
	return nil
}

/*	---
	--- SETTERS ---
	---
*/

func (cmd *Command) SetID(id string) *Command {
	cmd.id = id
	return cmd
}

func (cmd *Command) SetName(name string) *Command {
	cmd.name = name
	return cmd
}

func (cmd *Command) SetAbout(about string) *Command {
	cmd.about = about
	return cmd
}

func (cmd *Command) SetUsage(usage string) *Command {
	cmd.usage = usage
	return cmd
}

func (cmd *Command) AddAliases(aliases ...string) *Command {
	if cmd.aliases == nil {
		cmd.aliases = make([]string, 0)
	}
	for i := 0; i < len(aliases); i++ {
		if !cmd.IsAlias(aliases[i]) {
			cmd.aliases = append(cmd.aliases, aliases[i])
		}
	}
	return cmd
}

func (cmd *Command) SetAliases(aliases ...string) *Command {
	cmd.aliases = aliases
	return cmd
}

func (cmd *Command) SetRequiresArguments(requiresArgs bool) *Command {
	cmd.requiresArgs = requiresArgs
	return cmd
}

func (cmd *Command) SetRequiresPreprocessing(preprocess bool) *Command {
	cmd.preprocess = preprocess
	return cmd
}

func (cmd *Command) SetArgument(arg *CommandArg) *Command {
	if cmd.args == nil {
		cmd.args = make([]*CommandArg, 0)
	}

	for i := 0; i < len(cmd.args); i++ {
		if cmd.args[i].GetID() == arg.GetID() {
			cmd.args[i] = arg
			return cmd
		}
	}

	cmd.args = append(cmd.args, arg)
	return cmd
}

func (cmd *Command) AddArguments(args ...*CommandArg) *Command {
	cmd.args = append(cmd.args, args...)
	return cmd
}

func (cmd *Command) SetSubcommand(subcmd *Command) *Command {
	if cmd.subcmds == nil {
		cmd.subcmds = make(map[string]*Command)
	}
	cmd.subcmds[subcmd.GetID()] = subcmd
	return cmd
}

/*	---
	--- HELPERS ---
	---
*/

func (cmd *Command) ForEachArgument(method func(arg *CommandArg) error) error {
	args := cmd.GetArguments()
	for i := 0; i < len(args); i++ {
		if err := method(args[i]); err != nil {
			return err
		}
	}
	return nil
}
