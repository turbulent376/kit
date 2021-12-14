package test

import (
	"git.jetbrains.space/orbi/fcsd/kit/er"
	"github.com/stretchr/testify/assert"
	"testing"
)

func AssertAppErr(t *testing.T, err error, code string) {
	assert.Error(t, err)
	appEr, ok := er.Is(err)
	assert.True(t, ok)
	assert.Equal(t, code, appEr.Code())
}
