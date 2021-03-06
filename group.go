// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"errors"
	"fmt"
	"reflect"
	"unicode/utf8"
)

// The provided container is not a pointer to a struct
var ErrNotPointerToStruct = errors.New("provided data is not a pointer to struct")

// The provided short name is longer than a single character
var ErrShortNameTooLong = errors.New("short names can only be 1 character")

// Option flag information. Contains a description of the option, short and
// long name as well as a default value and whether an argument for this
// flag is optional.
type Option struct {
	// The short name of the option (a single character). If not 0, the
	// option flag can be 'activated' using -<ShortName>. Either ShortName
	// or LongName needs to be non-empty.
	ShortName rune

	// The long name of the option. If not "", the option flag can be
	// activated using --<LongName>. Either ShortName or LongName needs
	// to be non-empty.
	LongName string

	// The description of the option flag. This description is shown
	// automatically in the builtin help.
	Description string

	// The default value of the option. The default value is used when
	// the option flag is marked as having an OptionalArgument. This means
	// that when the flag is specified, but no option argument is given,
	// the value of the field this option represents will be set to
	// Default. This is only valid for non-boolean options.
	Default string

	// If true, specifies that the argument to an option flag is optional.
	// When no argument to the flag is specified on the command line, the
	// value of Default will be set in the field this option represents.
	// This is only valid for non-boolean options.
	OptionalArgument bool

	value   reflect.Value
	options reflect.StructTag
}

// An option group. The option group has a name and a set of options.
type Group struct {
	// The name of the group.
	Name string

	// A map of long names to option option descriptions.
	LongNames map[string]*Option

	// A map of short names to option option descriptions.
	ShortNames map[rune]*Option

	// A list of all the options in the group.
	Options []*Option

	// An error which occurred when creating the group.
	Error error

	data interface{}
}

// Set the value of an option to the specified value. An error will be returned
// if the specified value could not be converted to the corresponding option
// value type.
func (option *Option) Set(value *string) error {
	if option.isFunc() {
		return option.call(value)
	} else if value != nil {
		return convert(*value, option.value, option.options)
	} else {
		return convert("", option.value, option.options)
	}

	return nil
}

// Convert an option to a human friendly readable string describing the option.
func (option *Option) String() string {
	var s string
	var short string

	if option.ShortName != 0 {
		data := make([]byte, utf8.RuneLen(option.ShortName))
		utf8.EncodeRune(data, option.ShortName)
		short = string(data)

		if len(option.LongName) != 0 {
			s = fmt.Sprintf("-%s, --%s", short, option.LongName)
		} else {
			s = fmt.Sprintf("-%s", short)
		}
	} else if len(option.LongName) != 0 {
		s = fmt.Sprintf("--%s", option.LongName)
	}

	return s
}

// NewGroup creates a new option group with a given name and underlying data
// container. The data container is a pointer to a struct. The fields of the
// struct represent the command line options (using field tags) and their values
// will be set when their corresponding options appear in the command line
// arguments.
func NewGroup(name string, data interface{}) *Group {
	ret := &Group{
		Name:       name,
		LongNames:  make(map[string]*Option),
		ShortNames: make(map[rune]*Option),
		data:       data,
	}

	ret.Error = ret.scan()
	return ret
}
