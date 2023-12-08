package message

import (
	"encoding/json"
	"fmt"
)

var _ Logger = &StdoutMessageLogger{}

type StdoutMessageLogger struct{}

func (s *StdoutMessageLogger) Log(log Log) error {
	res, _ := json.Marshal(log)
	fmt.Println("send message: " + string(res))
	return nil
}
