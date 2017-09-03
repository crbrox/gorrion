package gorrion

import "fmt"

type gorrionErr string

// invalid target operation
const (
	ErrNotFoundAttr   gorrionErr = "not found attr"
	ErrExistentAttr   gorrionErr = "existent attribute"
	ErrExistentEntity gorrionErr = "existent entity"
	ErrNotFoundEntity gorrionErr = "not found entity"
)

// invalid object as an entity
const (
	ErrMissingEntityId    gorrionErr = "missing entity id"
	ErrMissingValueField  gorrionErr = "missing value in attr"
	ErrAttrNotAnObject    gorrionErr = "attr is not an object"
	ErrIdNotAString       gorrionErr = "id is not a string"
	ErrTypeNotAString     gorrionErr = "type is not a string"
	ErrAttrTypeNotAString gorrionErr = "attr type is not a string"
	ErrMDNotAnObject      gorrionErr = "metadata is not an object"
	ErrEmptyObject        gorrionErr = "empty object"

	ErrEmptyEntityID   gorrionErr = "empty entity id"
	ErrEmptyEntityType gorrionErr = "empty entity type"
)

// invalid attr set
const (
	// "id" is not a valid attribute
	ErrInvalidAttrID gorrionErr = idField + " is not valid as atrr"
	// "type" is not a valid attribute
	ErrInvalidAttrType gorrionErr = typeField + " is not valid as atrr"
)

const (
	ErrInvalidJSON        gorrionErr = "invalid JSON" //TODO: este puede ser innecesario
	ErrContentTypeNotJSON gorrionErr = "content-type is not application/json"
	ErrParsingJSON        gorrionErr = "error parsing JSON"
)

func (e gorrionErr) Error() string {
	return string(e)
}

func ErrToJSON(err error) string {
	return fmt.Sprintf(`{"error":%q}`, err.Error())
}

func (e gorrionErr) Status() int {
	var code = 500
	switch e {
	case ErrNotFoundAttr,
		ErrNotFoundEntity:
		code = 404
	case
		ErrExistentAttr,
		ErrExistentEntity,
		ErrMissingEntityId,
		ErrMissingValueField,
		ErrAttrNotAnObject,
		ErrIdNotAString,
		ErrTypeNotAString,
		ErrAttrTypeNotAString,
		ErrMDNotAnObject,
		ErrInvalidJSON,
		ErrEmptyObject,
		ErrInvalidAttrID,
		ErrInvalidAttrType,
		ErrEmptyEntityID,
		ErrEmptyEntityType,
		ErrContentTypeNotJSON,
		ErrParsingJSON:
		code = 400
	default:
		code = 500
	}
	return code
}
