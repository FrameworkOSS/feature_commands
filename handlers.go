package commands

import (
	"fmt"
	"strings"

	"github.com/FrameworkOSS/event"
	"github.com/FrameworkOSS/feature"
	"github.com/FrameworkOSS/feature_commands/handler"
	"github.com/FrameworkOSS/feature_commands/metadata"
	"github.com/FrameworkOSS/portal"
)

func (c *Commands) error(sourceFeatureID string, errs ...error) {
	for i := 0; i < len(errs); i++ {
		err := errs[i]
		if err != nil {
			if err == portal.ErrorInternallyHandled {
				continue
			}
			c.respond(nil, event.NewEventError(sourceFeatureID, err))
		}
	}
}

func (c *Commands) handler() (ech *handler.EventCommandHandler) {
	ech = handler.NewEventCommandHandler()
	ech.GetEventHandler().
		Handle(c.handleEventReady, event.EVENT_READY)
	ech.GetCommandHandler().
		Handle(c.handleCommandDefault).
		Handle(c.handleCommandExit, metadata.CmdExit).
		Handle(c.handleCommandChannelAdd, metadata.CmdChannelAdd).
		Handle(c.handleCommandChannelRemove, metadata.CmdChannelRemove).
		Handle(c.handleCommandCommandList, metadata.CmdCommands).
		Handle(c.handleCommandCommandAdd, metadata.CmdCommandAdd).
		Handle(c.handleCommandCommandRemove, metadata.CmdCommandRemove).
		Handle(c.handleCommandFeatureList, metadata.CmdFeatures).
		Handle(c.handleCommandFeatureAdd, metadata.CmdFeatureAdd).
		Handle(c.handleCommandFeatureRemove, metadata.CmdFeatureRemove).
		Handle(c.handleCommandFeatureOpen, metadata.CmdFeatureOpen).
		Handle(c.handleCommandFeatureClose, metadata.CmdFeatureClose)
	return
}

// handleEventReady says hello to new features with the stock list of commands.
func (c *Commands) handleEventReady(e *event.Event) error {
	//c.respond(e, handler.NewEventCommandAdd(c.p.ID(), portalCmds...))
	return nil
}

/* --- HANDLERS: COMMANDS --- */

// handleCommandDefault processes commands not native to this framework.
func (c *Commands) handleCommandDefault(cmd *handler.Command, e *event.Event) error {
	ft := c.p.Feature(e.GetProducer())
	if ft == nil {
		//Create a feature binding when the producer is not found.
		if err := c.p.FeatureAdd(feature.NewFeatureBinding(c.p).SetID(e.GetProducer())); err != nil {
			return err
		}
		ft = c.p.Feature(e.GetProducer())
		//return fmt.Errorf("call to portal:%s: %v", c.ID(), ErrorFeatureNotFound(e.GetProducer()))
	}

	op := strings.Split(cmd.GetID(), " ")
	call := op[0]

	//Make sure the command exists first.
	command := c.Command(call)
	if command == nil {
		return handler.ErrorCommandNotFound(cmd.GetID())
	}
	depth := 1
	for depth < len(op) {
		command = command.GetSubcommand(op[depth])
		if command == nil {
			break
		}
		depth++
	}
	if command == nil {
		return handler.ErrorCommandNotFound(strings.Join(op[:depth], " "))
	}

	var err error
	args := cmd.GetArguments()

	//Make sure arguments are valid to the given command definition.
	if command.GetRequiresPreprocessing() {
		for i := 0; i < len(args); i++ {
			arg := args[i]
			test := command.GetArgument(arg.GetID())
			if test == nil {
				fmt.Println(1)
				err = handler.ErrorCommandArgCallField(arg.GetID())
				break
			}
			if arg.GetType() != test.GetType() {
				err = handler.ErrorCommandArgCallType(arg.GetID(), arg.GetType(), test.GetType())
				break
			}
		}
	}
	if err != nil {
		return err
	}

	//Make sure required arguments are specified, regardless of argument values being valid.
	if command.GetRequiresArguments() {
		missing := make([]string, 0)
		test := command.GetRequiredArguments()
		for i := 0; i < len(test); i++ {
			found := false
			for j := 0; j < len(args); j++ {
				arg := args[j]
				if arg.GetID() == test[i] {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, test[i])
			}
		}
		if len(missing) > 0 {
			if len(missing) == 1 {
				err = handler.ErrorCommandArgCallField(missing[0])
			} else {
				err = handler.ErrorCommandArgCallFields(missing...)
			}
		}
	}
	if err != nil {
		return err
	}

	//Execute the command directly!
	cmdWrapper := c.getCommand(call)
	if cmdWrapper == nil {
		return handler.ErrorCommandNotFound(call)
	}

	if cmdWrapper.Handler != nil {
		if err := cmdWrapper.Handler(); err != nil {
			return err
		}
	} else {
		if c.p.OptionsGet().GetStrictListen() {
			//Send the command to the target feature directly.
			f := cmdWrapper.Feature
			if err := f.Input(e); err != nil {
				return err
			}
			return portal.ErrorInternallyHandled
		}
	}

	return nil
}

func (c *Commands) handleCommandExit(cmd *handler.Command, e *event.Event) error {
	exitCode := 0
	if arg := cmd.GetArgument("code"); arg != nil {
		exitCode = arg.GetValueNumber()
	}
	go c.p.Exit(e.GetProducer(), exitCode)
	return nil
}

func (c *Commands) handleCommandChannelAdd(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("channel")
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		channel := arg.GetValueString()
		if channel == "" {
			return handler.ErrorCommandArgCallValue("channel")
		}
		c.p.ChannelAdd(e.GetProducer(), channel)
		return nil
	})
}

func (c *Commands) handleCommandChannelRemove(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("channel")
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		channel := arg.GetValueString()
		if channel == "" {
			return handler.ErrorCommandArgCallValue("channel")
		}
		c.p.ChannelRemove(e.GetProducer(), channel)
		return nil
	})
}

func (c *Commands) handleCommandCommandList(cmd *handler.Command, e *event.Event) error {
	output := "pretty"
	if format := cmd.GetArgument("format"); format != nil {
		f := format.GetValueStringToLower()
		switch f {
		case "pretty", "csv", "raw":
			output = f
		default:
			return handler.ErrorCommandListInvalidFormat(f)
		}
	}

	excluded, included, err := handler.CmdGetExcludedIncludedList(cmd)
	if err != nil {
		return err
	}

	r := event.NewEventResponse(c.ID(), nil)

	//Header
	switch output {
	case "csv":
		r.AddStringNext("id,name,about,usage,aliases(;),preprocess,requiredArgs(;),args(;),subcmds(id;)\n")
	}

	//Rebuild the command list to include all subcommands
	commands := c.Commands()
	cmds := make([]*handler.Command, 0)
	var foreach func(op string, cmd *handler.Command)
	foreach = func(op string, cmd *handler.Command) {
		c, err := handler.NewCommandBytes(cmd.Bytes())
		if err != nil {
			panic(fmt.Sprintf("commands: failed to clone command: %v", err))
		}
		c.SetID(op)
		cmds = append(cmds, c)
		if sc := cmd.GetSubcommands(); len(sc) > 0 {
			for i := 0; i < len(sc); i++ {
				scmd := cmd.GetSubcommand(sc[i])
				foreach(op+" "+scmd.GetID(), scmd)
			}
		}
	}
	for i := 0; i < len(commands); i++ {
		cmd := c.Command(commands[i])
		foreach(commands[i], cmd)
	}

	//Body
	for i := 0; i < len(cmds); i++ {
		c := cmds[i]

		if excluded != nil {
			if handler.ListContains(excluded, c.GetID()) {
				continue
			}
		} else if included != nil {
			if !handler.ListContains(included, c.GetID()) {
				continue
			}
		}

		switch output {
		case "csv":
			if i > 0 {
				r.AddStringNext("\n")
			}
			r.AddStringNext(fmt.Sprintf("%s,%s,%s,%s,%s,%t,%s,%s,%s",
				c.GetID(),
				c.GetName(),
				c.GetAbout(),
				c.GetUsage(),
				strings.Join(c.GetAliases(), ";"),
				c.GetRequiresPreprocessing(),
				strings.Join(c.GetRequiredArguments(), ";"),
				strings.Join(c.GetArgumentsList(), ";"),
				strings.Join(c.GetSubcommands(), ";"),
			))
		case "pretty":
			if i > 0 {
				r.AddStringNext("\n\n")
			}
			r.AddStringNext(fmt.Sprintf("Command: %s\n-Name: %s\n-About: %s\n-Usage: %s",
				c.GetID(),
				c.GetName(),
				c.GetAbout(),
				c.GetUsage(),
			))
			if n := c.GetAliases(); len(n) > 0 {
				r.AddStringNext("\nAliases: " + strings.Join(n, ", "))
			}
			if c.GetRequiresPreprocessing() {
				r.AddStringNext("\nPreprocess: true")
			}
			if ra := c.GetRequiredArguments(); len(ra) > 0 {
				r.AddStringNext("\nRequired: " + strings.Join(ra, ", "))
			}
			if a := c.GetArgumentsList(); len(a) > 0 {
				r.AddStringNext("\nArguments: " + strings.Join(a, ", "))
			}
			if s := c.GetSubcommands(); len(s) > 0 {
				r.AddStringNext("\nSubcommands: " + strings.Join(s, ", "))
			}
		case "raw":
			r.AddOffsetNext()
			r.AddDataNext(c.Bytes())
		}
	}

	return c.respond(e, r)
}

func (c *Commands) handleCommandCommandAdd(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("command")
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		command, err := handler.NewCommandBytes(arg.GetValueBytes())
		if err != nil {
			return err
		}
		return c.CommandAdd(e.GetProducer(), command)
	})
}

func (c *Commands) handleCommandCommandRemove(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("command")
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		command := arg.GetValueString()
		if command == "" {
			return handler.ErrorCommandArgCallValue("command")
		}
		return c.CommandRemove(e.GetProducer(), command)
	})
}

func (c *Commands) handleCommandFeatureList(cmd *handler.Command, e *event.Event) error {
	output := "pretty"
	if format := cmd.GetArgument("format"); format != nil {
		f := format.GetValueStringToLower()
		switch f {
		case "pretty", "csv", "raw":
			output = f
		default:
			return handler.ErrorFeatureListInvalidFormat(f)
		}
	}

	excluded, included, err := handler.CmdGetExcludedIncludedList(cmd)
	if err != nil {
		return err
	}

	r := event.NewEventResponse(c.ID(), nil)

	//Header
	switch output {
	case "csv":
		r.AddStringNext("api,id,name,authors(;),description,version,ready,binding\n")
	}

	//Body
	features := c.p.Features()
	for i := 0; i < len(features); i++ {
		f := c.p.Feature(features[i])

		if excluded != nil {
			if handler.ListContains(excluded, f.ID()) {
				continue
			}
		} else if included != nil {
			if !handler.ListContains(included, f.ID()) {
				continue
			}
		}

		bindingID := ""
		if binding, ok := f.(*feature.FeatureBinding); ok {
			bindingID = binding.GetBinding().ID()
		}

		switch output {
		case "csv":
			if i > 0 {
				r.AddStringNext("\n")
			}
			r.AddStringNext(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%t,%s",
				f.API(),
				f.ID(),
				f.Name(),
				strings.Join(f.Authors(), ";"),
				f.Description(),
				f.Version(),
				c.p.FeatureIsReady(f.ID()),
				bindingID,
			))
		case "pretty":
			if i > 0 {
				r.AddStringNext("\n\n")
			}
			r.AddStringNext(fmt.Sprintf("Feature: %s\n-Name: %s\n-API: %d\n-Authors: %s\n-Description: %s\n-Version: %s\n-Ready: %t",
				f.ID(),
				f.Name(),
				f.API(),
				strings.Join(f.Authors(), ", "),
				f.Description(),
				f.Version(),
				c.p.FeatureIsReady(f.ID()),
			))
			if bindingID != "" {
				r.AddStringNext(fmt.Sprintf("\n-Binding: %s", bindingID))
			}
		case "raw":
			r.AddOffsetNext()
			r.AddDataNext(feature.NewFeatureBinding(f).CloneBinding().Bytes())
		}
	}

	return c.respond(e, r)
}

func (c *Commands) handleCommandFeatureAdd(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("feature")
	}
	feature := c.p.Feature(e.GetProducer())
	if feature == nil {
		return handler.ErrorFeatureNotFound(e.GetProducer())
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		return c.p.FeatureAdd(handler.NewFeatureBindingArg(feature, arg))
	})
}

func (c *Commands) handleCommandFeatureRemove(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("feature")
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		return c.p.FeatureRemove(arg.GetValueString())
	})
}

func (c *Commands) handleCommandFeatureOpen(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("feature")
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		return c.p.FeatureOpen(arg.GetValueString())
	})
}

func (c *Commands) handleCommandFeatureClose(cmd *handler.Command, e *event.Event) error {
	if len(cmd.GetArguments()) == 0 {
		return handler.ErrorCommandArgCallField("feature")
	}
	return cmd.ForEachArgument(func(arg *handler.CommandArg) error {
		featureID := arg.GetValueString()
		errs, retry := c.p.FeatureClose(featureID)
		if len(errs) > 0 {
			c.error(featureID, errs...)
		}
		if retry {
			return c.respond(e, handler.NewEventFeatureCloseRetry(featureID))
		}
		return nil
	})
}
