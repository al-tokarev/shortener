package resetpool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Name string
	Age  int
}

func (t *TestStruct) Reset() {
	t.Name = ""
	t.Age = 0
}

func TestNewWithNilFactory(t *testing.T) {
	p := New[*TestStruct](nil)

	obj := p.Get()

	require.Nil(t, obj)
}

func TestNewWithFactory(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{Name: "gopher"}
	})

	obj := p.Get()

	if obj == nil {
		t.Fatal("expected non-nil object, got nil")
	}

	require.Equal(t, "gopher", obj.Name)
}

func TestPutResets(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{}
	})

	obj := p.Get()
	obj.Name = "Alex"
	obj.Age = 23

	p.Put(obj)

	require.Empty(t, obj.Name)
	require.Empty(t, obj.Age)
}
