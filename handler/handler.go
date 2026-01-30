package handler

import (
	"strings"

	"github.com/FrameworkOSS/event"
)

type CommandHandlerMethod func(cmd *Command, e *event.Event) error

type CommandHandlerWrapper struct {
	command *Command
	handler CommandHandlerMethod
}

type CommandHandler struct {
	handlerDefault *CommandHandlerWrapper
	handlers       map[string]*CommandHandlerWrapper
}

func NewCommandHandler() (ch *CommandHandler) {
	ch = new(CommandHandler)
	ch.handlers = make(map[string]*CommandHandlerWrapper)
	return
}

// Process executes a matching handler for a command.
func (ch *CommandHandler) Process(cmd *Command, e *event.Event) error {
	op := strings.Split(cmd.GetID(), " ")
	call := op[0]
	if handler := ch.GetHandler(call); handler != nil {
		return handler(cmd, e)
	}
	if ch.handlerDefault != nil {
		return ch.handlerDefault.handler(cmd, e)
	}
	return ErrorCommandHandlerNoMatch(call)
}

// Handle assigns a Go function to receive the specified commands.
// The method receives one command at a time, but can be in any order and must be safe for multiple concurrent calls.
// The command definitions must be provided to scan for aliases to keep in sync.
// Specifying no commands here will assign the method to be the default command handler.
func (ch *CommandHandler) Handle(method CommandHandlerMethod, command ...*Command) *CommandHandler {
	if len(command) > 0 {
		for i := 0; i < len(command); i++ {
			cmd := command[i]
			if cmd == nil {
				panic("CommandHandler: Handle: command must never be nil")
			}
			wrapper := &CommandHandlerWrapper{command: cmd, handler: method}
			ch.handlers[cmd.GetID()] = wrapper
			for _, alias := range cmd.GetAliases() {
				ch.handlers[alias] = wrapper
			}
		}
	} else {
		ch.handlerDefault = &CommandHandlerWrapper{handler: method}
	}
	return ch
}

// Unhandle removes one or more handlers. Specifying no command IDs will remove the default command handler.
func (ch *CommandHandler) Unhandle(commandID ...string) *CommandHandler {
	if len(commandID) > 0 {
		for i := 0; i < len(commandID); i++ {
			id := commandID[i]
			if id == "" {
				panic("CommandHandler: Unhandle: commandID must never be empty string")
			}
			cmd, exists := ch.handlers[id]
			if !exists {
				continue
			}
			for _, alias := range cmd.command.GetAliases() {
				delete(ch.handlers, alias)
			}
			delete(ch.handlers, id)
		}
	} else {
		ch.handlerDefault = nil
	}
	return ch
}

// GetCommandIDs returns the list of handled command IDs.
func (ch *CommandHandler) GetCommandIDs() []string {
	commandIDs := make([]string, 0)
	for commandID, wrapper := range ch.handlers {
		if commandID == wrapper.command.GetID() {
			commandIDs = append(commandIDs, commandID)
		}
	}
	return commandIDs
}

// GetHandler returns a handler for a command ID.
func (ch *CommandHandler) GetHandler(commandID string) CommandHandlerMethod {
	handler, exists := ch.handlers[commandID]
	if exists {
		return handler.handler
	}
	return nil
}

// GetHandlersMapClone creates a clone of the handler mapping and returns the clone. It is up to the caller to dereference the clone.
func (ch *CommandHandler) GetHandlersMapClone() map[string]*CommandHandlerWrapper {
	handlers := make(map[string]*CommandHandlerWrapper)
	for commandID, handler := range ch.handlers {
		handlers[commandID] = handler
	}
	return handlers
}

// EventCommandHandler provides an EventHandler wrapper around a CommandHandler that will proxy incoming call events to it.
type EventCommandHandler struct {
	eh *event.EventHandler
	ch *CommandHandler
}

func NewEventCommandHandler() (ech *EventCommandHandler) {
	ech = new(EventCommandHandler)
	ech.eh = event.NewEventHandler().
		Handle(ech.processCall, "call")
	ech.ch = NewCommandHandler()
	return
}

func (ech *EventCommandHandler) Process(e *event.Event) error {
	return ech.eh.Process(e)
}

func (ech *EventCommandHandler) processCall(e *event.Event) error {
	call, args, err := NewCommandArgsEvent(e)
	if err != nil {
		return err
	}
	cmd := NewCommand().SetID(call).AddArguments(args...)
	return ech.ch.Process(cmd, e)
}

func (ech *EventCommandHandler) GetEventHandler() *event.EventHandler {
	return ech.eh
}

func (ech *EventCommandHandler) GetCommandHandler() *CommandHandler {
	return ech.ch
}
