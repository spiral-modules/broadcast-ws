package ws

import "encoding/json"

// Command contains information send by user.
type Command struct {
	// Cmd type.
	Cmd string `json:"cmd"`

	// Args contains command specific payload.
	Args json.RawMessage `json:"args"`
}

// Unmarshal command data.
func (cmd *Command) Unmarshal(v interface{}) error {
	return json.Unmarshal(cmd.Args, v)
}
