package db

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/go-msvc/errors"
)

type SqlTime time.Time

func (t *SqlTime) Scan(value interface{}) error {
	//scan from UTC
	if byteArray, ok := value.([]uint8); ok {
		strValue := string(byteArray)
		timeValue, err := time.ParseInLocation("2006-01-02 15:04:05", strValue, time.UTC)
		if err != nil {
			return err
		}
		*t = SqlTime(timeValue)
		return nil
	}
	if value == nil {
		return nil
	}
	return errors.Errorf("%T is not []uint8", value)
}

func (t SqlTime) Value() (driver.Value, error) {
	return t.String(), nil
}

func (t SqlTime) String() string {
	return time.Time(t).UTC().Format("2006-01-02 15:04:05")
}

func (t *SqlTime) UnmarshalJSON(v []byte) error {
	s := string(v)
	if len(s) < 2 || !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
		return errors.Errorf("invalid time string %s (expects quoted \"2006-01-02 15:04:05\")", s)
	}
	return t.Scan(v[1 : len(v)-1])
}

func (t SqlTime) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", t.String())
	return []byte(s), nil
}
