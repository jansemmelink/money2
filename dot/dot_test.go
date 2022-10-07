package dot_test

import (
	"testing"

	"github.com/jansemmelink/money/dot"
)

func TestStruct(t *testing.T) {
	type Y struct {
		Age int `json:"age"`
	}
	type X struct {
		Name string `json:"name,omitempty"`
		Y    Y      `json:"y"`
	}
	x := X{}

	assert(t, dot.Set(&x, "y.age", 10))
	if x.Y.Age != 10 {
		t.Fatalf("%+v != 10", x)
	}

	assert(t, dot.Set(&x, "y.age", "11"))
	if x.Y.Age != 11 {
		t.Fatalf("%+v != 11", x)
	}

	assert(t, dot.Set(&x, "y.age", []string{"12"}))
	if x.Y.Age != 12 {
		t.Fatalf("%+v != 12", x)
	}

	assert(t, dot.Set(&x, "name", "T1"))
	if x.Name != "T1" {
		t.Fatalf("%+v != T1", x)
	}
	assert(t, dot.Set(&x, "name", []string{"T2"})) //one value must convert to string
	if x.Name != "T2" {
		t.Fatalf("%+v != T2", x)
	}
	if err := dot.Set(&x, "name", []string{"T3", "T4"}); err == nil {
		//multiple values not good for string, must fail
		t.Fatalf("did not fail to set []string into string")
	}

	assert(t, dot.Set(&x, "name", ""))
	if x.Name != "" {
		t.Fatalf("%+v != ", x)
	}
}

func assert(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("failed: %+v", err)
	}
}
