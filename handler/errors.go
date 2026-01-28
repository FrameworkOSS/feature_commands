package handler

import "fmt"

var (
	ErrorListMultipleSplitters                  = fmt.Errorf("list: cannot use multiple splitters")
	ErrorListNoSplitters                        = fmt.Errorf("list: no splitter was found")
	ErrorFeatureListCannotCombineExcludeInclude = fmt.Errorf("features: cannot combine exclude and include lists!")
	ErrorCommandListCannotCombineExcludeInclude = fmt.Errorf("commands: cannot combine exclude and include lists!")
)

// ErrorFeatureNotFound indicates that a feature with the given ID was not found in the portal network.
func ErrorFeatureNotFound(featureID string) error {
	return fmt.Errorf("feature: not found: %s", featureID)
}

// ErrorCommandNotFound indicates that a command with the given ID was not found in the portal network.
func ErrorCommandNotFound(commandID string) error {
	return fmt.Errorf("command: not found: %s", commandID)
}

// ErrorCommandExists indicates that a command with the given ID already exists somewhere in the portal network.
func ErrorCommandExists(commandID string) error {
	return fmt.Errorf("command: already exists: %s", commandID)
}

// ErrorCommandRegistrationField indicates that a command registration failed due to a missing or an invalid field.
func ErrorCommandRegistrationField(field string) error {
	return fmt.Errorf("command: registration failed: missing or invalid field: %s", field)
}

// ErrorCommandArgRegistrationField indicates that a command argument registration failed due to a missing or an invalid field.
func ErrorCommandArgRegistrationField(field string) error {
	return fmt.Errorf("command: argument: registration failed: missing or invalid field: %s", field)
}

// ErrorCommandArgCallField indicates that a command argument was not found when processing a command call.
func ErrorCommandArgCallField(field string) error {
	return fmt.Errorf("command: argument: call failed: missing or invalid field: %s", field)
}

// ErrorCommandArgCallFields indicates that multiple command arguments were not found when processing a command call.
func ErrorCommandArgCallFields(fields ...string) error {
	return fmt.Errorf("command: argument: call failed: missing or invalid fields: %v", fields)
}

// ErrorCommandArgCallType indicates that a command argument provided an invalid type expectation when processing a command call.
func ErrorCommandArgCallType(field string, found, expected CommandArgType) error {
	return fmt.Errorf("command: argument: call failed: invalid type for field: %s: found %s, expected %s", field, found, expected)
}

// ErrorCommandArgCallValue indicates that a command argument was provided an inappropriate value when processing a command call.
func ErrorCommandArgCallValue(field string) error {
	return fmt.Errorf("command: argument: call failed: invalid value for field: %s", field)
}

// ErrorCommandArgInvalidOffsetExceedsLength indicates that a command argument specified an offset which exceeds the data length.
func ErrorCommandArgInvalidOffsetExceedsLength(offset, length int) error {
	return fmt.Errorf("command: argument: event: invalid offset: offset (%d) exceeds data length (%d)", offset, length)
}

// ErrorCommandArgInvalidData indicates that a command argument has invalid data and could not be interpreted.
func ErrorCommandArgInvalidData(arg int, offset uint64, length int, err error) error {
	return fmt.Errorf("command: argument: event: invalid data: argument index (%d) has invalid data at offset %d with length %d: %v", arg, offset, length, err)
}

func ErrorCommandHandlerNoMatch(commandID string) error {
	return fmt.Errorf("CommandHandler: no match: %s", commandID)
}

func ErrorFeatureListInvalidFormat(format string) error {
	return fmt.Errorf("features: invalid format for listing: %s", format)
}

func ErrorCommandListInvalidFormat(format string) error {
	return fmt.Errorf("commands: invalid format for listing: %s", format)
}
