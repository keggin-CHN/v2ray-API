package runtime

import "encoding/json"

func (s State) JSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
