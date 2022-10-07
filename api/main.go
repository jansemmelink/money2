package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jansemmelink/money/bank"
	"github.com/jansemmelink/money/dot"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func main() {
	addr := flag.String("addr", "localhost:12345", "HTTP Server address")
	flag.Parse()

	mux := mux.NewRouter()
	mux.HandleFunc("/accounts", hdlr(getAccounts)).Methods(http.MethodGet)
	http.Handle("/", mux)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		panic(fmt.Sprintf("HTTP server failed on addr(%s): %+v", *addr, err))
	}
}

func hdlr(f interface{}) http.HandlerFunc {
	funcType := reflect.TypeOf(f)
	if funcType.Kind() != reflect.Func {
		panic("handler is not a function")
	}

	funcValue := reflect.ValueOf(f)

	hasReq := funcType.NumIn() == 2
	var reqType reflect.Type
	if hasReq {
		reqType = funcType.In(1)
	}

	type CtxUuid struct{}

	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		ctx := context.Background()

		uuid := uuid.New().String()
		ctx = context.WithValue(ctx, CtxUuid{}, uuid)

		var err error
		var res interface{}
		defer func() {
			if err != nil {
				log.Errorf("failed: %+v", err)
				http.Error(httpRes, err.Error(), http.StatusInternalServerError)
				return
			}

		}()

		args := []reflect.Value{
			reflect.ValueOf(ctx),
		}

		if hasReq {
			reqPtrValue := reflect.New(reqType)
			//parse body into request
			if err = json.NewDecoder(httpReq.Body).Decode(reqPtrValue.Interface()); err != nil && err != io.EOF {
				err = errors.Wrapf(err, "failed to decode JSON request body into %v", reqType)
				return
			}
			//parse URL params into request (over writing body if duplicate)
			for n, v := range httpReq.URL.Query() {
				if err = dot.Set(reqPtrValue.Interface(), n, v); err != nil {
					err = errors.Wrapf(err, "cannot set URL query param %s=%v", n, v)
					return
				}
			}
			//apply URL path params
			for n, v := range mux.Vars(httpReq) {
				if err = dot.Set(reqPtrValue.Interface(), n, v); err != nil {
					err = errors.Wrapf(err, "cannot set URL path param %s=%v", n, v)
					return
				}
			}

			//validate the request
			if validator, ok := reqPtrValue.Interface().(Validator); ok {
				if err = validator.Validate(); err != nil {
					err = errors.Wrapf(err, "invalid request")
					return
				}
				log.Debugf("Validated (%T)%+v", reqPtrValue.Interface(), reqPtrValue.Elem().Interface())
			} else {
				log.Debugf("NOT Validated (%T)%+v", reqPtrValue.Interface(), reqPtrValue.Elem().Interface())
			}

			args = append(args, reflect.ValueOf(reqPtrValue.Elem().Interface()))
		}

		results := funcValue.Call(args)
		if results[1].Interface() != nil {
			err = results[1].Interface().(error)
			if err != nil {
				err = errors.Wrapf(err, "handler failed")
				return
			}
		}

		res = results[0].Interface()
		if res != nil {
			var jsonRes []byte
			jsonRes, err = json.Marshal(res)
			if err != nil {
				err = errors.Wrapf(err, "cannot encode response")
				return
			}

			httpRes.Header().Set("Content-Type", "application/json")
			httpRes.Write(jsonRes)
		}
	}
}

type AccountFilter struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Limit int    `json:"limit"`
}

func getAccounts(ctx context.Context, req AccountFilter) ([]bank.Account, error) {
	accList, err := bank.GetAccounts(req.Name, req.Type, req.Limit)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get accounts")
	}
	return accList, nil
}

type Validator interface {
	Validate() error
}
