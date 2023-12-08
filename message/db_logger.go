package message

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type DbLog struct {
	gorm.Model
	Uid       string
	Message   string
	Status    int
	CreatedAt time.Time
}

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("log.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

var _ Logger = &DbMessageLogger{}

type DbMessageLogger struct{}

func (m *DbMessageLogger) Log(log Log) error {
	_ = db.AutoMigrate(&DbLog{})
	db.Create(&DbLog{
		Uid:       log.Uid,
		Message:   log.Message,
		Status:    0,
		CreatedAt: time.Now(),
	})
	return nil
}
