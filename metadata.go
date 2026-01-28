package commands

import "github.com/FrameworkOSS/portal/features/commands/handler"

const (
	KEY_COMMANDS = "\x01"
	KEY_RESPONSE = "\x02"
)

/* --- PORTAL COMMANDS --- */
var (
	portalCmds = []*handler.Command{
		cmdExit,

		//Order preference: list, add, remove, open, close, read, write
		cmdChannelAdd, cmdChannelRemove,
		cmdCommands, cmdCommandAdd, cmdCommandRemove,
		cmdFeatures, cmdFeatureAdd, cmdFeatureRemove, cmdFeatureOpen, cmdFeatureClose,
	}
)

/* --- COMMANDS --- */
var (
	cmdExit = handler.NewCommand().
		SetID("exit").
		SetName("Exit").
		SetAbout("Tells the portal to start performing close routines.").
		SetUsage("Provide an optional exit code to return after close finishes.").
		SetAliases("close", "quit", "shutdown", "die").
		SetRequiresPreprocessing(true).
		SetArgument(handler.NewCommandArg().
			SetID("code").
			SetName("exit code").
			SetAbout("The optional exit code to return.").
			SetUsage("Provide a 0-255 number to return back to the portal's caller.").
			SetAliases("exit", "exitcode", "return").
			SetType(handler.CommandArgTypeNumber).
			SetRequiresValue(true),
		)
	cmdChannelAdd = handler.NewCommand().
			SetID("channel_add").
			SetName("Channel Add").
			SetAbout("Adds the caller to a channel.").
			SetUsage("Provide the channel to add your feature to.").
			SetAliases("chadd", "chanadd", "channeladd", "addch", "addchan", "addchannel").
			SetRequiresArguments(true).
			SetRequiresPreprocessing(true).
			SetArgument(handler.NewCommandArg().
				SetID("channel").
				SetName("channel").
				SetAbout("The channel to add.").
				SetUsage("Provide the channel to join into.").
				SetAliases("ch", "chan").
				SetType(handler.CommandArgTypeString).
				SetRequired(true).
				SetRequiresValue(true).
				SetRepeatable(true),
		)
	cmdChannelRemove = handler.NewCommand().
				SetID("channel_remove").
				SetName("Channel Remove").
				SetAbout("Removes the caller from a channel.").
				SetUsage("Provide the channel to remove your feature from.").
				SetAliases(
			"chrm", "chremove", "chdel", "chanrm", "chanremove", "chandel", "channelrm", "channelremove", "channeldel",
			"rmch", "removech", "delch", "rmchan", "removechan", "delchan", "rmchannel", "removechannel", "delchannel",
		).
		SetRequiresArguments(true).
		SetRequiresPreprocessing(true).
		SetArgument(handler.NewCommandArg().
			SetID("channel").
			SetName("channel").
			SetAbout("The channel to remove.").
			SetUsage("Provide the channel to depart from.").
			SetAliases("ch", "chan").
			SetType(handler.CommandArgTypeString).
			SetRequired(true).
			SetRequiresValue(true).
			SetRepeatable(true),
		)
	cmdCommands = handler.NewCommand().
			SetID("commands").
			SetName("Commands").
			SetAbout("Responds with a list of commands registered to the portal.").
			SetUsage("Optionally provide a format to use for output.").
			SetAliases("cs", "cmds", "coms", "comms").
			SetRequiresPreprocessing(true).
			SetArgument(cmdArgFormat).
			SetArgument(cmdArgExclude).
			SetArgument(cmdArgInclude)
	cmdCommandAdd = handler.NewCommand().
			SetID("command_add").
			SetName("Command Add").
			SetAbout("Adds a command to the portal.").
			SetUsage("Provide the command to add to the portal.").
			SetAliases("cmdadd", "commadd", "commandadd", "addcmd", "addcomm", "addcommand").
			SetRequiresArguments(true).
			SetRequiresPreprocessing(true).
			SetArgument(handler.NewCommandArg().
				SetID("command").
				SetName("command").
				SetAbout("The command to add.").
				SetUsage("Provide the command to join into.").
				SetAliases("cmd", "comm").
				SetType(handler.CommandArgTypeString).
				SetRequired(true).
				SetRequiresValue(true).
				SetRepeatable(true),
		)
	cmdCommandRemove = handler.NewCommand().
				SetID("command_remove").
				SetName("Command Remove").
				SetAbout("Removes the command from the portal.").
				SetUsage("Provide the command to remove from the portal.").
				SetAliases(
			"cmdrm", "cmdremove", "cmddel", "commrm", "commremove", "commdel", "commandrm", "commandremove", "commanddel",
			"rmcmd", "removecmd", "delcmd", "rmcomm", "removecomm", "delcomm", "rmcommand", "removecommand", "delcommand",
		).
		SetRequiresArguments(true).
		SetRequiresPreprocessing(true).
		SetArgument(handler.NewCommandArg().
			SetID("command").
			SetName("command").
			SetAbout("The command to remove.").
			SetUsage("Provide the command to depart from.").
			SetAliases("cmd", "comm").
			SetType(handler.CommandArgTypeString).
			SetRequired(true).
			SetRequiresValue(true).
			SetRepeatable(true),
		)
	cmdFeatures = handler.NewCommand().
			SetID("features").
			SetName("Features").
			SetAbout("Responds with a list of features registered to the portal.").
			SetUsage("Optionally provide a format to use for output.").
			SetAliases("fs", "flist", "feats").
			SetRequiresPreprocessing(true).
			SetArgument(cmdArgFormat).
			SetArgument(cmdArgExclude).
			SetArgument(cmdArgInclude)
	cmdFeatureAdd = handler.NewCommand().
			SetID("feature_add").
			SetName("Feature Add").
			SetAbout("Binds a feature to the caller.").
			SetUsage("Provide the feature to bind to your feature.").
			SetAliases("fadd", "featadd", "featureadd", "addf", "addfeat", "addfeature").
			SetRequiresArguments(true).
			SetRequiresPreprocessing(true).
			SetArgument(handler.NewCommandArg().
				SetID("feature").
				SetName("feature").
				SetAbout("The feature to add.").
				SetUsage("Provide the feature to bind to.").
				SetAliases("f", "feat", "feature").
				SetType(handler.CommandArgTypeRaw).
				SetRequired(true).
				SetRequiresValue(true).
				SetRepeatable(true),
		)
	cmdFeatureRemove = handler.NewCommand().
				SetID("feature_remove").
				SetName("Feature Remove").
				SetAbout("Removes a feature from the portal.").
				SetUsage("Provide the feature to remove from the portal.").
				SetAliases(
			"frm", "fremove", "fdel", "featrm", "featremove", "featdel", "featurerm", "featureremove", "featuredel",
			"rmf", "removef", "delf", "rmfeat", "removefeat", "delfeat", "rmfeature", "removefeature", "delfeature",
		).
		SetRequiresArguments(true).
		SetRequiresPreprocessing(true).
		SetArgument(handler.NewCommandArg().
			SetID("feature").
			SetName("feature").
			SetAbout("The feature to remove.").
			SetUsage("Provide the feature to depart from.").
			SetAliases("f", "feat", "feature").
			SetType(handler.CommandArgTypeString).
			SetRequired(true).
			SetRequiresValue(true).
			SetRepeatable(true),
		)
	cmdFeatureOpen = handler.NewCommand().
			SetID("feature_open").
			SetName("Feature Open").
			SetAbout("Opens a feature into the portal.").
			SetUsage("Provide the feature to open into the portal.").
			SetAliases(
			"fo", "featureopen", "featopen", "openfeature", "openfeat",
		).
		SetRequiresArguments(true).
		SetRequiresPreprocessing(true).
		SetArgument(handler.NewCommandArg().
			SetID("feature").
			SetName("feature").
			SetAbout("The feature to open.").
			SetUsage("Provide the feature to be opened.").
			SetAliases("f", "feat", "feature").
			SetType(handler.CommandArgTypeString).
			SetRequired(true).
			SetRequiresValue(true).
			SetRepeatable(true),
		)
	cmdFeatureClose = handler.NewCommand().
			SetID("feature_close").
			SetName("Feature Close").
			SetAbout("Closes a feature from the portal.").
			SetUsage("Provide the feature to close from the portal.").
			SetAliases(
			"fc", "featureclose", "featclose", "closefeature", "closefeat",
		).
		SetRequiresArguments(true).
		SetRequiresPreprocessing(true).
		SetArgument(handler.NewCommandArg().
			SetID("feature").
			SetName("feature").
			SetAbout("The feature to close.").
			SetUsage("Provide the feature to be closed.").
			SetAliases("f", "feat", "feature").
			SetType(handler.CommandArgTypeString).
			SetRequired(true).
			SetRequiresValue(true).
			SetRepeatable(true),
		)
)

/* --- SHARED COMMAND ARGUMENTS --- */
var (
	cmdArgFormat = handler.NewCommandArg().
			SetID("format").
			SetName("format").
			SetAbout("The format to describe the command list using.").
			SetUsage("Available formats: pretty (default), csv").
			SetAliases("f", "form", "style", "type").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true)
	cmdArgExclude = handler.NewCommandArg().
			SetID("exclude").
			SetName("exclude list").
			SetAbout("The entries to exclude from the listing.").
			SetUsage("Specify what should be excluded with a delimited list: (raw:0x00) , ; : |").
			SetAliases("e", "ex", "x", "excluded").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true).
			SetRepeatable(true)
	cmdArgInclude = handler.NewCommandArg().
			SetID("include").
			SetName("include list").
			SetAbout("The entries to include in the listing.").
			SetUsage("Specify what should be included with a delimited list: (raw:0x00) , ; : |").
			SetAliases("i", "in", "included").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true).
			SetRepeatable(true)
)
