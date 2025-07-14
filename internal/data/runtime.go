package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {

	tmp := fmt.Sprintf("%d mins", r)
	quotedJsonValue := strconv.Quote(tmp)
	return []byte(quotedJsonValue), nil

}
