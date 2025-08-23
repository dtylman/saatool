package ui

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// IsDirty checks if there are unsaved changes in an object by calculating its hash and comparing it to a stored hash.
func IsDirty(obj interface{}, prevHash []byte) (bool, []byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return false, nil, fmt.Errorf("failed to marshal object: %v", err)
	}
	hash := md5.Sum(data)

	if bytes.Equal(hash[:], prevHash) {
		return false, hash[:], nil
	}
	return true, hash[:], nil
}
