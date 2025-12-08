package entity

import (
	"encoding/json"
	"strconv"
	"strings"
)

func ParseUint32SliceFromJSONArrayString(s string) ([]uint32, error) {
	var raw []json.Number
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	out := make([]uint32, 0, len(raw))
	for _, n := range raw {
		u64, err := strconv.ParseUint(n.String(), 10, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, uint32(u64))
	}
	return out, nil
}
