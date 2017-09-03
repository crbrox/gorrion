package gorrion

import (
	"encoding/json"
	"fmt"
	"testing"
)

func testGetErrorCases() map[gorrionErr]int {
	return map[gorrionErr]int{
		ErrNotFoundAttr:                 404,
		ErrExistentAttr:                 400,
		ErrExistentEntity:               400,
		ErrNotFoundEntity:               404,
		ErrMissingEntityId:              400,
		ErrMissingValueField:            400,
		ErrAttrNotAnObject:              400,
		ErrIdNotAString:                 400,
		ErrTypeNotAString:               400,
		ErrAttrTypeNotAString:           400,
		ErrMDNotAnObject:                400,
		ErrEmptyObject:                  400,
		ErrInvalidJSON:                  400,
		ErrContentTypeNotJSON:           400,
		ErrParsingJSON:                  400,
		gorrionErr("[NOT ERRROR CODE]"): 500,
	}
}

func TestErrorStatus(t *testing.T) {
	var cases = testGetErrorCases()

	for err, code := range cases {
		if got := err.Status(); got != code {
			t.Error(gotWanted(got, code) + fmt.Sprintf("(%s)", err))
		}
	}

}

func TestErrToJSON(t *testing.T) {
	var cases = testGetErrorCases()

	for err := range cases {
		o := map[string]interface{}{
			"error": err.Error(),
		}
		res, errJ := json.Marshal(o)
		if errJ != nil {
			t.Fatal(unexpected(errJ))
		}

		if got, wanted := ErrToJSON(err), string(res); got != wanted {
			t.Error(gotWanted(got, wanted) + fmt.Sprintf("(%s)", err))
		}
	}
}
