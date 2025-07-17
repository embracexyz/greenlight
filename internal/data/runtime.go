package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrorInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r Runtime) MarshalJSON() ([]byte, error) {

	tmp := fmt.Sprintf("%d mins", r)
	quotedJsonValue := strconv.Quote(tmp)
	return []byte(quotedJsonValue), nil

}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// expected: "n mins"
	// 去掉quote
	unquotedValues, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrorInvalidRuntimeFormat
	}

	// split 为2
	splitValues := strings.Split(unquotedValues, " ")
	if len(splitValues) != 2 || splitValues[1] != "mins" {
		return ErrorInvalidRuntimeFormat
	}

	// 转int32，然后类型转换
	runtimeValue, err := strconv.ParseInt(splitValues[0], 0, 32)
	if err != nil {
		return ErrorInvalidRuntimeFormat
	}

	*r = Runtime(runtimeValue)
	return nil
}
