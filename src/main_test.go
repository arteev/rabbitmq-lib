package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

func TestMapArgs(t *testing.T) {
	m := MapArgs()
	assert.NotZero(t, m)
	obj, ok := getObjectEx(m)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{}, obj)

	key := "param"
	value := "value"
	ok = MapArgsAdd(m, key, value)
	assert.True(t, ok)
	gotMap := getObject(m).(map[string]interface{})
	assert.Equal(t, map[string]interface{}{
		"param": "value",
	}, gotMap)

	ok = MapArgsAdd(0, "k", "v")
	assert.False(t, ok)

}

func TestObjects(t *testing.T) {

	obj := struct{}{}
	ptr := uintptr(unsafe.Pointer(&obj))
	putObject(ptr, obj)
	assert.Contains(t, objects, ptr)

	got := getObject(0)
	assert.Nil(t, got)

	got = getObject(ptr)
	assert.Equal(t, obj, got)

	_, ok := getObjectEx(0)
	assert.False(t, ok)

	got, ok = getObjectEx(ptr)
	assert.True(t, ok)
	assert.Equal(t, obj, got)

	FreeObject(ptr)
	assert.NotContains(t, objects, ptr)
}
