package message

type Log struct {
	Uid     string `json:"uid"`
	Message string `json:"message"`
}

type Logger interface {
	Log(log Log) error
}
