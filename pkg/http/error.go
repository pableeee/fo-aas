package http

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Body   string `json:"body"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)

	return fmt.Sprintf("error: %s", string(b))
}
