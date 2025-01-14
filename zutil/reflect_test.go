package zutil_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

type TestSt struct {
	Name string
	I    int `z:"iii"`
	Note int `json:"note,omitempty"`
}

func (*TestSt) RunTest(t *testing.T) {
	t.Log("RunTest")
}

func (TestSt) RunTest2() {}

type TestSt2 struct {
	Name  string
	Test2 bool
}

func (tt *TestSt2) RunTest(t *testing.T) {
	t.Log("RunTest", tt.Name)
}

func TestRunAllMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := zutil.RunAllMethod(&TestSt{}, t)
	t.Log(err)
	tt.Equal(true, err != nil)

	err = zutil.RunAllMethod(&TestSt2{Name: "AllMethod"}, t)
	t.Log(err)
	tt.Equal(true, err == nil)

	err = zutil.RunAssignMethod(&TestSt2{Name: "AssignMethod"}, func(methodName string) bool {
		t.Log("methodName:", methodName)
		return true
	}, t)
	t.Log(err)
	tt.Equal(true, err == nil)
}

func TestGetAllMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i := 0
	err := zutil.GetAllMethod(&TestSt{}, func(numMethod int, m reflect.Method) error {
		t.Log(m.Name)
		i++
		if m.Name != "RunTest" && m.Name != "RunTest2" {
			return errors.New("mismatch")
		}
		return nil
	})
	tt.Equal(2, i)
	tt.Equal(true, err == nil)

	err = zutil.GetAllMethod("test", nil)
	t.Log(err)
	// tt.Equal(true, err != nil)

	err = zutil.GetAllMethod(&TestSt{}, nil)
	t.Log(err)
	// tt.Equal(true, err == nil)
}

func TestReflectStructField(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var test = &TestSt{}
	tf := reflect.TypeOf(test)
	// fieldPtr := uintptr(unsafe.Pointer(test))
	err := zutil.ReflectStructField(tf, func(numField int, fieldTag string,
		field reflect.StructField) error {
		// fieldPtrOffset := fieldPtr + field.Offset
		switch field.Type.Kind() {
		case reflect.String:
			// noinspection GoVetUnsafePointer
			// *((*string)(unsafe.Pointer(fieldPtrOffset))) = "ok"
		}
		return nil
	})
	tt.EqualNil(err)
	t.Log(test)
}

func TestReflectForNumField(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var test = &struct {
		UpdatedAt time.Time
		*TestSt2
		T2p *TestSt2
		T2  TestSt2
		TestSt
		New     bool
		Updated uint8
	}{}
	rv := reflect.ValueOf(test)
	rv = rv.Elem()
	err := zutil.ReflectForNumField(rv, func(fieldName, fieldTag string, kind reflect.Kind, field reflect.Value) error {
		t.Log(fieldTag, kind, field.Kind())
		return nil
	})
	tt.EqualNil(err)
}

func TestSetValue(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Log(666)
	vv := &TestSt2{Name: "1"}

	v := reflect.ValueOf(vv)
	err := zutil.ReflectForNumField(v.Elem(), func(fieldName, fieldTag string,
		kind reflect.Kind, field reflect.Value) error {
		if fieldName == "Test2" {
			tt.Log(fieldName, true)
			return zutil.SetValue(kind, field, true)
		}
		tt.Log(fieldName, "new")
		return zutil.SetValue(kind, field, "new")
	})
	t.EqualNil(err)
	t.Equal("new", vv.Name)
	t.Equal(true, vv.Test2)
}
