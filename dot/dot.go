package dot

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

//set dot notation names in values
func Set(tgtPtr interface{}, n string, v interface{}) error {
	log.Debugf("Set(%T, %s=(%T)%v", tgtPtr, n, v, v)
	xt := reflect.TypeOf(tgtPtr)
	if xt.Kind() != reflect.Ptr {
		return errors.Errorf("cannot set %T != pointer", tgtPtr)
	}

	names := strings.SplitN(n, ".", 2)
	log.Debugf("%d names: %+v", len(names), names)
	if len(names) == 0 {
		return errors.Errorf("NYI")
	}
	if names[0] == "" {
		return errors.Errorf("cannot set empty name")
	}
	return set(reflect.ValueOf(tgtPtr).Elem(), names, v)
}

func set(x reflect.Value, names []string, v interface{}) error {
	log.Debugf("set(%v, %s=(%T)%v", x.Type(), names, v, v)
	switch x.Kind() {
	case reflect.Int:
		if reflect.TypeOf(v).Kind() != reflect.Int {
			s := fmt.Sprintf("%v", v)
			i64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				//try to remove outer "[...]" if e.g. passed in []string{onevalue} likely from url params
				if i64, err = strconv.ParseInt(strings.Trim(s, "[]\"'"), 10, 64); err != nil {
					return errors.Errorf("%T cannot parse as int", v)
				}
			}
			x.Set(reflect.ValueOf(int(i64)))
		} else {
			x.Set(reflect.ValueOf(v))
		}
		return nil
	case reflect.String:
		if reflect.TypeOf(v).Kind() != reflect.String {
			if as, ok := v.([]string); ok {
				switch len(as) {
				case 0:
					x.Set(reflect.ValueOf(""))
				case 1:
					x.Set(reflect.ValueOf(as[0]))
				default:
					return errors.Errorf("cannot set %d strings into a string", len(as))
				}
			} else {
				x.Set(reflect.ValueOf(fmt.Sprintf("%v", v)))
			}
		} else {
			x.Set(reflect.ValueOf(v))
		}
		return nil
	case reflect.Map:
		return errors.Errorf("NYI Map")
	case reflect.Struct:
		return setStruct(x, names, v)
	case reflect.Slice:
		return errors.Errorf("NYI Slice")
	default:
		log.Debugf("x=%T kind=%v", x, x.Kind())
		return errors.Errorf("cannot set kind %v", x.Kind())
	}
}

func setStruct(structValue reflect.Value, names []string, v interface{}) error {
	log.Debugf("setStruct(%v, %+v=(%T)%v", structValue.Type(), names, v, v)
	t := structValue.Type()
	log.Debugf("struct type %s", t.Name())
	for n := 0; n < t.NumField(); n++ {
		f := t.Field(n)
		jsonTags := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		log.Debugf("field %+v", jsonTags)
		if len(jsonTags) > 0 && jsonTags == names[0] {
			return set(structValue.Field(n), names[1:], v)
		}
	}
	return errors.Errorf("field(%s) not found", names[0])
}
