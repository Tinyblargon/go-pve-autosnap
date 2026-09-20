package filter

import (
	"errors"
	"strconv"
)

const operator = '@'
const operatorEndByte = ':'

const (
	errEmptyFilter         = "filter is empty"
	errMissingOperatorAt0  = "missing operator at 0"
	errNoOperator          = "no operator"
	errNoOperatorEnd       = "no operator end"
	errInvalidOperator     = "invalid operator"
	errInvalidOperatorChar = "invalid operator char"
	errUndefinedOperator   = "undefined operator"
)

func (f Filters) Parse(raw string) (*Filter, error) {
	if len(raw) == 0 {
		return nil, errors.New(errEmptyFilter)
	}
	if raw[0] != operator {
		return nil, &ParseError{message: errMissingOperatorAt0}
	}
	var index int = 1
	var localIndex int
	var err error
	var alignment bool
	var step FilterStep
	filter := Filter{
		step:      make([]FilterStep, 0, 10),
		alignment: make([]bool, 0, 10),
	}
	for index < len(raw) {
		localIndex, alignment, step, err = f.parseOperator(raw[index:])
		index += localIndex
		if err != nil {
			if parseErr, ok := err.(*ParseError); ok {
				parseErr.charNumber = index
				return nil, parseErr
			}
			return nil, err
		}
		filter.alignment = append(filter.alignment, alignment)
		filter.step = append(filter.step, step)
	}
	return &filter, nil
}

// `@` is already removed when we get here
func (f Filters) parseOperator(raw string) (int, bool, FilterStep, error) {
	var index int
	var affinity bool = true
	switch raw[index] {
	case '-':
		affinity = false
		index++
	case '+':
		index++
	}
	operatorBeginning := index
	var operatorEnd int
	if len(raw) == index+1 {
		return index, false, nil, &ParseError{message: errNoOperator}
	}
	for index < len(raw) {
		if operatorChars[raw[index]] {
			index++
			continue
		}
		if raw[index] == operatorEndByte {
			operatorEnd = index
			if operatorBeginning == operatorEnd { // we only have a `:`, no operator name was provided
				return index, false, nil, &ParseError{message: errInvalidOperator}
			}
			index++
			break
		}
		return index, false, nil, &ParseError{message: errInvalidOperatorChar}
	}
	if operatorEnd == 0 {
		return index, false, nil, &ParseError{message: errNoOperatorEnd}
	}

	operatorName := raw[operatorBeginning:operatorEnd]
	constructor, ok := f.a[Operator(operatorName)]
	if !ok {
		return operatorBeginning, false, nil, &ParseError{message: errUndefinedOperator}
	}

	const empty byte = '0'
	var quote byte = empty
	var escape bool
	var parameterEnd int
loop:
	for index < len(raw) {
		if !escape && raw[index] == '\\' {
			escape = true
			index++
			continue
		}
		escape = false
		if quote == empty {
			switch raw[index] {
			case '"', '\'', '`':
				quote = raw[index]
			case operator:
				index++
				parameterEnd = index - 1
				break loop
			}
			index++
			continue
		}
		if raw[index] == quote { // End of quoted string
			quote = empty
		}
		index++
	}
	var args string
	if parameterEnd == 0 {
		if index > operatorEnd+1 {
			args = raw[operatorEnd+1 : index]
		}
	} else {
		if parameterEnd > operatorEnd+1 {
			args = raw[operatorEnd+1 : parameterEnd]
		}
	}
	filter, err := constructor(args)
	if err != nil {
		return index, false, nil, err
	}

	return index, affinity, filter, nil
}

func split(raw string) []string {
	const empty byte = '0'
	var quote byte = empty
	var escape bool
	args := make([]string, 0)
	var index int
	params := []byte{}
	for index = range raw {
		if !escape && raw[index] == '\\' {
			escape = true
			continue
		}
		escape = false
		if quote == empty {
			switch raw[index] {
			case '"', '\'', '`':
				quote = raw[index]
				continue
			case ',', ' ', ';':
				args = append(args, string(params))
				params = []byte{}
				continue
			}
		} else if raw[index] == quote { // End of quoted string
			quote = empty
			continue
		}
		params = append(params, raw[index])
	}
	args = append(args, string(params))
	return args
}

type ParseError struct {
	message    string
	charNumber int
}

func (p *ParseError) Error() string {
	return p.message + " at: " + strconv.Itoa(p.charNumber)
}

var operatorChars = [256]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
	'A': true,
	'B': true,
	'C': true,
	'D': true,
	'E': true,
	'F': true,
	'G': true,
	'H': true,
	'I': true,
	'J': true,
	'K': true,
	'L': true,
	'M': true,
	'N': true,
	'O': true,
	'P': true,
	'Q': true,
	'R': true,
	'S': true,
	'T': true,
	'U': true,
	'V': true,
	'W': true,
	'X': true,
	'Y': true,
	'Z': true,
	'_': true,
	'a': true,
	'b': true,
	'c': true,
	'd': true,
	'e': true,
	'f': true,
	'g': true,
	'h': true,
	'i': true,
	'j': true,
	'k': true,
	'l': true,
	'm': true,
	'n': true,
	'o': true,
	'p': true,
	'q': true,
	'r': true,
	's': true,
	't': true,
	'u': true,
	'v': true,
	'w': true,
	'x': true,
	'z': true,
}
