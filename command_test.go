package ws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommand_Unmarshal(t *testing.T) {
	cmd := &Command{
		Cmd:  "",
		Args: []byte(`"hello"`),
	}

	var str string

	assert.NoError(t, cmd.Unmarshal(&str))
	assert.Equal(t, "hello", str)
}
