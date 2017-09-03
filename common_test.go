package gorrion

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func equalObjects(o1, o2 interface{}) bool {
	j1, err := json.Marshal(o1)
	if err != nil {
		return false
	}
	j2, err := json.Marshal(o2)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(j1, j2)
}

func gotWanted(a, b interface{}) string {
	return fmt.Sprintf("wanted: %v, got %v ", b, a)

}

func unexpected(e error) string {
	return fmt.Sprintf("unexpected error %v ", e)
}
