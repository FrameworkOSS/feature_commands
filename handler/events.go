package handler

import (
	"github.com/FrameworkOSS/event"
	"github.com/FrameworkOSS/feature"
)

func NewEventCall(feature, command string, args ...*CommandArg) (e *event.Event) {
	e = event.NewEvent().
		SetID("call").
		SetProducer(feature)

	offsets := make([]uint64, len(args)+1)
	offsets[0] = 0
	data := []byte(command)
	for i := 0; i < len(args); i++ {
		argBytes := args[i].Bytes()
		offsets[i+1] = uint64(len(data))
		data = append(data, argBytes...)
	}
	e.SetOffsets(offsets...)
	e.SetData(data)
	return
}

func NewEventCommandAdd(feature string, command ...*Command) (e *event.Event) {
	args := make([]*CommandArg, len(command))
	for i := 0; i < len(args); i++ {
		args[i] = NewCommandArg().SetID("command").SetValue(command[i].Bytes())
	}
	e = NewEventCall(feature, "command_add", args...)
	for i := 0; i < len(args); i++ {
		args[i].Close()
	}
	return
}

func NewEventCommandRemove(feature string, commandID ...string) (e *event.Event) {
	args := make([]*CommandArg, len(commandID))
	for i := 0; i < len(args); i++ {
		args[i] = NewCommandArg().SetID("command").SetValueString(commandID[i])
	}
	e = NewEventCall(feature, "command_remove", args...)
	return
}

func NewEventChannelAdd(feature string, channel ...string) (e *event.Event) {
	args := make([]*CommandArg, len(channel))
	for i := 0; i < len(args); i++ {
		args[i] = NewCommandArg().SetID("channel").SetValueString(channel[i])
	}
	e = NewEventCall(feature, "channel_add", args...)
	return
}

func NewEventChannelRemove(feature string, channel ...string) (e *event.Event) {
	args := make([]*CommandArg, len(channel))
	for i := 0; i < len(args); i++ {
		args[i] = NewCommandArg().SetID("channel").SetValueString(channel[i])
	}
	e = NewEventCall(feature, "channel_remove", args...)
	return
}

func NewEventFeatureAdd(ft string, binding ...feature.Feature) (e *event.Event) {
	args := make([]*CommandArg, len(binding))
	for i := 0; i < len(binding); i++ {
		binder := binding[i]
		args[i] = NewCommandArg().SetID("feature").SetValue(
			feature.NewFeatureBinding(nil).
				SetID(binder.ID()).
				SetName(binder.Name()).
				SetAuthors(binder.Authors()...).
				SetDescription(binder.Description()).
				SetVersion(binder.Version()).Bytes(),
		)
	}
	e = NewEventCall(ft, "feature_add", args...)
	return
}

func NewEventFeatureRemove(feature string, binding ...string) (e *event.Event) {
	args := make([]*CommandArg, len(binding))
	for i := 0; i < len(binding); i++ {
		args[i] = NewCommandArg().SetID("feature").SetValueString(binding[i])
	}
	e = NewEventCall(feature, "feature_remove", args...)
	return
}

func NewEventFeatureOpen(feature string, binding ...string) (e *event.Event) {
	args := make([]*CommandArg, len(binding))
	for i := 0; i < len(binding); i++ {
		args[i] = NewCommandArg().SetID("feature").SetValueString(binding[i])
	}
	e = NewEventCall(feature, "feature_open", args...)
	return
}

func NewEventFeatureClose(feature string, binding ...string) (e *event.Event) {
	args := make([]*CommandArg, len(binding))
	for i := 0; i < len(binding); i++ {
		args[i] = NewCommandArg().SetID("feature").SetValueString(binding[i])
	}
	e = NewEventCall(feature, "feature_close", args...)
	return
}

func NewEventFeatureCloseRetry(feature string) *event.Event {
	return event.NewEvent().
		SetID("retry_close").
		SetProducer(feature)
}
