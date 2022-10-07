package bank

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-msvc/errors"
)

type Amount struct {
	mc int64 //millicents
}

//pass in string in main currency and may include cents
//pass in any int variant in main currency, no cents
//pass in float in main currency, may include cents
func NewAmount(v interface{}) (Amount, error) {
	switch v := v.(type) {
	case string:
		s := strings.Trim(v, " \t")
		sign := int64(1)
		if s[0] == '-' {
			sign = -1
			s = s[1:]
		}
		parts := strings.SplitN(s, ".", 2)
		vMain := int64(0)
		if len(parts) > 0 {
			var err error
			vMain, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return Amount{}, fmt.Errorf("invalid main=\"%s\"", parts[0])
			}
		}
		vCents := int64(0)
		if len(parts) > 1 {
			var err error
			for len(parts[1]) < 3 {
				parts[1] = parts[1] + "0"
			}
			vCents, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return Amount{}, fmt.Errorf("invalid cents=\"%s\"", parts[1])
			}
		}
		aa := Amount{mc: sign * (vMain*1000 + vCents)}
		log.Debugf("\"%s\" -> %v %v %v -> {%d}=%v", v, sign, vMain, vCents, aa.mc, aa)
		return aa, nil
	case int:
		return Amount{mc: int64(v) * 100000}, nil
	case int8:
		return Amount{mc: int64(v) * 100000}, nil
	case uint8:
		return Amount{mc: int64(v) * 100000}, nil
	case int16:
		return Amount{mc: int64(v) * 100000}, nil
	case uint16:
		return Amount{mc: int64(v) * 100000}, nil
	case int32:
		return Amount{mc: int64(v) * 100000}, nil
	case uint32:
		return Amount{mc: int64(v) * 100000}, nil
	case int64:
		return Amount{mc: int64(v) * 100000}, nil
	case uint64:
		return Amount{mc: int64(v) * 100000}, nil
	case float32:
		return Amount{mc: int64(v) * 100000}, nil
	case float64:
		return Amount{mc: int64(v) * 100000}, nil
	}
	return Amount{}, fmt.Errorf("(%T)%v is not a valid amount", v, v)
}

func (a Amount) MilliCents() int64 {
	return a.mc
}

func (a Amount) Cents() int64 {
	return a.mc / 1000
}

func (a Amount) Add(b Amount) Amount {
	aa := Amount{mc: a.mc + b.mc}
	log.Debugf("%s + %s = %s", a, b, aa)
	return aa
}

func (a Amount) Sub(b Amount) Amount {
	return Amount{mc: a.mc - b.mc}
}

func (a Amount) String() string {
	m := a.mc / 1000
	c := (a.mc % 1000) / 10
	if c < 0 {
		c = -c
	}
	return fmt.Sprintf("%d.%02d", m, c)
}

func (a *Amount) Scan(value interface{}) error {
	if byteArray, ok := value.([]uint8); ok {
		strValue := string(byteArray)
		//if int value - assume
		i64, err := strconv.ParseInt(strValue, 10, 64)
		if err == nil {
			//simple integer value is full currency unit, i.e. "1" = 100*1000 millicents
			i64 *= 100000 //=100c*1000mc/c
		} else {
			//not simple int - assume currentcy with decimal
			//e.g. 123.45 = 12345000 millicents
			dotIndex := strings.Index(strValue, ".")
			if dotIndex < 0 {
				return errors.Errorf("\"%s\" is not formatted as \"123.45\" or \"123\"")
			}
			nrDec := len(strValue) - dotIndex - 1 //e.g. 6 for "123.45" with dotIndex=3, or 4 for "1.23" and dotIndex=1 or 7 for "1234.56" and dotIndex=4
			s := strValue[0:dotIndex] + strValue[dotIndex+1:]
			i64, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				return errors.Errorf("\"%s\" removed dot -> \"%s\" is not valid int", strValue, s)
			} else {
				for nrDec < 5 {
					i64 *= 10
					nrDec--
				}
			}
		}
		a.mc = i64
		return nil
	} //if []byte
	if value == nil {
		a.mc = 0
		return nil
	}
	return errors.Errorf("%T is not []uint8", value)
}

func (a Amount) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a *Amount) UnmarshalJSON(v []byte) error {
	s := string(v)
	if len(s) < 2 || !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
		return errors.Errorf("invalid amount string %s (expects quoted float e.g. \"123.45\")", s)
	}
	return a.Scan(v[1 : len(v)-1])
}

func (a Amount) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", a.String())
	return []byte(s), nil
}
