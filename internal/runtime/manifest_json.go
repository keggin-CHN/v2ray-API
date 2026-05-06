package runtime

import "encoding/json"

func (m Manifest) JSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}
