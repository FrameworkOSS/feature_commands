package handler

import "strings"

func CmdGetExcludedIncludedList(cmd *Command) (excluded, included []string, ret error) {
	exclude := cmd.GetArgumentsID("exclude")
	include := cmd.GetArgumentsID("include")
	if countBool(true, len(exclude) > 0, len(include) > 0) == 2 {
		ret = ErrorFeatureListCannotCombineExcludeInclude
		return
	}

	if len(exclude) > 0 {
		for i := range exclude {
			list, err := ListSplitFromArg(exclude[i])
			if err != nil && err != ErrorListNoSplitters {
				ret = err
				return
			}
			if excluded == nil {
				excluded = make([]string, 0)
			}
			excluded = append(excluded, list...)
		}
	} else if len(include) > 0 {
		for i := range include {
			list, err := ListSplitFromArg(include[i])
			if err != nil && err != ErrorListNoSplitters {
				ret = err
				return
			}
			if included == nil {
				included = make([]string, 0)
			}
			included = append(included, list...)
		}
	}

	return
}

func ListSplitFromArg(arg *CommandArg) ([]string, error) {
	str := arg.GetValueString()
	if str == "" {
		return nil, nil
	}

	hasNull := strings.Contains(str, "\x00")
	hasComma := strings.Contains(str, ",")
	hasSemicolon := strings.Contains(str, ";")
	hasColon := strings.Contains(str, ":")
	hasPipe := strings.Contains(str, "|")
	count := countBool(true, hasNull, hasComma, hasSemicolon, hasColon, hasPipe)
	if count > 1 {
		return nil, ErrorListMultipleSplitters
	}
	if count == 0 {
		return []string{str}, ErrorListNoSplitters
	}

	char := "\x00" //Default to null so we don't pass an empty delimiter to strings.Split
	if hasComma {
		char = ","
	} else if hasSemicolon {
		char = ";"
	} else if hasColon {
		char = ":"
	} else if hasPipe {
		char = "|"
	}

	return strings.Split(str, char), nil
}

func ListContains(list []string, contained ...string) bool {
	for i := range list {
		for j := range contained {
			if list[i] == contained[j] {
				return true
			}
		}
	}
	return false
}

func countBool(flag bool, bools ...bool) int {
	counter := 0
	for i := range bools {
		if bools[i] == flag {
			counter++
		}
	}
	return counter
}
