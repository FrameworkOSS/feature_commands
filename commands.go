package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/FrameworkOSS/portal/features/commands/handler"
	"github.com/FrameworkOSS/portal/portal"
)

type Commands struct {
	*portal.PortalMutex
	p         *portal.Portal
	processor *handler.EventCommandHandler
	resps     []*portal.Event
	hello     *portal.Event                      //Reconstructed each time the command list is updated!
	commands  map[string]*handler.CommandWrapper //command:Command (with a pointer to its Feature)
}

func NewCommands(p *portal.Portal) (c *Commands) {
	c = new(Commands)
	c.PortalMutex = portal.NewPortalMutex()
	c.processor = c.handler()
	c.p = p
	_, _ = c.Close()
	return
}

func (c *Commands) API() int {
	return 0
}

func (c *Commands) ID() string {
	return "commands"
}

func (c *Commands) Name() string {
	return "Commands"
}

func (c *Commands) Authors() []string {
	return []string{"JoshuaDoes"}
}

func (c *Commands) Description() string {
	return "Provides a call and response interface over the event protocol with the mundane concepts of commands, subcommands and their arguments."
}

func (c *Commands) Version() string {
	return "v0.0.1"
}

func (c *Commands) Open() error {
	/*if fs := c.p.Features(); len(fs) > 0 {
		for i := range fs {
			c.send(handler.NewEventFeatureAdd(c.p.ID(), c.p.Feature(fs[i])))
		}
	}*/
	if err := c.BatchCommandAdd(c.p.ID(), portalCmds...); err != nil {
		return err
	}
	c.send(portal.NewEventReady(c.ID(), true))
	return nil
}

func (c *Commands) Close() (errs []error, retry bool) {
	if cmds := c.Commands(); len(cmds) > 0 {
		if err := c.BatchCommandRemove(c.ID(), cmds...); err != nil {
			errs = []error{err}
			retry = true
		}
	}
	c.commands = make(map[string]*handler.CommandWrapper)
	return
}

func (c *Commands) Input(req *portal.Event) error {
	return c.processor.Process(req)
}

func (c *Commands) Output() (resp *portal.Event, err error) {
	if len(c.resps) > 0 {
		c.LockKey(KEY_RESPONSE)
		resp = c.resps[0]
		c.resps = c.resps[1:]
		c.UnlockKey(KEY_RESPONSE)
	}
	return
}

func (c *Commands) send(req *portal.Event) error {
	return c.respond(nil, req)
}

func (c *Commands) respond(ctx, resp *portal.Event) error {
	if resp != nil {
		if ctx != nil {
			resp.SetChannel(ctx.GetChannel()).SetParticipants(ctx.GetParticipants()...).AddParticipants(ctx.GetProducer())
		}
		c.LockKey(KEY_RESPONSE)
		c.resps = append(c.resps, resp)
		c.UnlockKey(KEY_RESPONSE)
	}
	return nil
}

func (c *Commands) NewCommandLineEvent(feature string, op ...string) (*portal.Event, error) {
	if len(op) == 0 {
		return nil, handler.ErrorCommandNotFound("")
	}
	if c.p == nil {
		return nil, handler.ErrorFeatureNotFound("portal")
	}

	cmd := op[0]
	command := c.Command(cmd)
	if command == nil {
		cmds := c.Commands()
		match, distance := LevenshteinDistance(cmd, cmds...)
		msg := fmt.Sprintf("unknown command: %s", cmd)
		if distance < 2.0 {
			msg += fmt.Sprintf(", did you mean?: %s", match)
		}
		return nil, fmt.Errorf("%s", msg)
	}
	wrapper := c.getCommand(cmd)

	op = op[1:]
	for len(op) > 0 {
		if op[0][0] == '-' {
			break
		}
		subcmd := command.GetSubcommand(op[0])
		cmd = cmd + " " + op[0]
		if subcmd == nil {
			scmds := command.GetSubcommands()
			match, distance := LevenshteinDistance(op[0], scmds...)
			msg := fmt.Sprintf("unknown subcommand: %s", op[0])
			if distance < 2.0 {
				msg += fmt.Sprintf(", did you mean?: %s", match)
			}
			return nil, fmt.Errorf("%s", msg)
		}
		op = op[1:]
		command = subcmd
	}

	//Check to see if this is just a passthrough command.
	args := make([]*handler.CommandArg, 0)
	if test := command.GetArguments(); len(test) == 0 && !command.GetRequiresPreprocessing() {
		for i := 0; i < len(op); i++ {
			args = append(args, handler.NewCommandArg().SetValueString(op[i]))
		}
	}

	if len(args) == 0 {
		inArg := false
		inArgName := ""
		names := make([]string, 0)
		values := make([]string, 0)
		for i := 0; i < len(op); i++ {
			str := op[i]
			if str[0] == '-' {
				if len(str) == 1 {
					return nil, fmt.Errorf("pipes not yet supported!")
				}
				if inArg {
					names = append(names, inArgName)
					values = append(values, "")
				}
				inArg = true
				inArgName = str[1:]
			} else {
				if !inArg {
					return nil, fmt.Errorf("must provide argument name for value: %s", str)
				}
				names = append(names, inArgName)
				values = append(values, str)
				inArg = false
			}
		}
		if inArg {
			names = append(names, inArgName)
			values = append(values, "")
		}
		if len(names) != len(values) {
			return nil, fmt.Errorf("mismatch names and values!")
		}

		for i := 0; i < len(names); i++ {
			name := names[i]
			arg := command.GetArgument(name)
			if arg == nil {
				return nil, fmt.Errorf("invalid argument: %s", name)
			}
			if i > 0 && !arg.IsRepeatable() {
				for j := i - 1; j >= 0; j-- {
					if args[j].GetID() == name {
						return nil, fmt.Errorf("not allowed to repeat argument: %s", name)
					}
				}
			}
			val := values[i]
			if arg.IsValueRequired() && val == "" {
				return nil, fmt.Errorf("missing value: %s", name)
			}
			if val != "" {
				switch arg.GetType() {
				case handler.CommandArgTypeRaw:
					arg.SetValue([]byte(val))
				case handler.CommandArgTypeString:
					arg.SetValueString(val)
				case handler.CommandArgTypeNumber:
					i, err := strconv.Atoi(val)
					if err != nil {
						return nil, fmt.Errorf("invalid number for argument: %s: %s", name, val)
					}
					arg.SetValueNumber(i)
				}
			}
			args = append(args, arg)
		}
	}

	e := handler.NewEventCall(feature, cmd, args...)

	if wrapper.Feature != nil {
		e.AddParticipants(wrapper.Feature.ID())
	}

	return e, nil
}

func (c *Commands) Command(commandID string) (cmd *handler.Command) {
	if w := c.getCommand(commandID); w != nil {
		cmd = w.Command
	}
	return
}

func (c *Commands) getCommand(commandID string) *handler.CommandWrapper {
	c.LockKey(KEY_COMMANDS)
	w := c.commands[commandID]
	c.UnlockKey(KEY_COMMANDS)

	return w
}

func (c *Commands) Commands() []string {
	c.LockKey(KEY_COMMANDS)
	commands := make([]string, 0)
	for k := range c.commands {
		commands = append(commands, k)
	}
	c.UnlockKey(KEY_COMMANDS)

	sort.Strings(commands)

	return commands
}

func (c *Commands) CommandAdd(featureID string, cmd *handler.Command) error {
	c.p.Log("CommandAdd(feature = %s, command = %s)", featureID, cmd.GetID())

	if strings.Contains(cmd.GetID(), " ") {
		return handler.ErrorCommandRegistrationField("id")
	}

	if commandID := cmd.GetID(); c.Command(commandID) != nil {
		for _, cmd := range portalCmds {
			if commandID == cmd.GetID() {
				return nil //Silently ignore attempts to replace a portal command.
			}
		}
		return handler.ErrorCommandExists(commandID)
	}

	f := c.p.Feature(featureID)
	if f == nil {
		return handler.ErrorFeatureNotFound(featureID)
	}

	nc, err := handler.NewCommandWrapper(f, cmd)
	if err != nil {
		return err
	}

	c.LockKey(KEY_COMMANDS)
	c.commands[cmd.GetID()] = nc
	c.UnlockKey(KEY_COMMANDS)

	c.send(handler.NewEventCommandAdd(featureID, cmd))

	return nil
}

func (c *Commands) CommandRemove(callerFeatureID, commandID string) error {
	if c.Command(commandID) == nil {
		return handler.ErrorCommandNotFound(commandID)
	}

	c.p.Log("CommandRemove(command = %s)", commandID)

	c.LockKey(KEY_COMMANDS)
	delete(c.commands, commandID)
	c.UnlockKey(KEY_COMMANDS)

	c.send(handler.NewEventCommandRemove(callerFeatureID, commandID))

	return nil
}

func (c *Commands) BatchCommandAdd(featureID string, commands ...*handler.Command) error {
	for i := 0; i < len(commands); i++ {
		if err := c.CommandAdd(featureID, commands[i]); err != nil {
			for j := 0; j < i; j++ {
				c.CommandRemove(featureID, commands[j].GetID())
			}
			return err
		}
	}
	return nil
}

func (c *Commands) BatchCommandRemove(callerFeatureID string, commands ...string) error {
	for i := 0; i < len(commands); i++ {
		if err := c.CommandRemove(callerFeatureID, commands[i]); err != nil {
			return err
		}
	}
	return nil
}

func Find(p *portal.Portal) (c *Commands) {
	if p == nil {
		return
	}
	if f := p.Feature("commands"); f != nil {
		if found, ok := f.(*Commands); ok {
			c = found
		}
		if f, ok := f.(*portal.FeatureBinding); ok {
			if bind := f.GetBinding(); bind != nil {
				if f, ok := bind.(*Commands); ok {
					c = f
				}
			}
		}
	}
	return
}
