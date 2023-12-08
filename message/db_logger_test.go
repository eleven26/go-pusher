package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	dbMessageLogger := &DbMessageLogger{}
	_ = dbMessageLogger.Log(Log{
		Uid:     "123",
		Message: "Hello World",
	})

	var log DbLog
	db.First(&log)
	assert.Equal(t, log.Uid, "123")
	assert.Equal(t, log.Message, "Hello World")
}
