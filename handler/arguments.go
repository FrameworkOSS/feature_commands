package handler

import (
	"fmt"
	"strings"

	"github.com/FrameworkOSS/portal/features/wires/wire"
	"github.com/FrameworkOSS/portal/portal"
	"github.com/JoshuaDoes/crunchio"
)

const (
	CommandArgTypeAuto CommandArgType = iota //Attempt to automatically detect when multiple argument types are compatible.
	CommandArgTypeRaw
	CommandArgTypeString
	CommandArgTypeNumber
	CommandArgTypeCurrency
	CommandArgTypePassword
	CommandArgTypeEmail
	CommandArgTypePhone
	CommandArgTypeShipping
	CommandArgTypeIPv4
	CommandArgTypeIPv6
)

// CommandArg defines the layout of a command argument.
type CommandArg struct {
	id          string           //Unique identifier for this argument within the command.
	name        string           //Display name for argument.
	about       string           //Description of what this argument does.
	usage       string           //Additional instructions on how to use this argument.
	aliases     []string         //Alternative names to reference this argument.
	required    bool             //True when argument must be specified to satisfy the command.
	requiresVal bool             //True when argument must be given a value to satisfy the command.
	repeatable  bool             //Whether or not this argument may be specified multiple times.
	valtype     CommandArgType   //Identifies a portal-recognized value type that this argument accepts.
	val         *crunchio.Buffer //The default value for this argument, if any.
}

// CommandArgType is an enumeration of the supported argument value types.
type CommandArgType int64

func (t CommandArgType) String() string {
	switch t {
	case CommandArgTypeRaw:
		return "raw"
	case CommandArgTypeString:
		return "string"
	case CommandArgTypeNumber:
		return "number"
	case CommandArgTypeCurrency:
		return "currency"
	case CommandArgTypePassword:
		return "password"
	case CommandArgTypeEmail:
		return "email"
	case CommandArgTypePhone:
		return "phone"
	case CommandArgTypeShipping:
		return "shipping"
	case CommandArgTypeIPv4:
		return "ip4"
	case CommandArgTypeIPv6:
		return "ip6"
	}
	return ""
}

func NewCommandArg() *CommandArg {
	arg := new(CommandArg)
	arg.aliases = make([]string, 0)
	return arg
}

func NewFeatureBindingArg(f portal.Feature, arg *CommandArg) (fb *portal.FeatureBinding) {
	w := wire.NewWire(arg.GetValueBytes())
	defer w.Close()

	id := w.GetValues(0)
	name := w.GetValues(1)
	authors := w.GetValues(2)
	description := w.GetValues(3)
	version := w.GetValues(4)

	fb = portal.NewFeatureBinding(f).
		SetID(string(id[0])).
		SetName(string(name[0])).
		SetDescription(string(description[0])).
		SetVersion(string(version[0]))

	strAuthors := make([]string, len(authors))
	for i := 0; i < len(authors); i++ {
		strAuthors[i] = string(authors[i])
	}
	fb.SetAuthors(strAuthors...)

	return
}

func NewCommandArgsEvent(e *portal.Event) (string, []*CommandArg, error) {
	if e == nil {
		return "", nil, nil
	}
	offsets := e.GetOffsets()
	/*if len(offsets) <= 1 {
		return "", nil, nil
	}*/
	args := make([]*CommandArg, 0)
	if offsets[uint64(len(offsets)-1)] >= uint64(e.GetDataSize()) {
		return "", nil, ErrorCommandArgInvalidOffsetExceedsLength(len(offsets)-1, e.GetDataSize())
	}

	//The first argument of an event is always the command being called!
	for i := 1; i < len(offsets); i++ {
		//fmt.Printf("OFFSET %d/%d: %d\n", i+1, len(offsets), offsets[i])
		b := e.GetArgument(i)
		//fmt.Printf("DATA %d/%d: %s\n", i+1, len(offsets), string(b))

		arg, err := NewCommandArgBytes(b)
		if err != nil {
			return "", nil, ErrorCommandArgInvalidData(i, offsets[i], len(b), err)
		}
		args = append(args, arg)
	}
	//fmt.Printf("%s produced %d args: %X\n", e.GetID(), len(args), args[0].GetValue().Bytes())
	return string(e.GetArgument(0)), args, nil
}

func NewCommandArgBytes(p []byte) (*CommandArg, error) {
	if len(p) > 0 {
		w := wire.NewWire(p)
		defer w.Close()

		id := w.GetValues(0)
		name := w.GetValues(1)
		about := w.GetValues(2)
		usage := w.GetValues(3)
		aliases := w.GetValues(4)
		required := w.GetValues(5)
		requiresVal := w.GetValues(6)
		repeatable := w.GetValues(7)
		valueType := w.GetValues(8)
		value := w.GetValues(9)

		arg := NewCommandArg()
		arg.SetID(string(id[0]))
		if len(name) > 0 {
			arg.SetName(string(name[0]))
		}
		if len(about) > 0 {
			arg.SetAbout(string(about[0]))
		}
		if len(usage) > 0 {
			arg.SetUsage(string(usage[0]))
		}

		for i := 0; i < len(aliases); i++ {
			arg.AddAliases(string(aliases[i]))
		}

		if len(required) > 0 && required[0][0] > 0 {
			arg.SetRequired(true)
		}
		if len(requiresVal) > 0 && requiresVal[0][0] > 0 {
			arg.SetRequiresValue(true)
		}
		if len(repeatable) > 0 && repeatable[0][0] > 0 {
			arg.SetRepeatable(true)
		}

		if len(value) > 0 {
			arg.SetValue(value[0])
		}

		if len(valueType) > 0 {
			arg.SetType(CommandArgType(wire.ReadIV64(valueType[0])))
		}

		return arg, nil
	}

	return nil, nil
}

/*	---
	--- GETTERS ---
	---
*/

func (arg *CommandArg) Bytes() []byte {
	w := wire.NewWire()
	defer w.Close()

	w.AddField(0, []byte(arg.GetID()))
	if name := arg.GetName(); name != "" {
		w.AddField(1, []byte(name))
	}
	if about := arg.GetAbout(); about != "" {
		w.AddField(2, []byte(about))
	}
	if usage := arg.GetUsage(); usage != "" {
		w.AddField(3, []byte(usage))
	}

	aliases := arg.GetAliases()
	for i := 0; i < len(aliases); i++ {
		w.AddField(4, []byte(aliases[i]))
	}

	if arg.IsRequired() {
		w.AddField(5, []byte{1})
	}
	if arg.IsValueRequired() {
		w.AddField(6, []byte{1})
	}
	if arg.IsRepeatable() {
		w.AddField(7, []byte{1})
	}

	if arg.GetType() != CommandArgTypeAuto {
		w.AddField(8, wire.WriteIV64(int64(arg.GetType())))
	}

	if val := arg.GetValue(); val != nil && val.Size() > 0 {
		w.AddField(9, val.Bytes())
	}

	return w.Bytes()
}

func (arg *CommandArg) GetID() string {
	return arg.id
}

func (arg *CommandArg) GetName() string {
	return arg.name
}

func (arg *CommandArg) GetAbout() string {
	return arg.about
}

func (arg *CommandArg) GetUsage() string {
	return arg.usage
}

func (arg *CommandArg) GetAliases() []string {
	return arg.aliases
}

func (arg *CommandArg) IsAlias(name string) bool {
	for _, alias := range arg.aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func (arg *CommandArg) IsRequired() bool {
	return arg.required
}

func (arg *CommandArg) IsValueRequired() bool {
	return arg.requiresVal
}

func (arg *CommandArg) IsRepeatable() bool {
	return arg.repeatable
}

func (arg *CommandArg) GetType() CommandArgType {
	return arg.valtype
}

func (arg *CommandArg) GetValueEmpty() bool {
	return arg.val == nil || arg.val.Size() == 0
}

func (arg *CommandArg) GetValue() *crunchio.Buffer {
	return arg.val
}

func (arg *CommandArg) GetValueBytes() []byte {
	if arg.GetValueEmpty() {
		return nil
	}
	p := arg.val.Bytes()
	arg.val.Seek(0, 0)
	return p
}

func (arg *CommandArg) GetValueNumber() int {
	if arg.GetValueEmpty() {
		return -1
	}
	i := wire.ReadIV64Next(arg.val.Buffer())
	arg.val.Seek(0, 0)
	return int(i)
}

func (arg *CommandArg) GetValueString() string {
	if arg.GetValueEmpty() {
		return ""
	}
	str := string(arg.val.Bytes())
	arg.val.Seek(0, 0)
	return str
}

func (arg *CommandArg) GetValueStringToLower() string {
	return strings.ToLower(arg.GetValueString())
}

/*	---
	--- SETTERS ---
	---
*/

func (arg *CommandArg) SetID(id string) *CommandArg {
	arg.id = id
	return arg
}

func (arg *CommandArg) SetName(name string) *CommandArg {
	arg.name = name
	return arg
}

func (arg *CommandArg) SetAbout(about string) *CommandArg {
	arg.about = about
	return arg
}

func (arg *CommandArg) SetUsage(usage string) *CommandArg {
	arg.usage = usage
	return arg
}

func (arg *CommandArg) AddAliases(aliases ...string) *CommandArg {
	if arg.aliases == nil {
		arg.aliases = make([]string, 0)
	}
	arg.aliases = append(arg.aliases, aliases...)
	return arg
}

func (arg *CommandArg) SetAliases(aliases ...string) *CommandArg {
	arg.aliases = aliases
	return arg
}

func (arg *CommandArg) SetRequired(required bool) *CommandArg {
	arg.required = required
	return arg
}

func (arg *CommandArg) SetRequiresValue(required bool) *CommandArg {
	arg.requiresVal = required
	return arg
}

func (arg *CommandArg) SetRepeatable(repeatable bool) *CommandArg {
	arg.repeatable = repeatable
	return arg
}

func (arg *CommandArg) SetType(valtype CommandArgType) *CommandArg {
	arg.valtype = valtype
	return arg
}

func (arg *CommandArg) SetValue(p []byte) *CommandArg {
	if arg.val != nil {
		arg.val.Close()
	}
	arg.val = crunchio.NewBuffer(arg.GetID(), p)
	arg.val.Seek(0, 0)
	arg.SetType(CommandArgTypeRaw)
	return arg
}

func (arg *CommandArg) SetValueNumber(num int) *CommandArg {
	if arg.val != nil {
		arg.val.Close()
	}
	arg.val = crunchio.NewBuffer(arg.GetID(), wire.WriteIV64(int64(num)))
	arg.val.Seek(0, 0)
	arg.SetType(CommandArgTypeNumber)
	return arg
}

func (arg *CommandArg) SetValueString(str string) *CommandArg {
	if arg.val != nil {
		arg.val.Close()
	}
	arg.val = crunchio.NewBuffer(arg.GetID(), make([]byte, len(str)))
	arg.val.Buffer().WriteBytesNext([]byte(str))
	arg.val.Seek(0, 0)
	arg.SetType(CommandArgTypeString)
	return arg
}

func (arg *CommandArg) Close() error {
	if arg.val != nil {
		return arg.val.Close()
	}
	return fmt.Errorf("arg: value already closed")
}
