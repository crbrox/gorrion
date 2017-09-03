package gorrion

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEntityToKeyValues(t *testing.T) {

	var cases = []struct {
		attrs  map[string]Attribute
		wanted map[string]interface{}
	}{
		{map[string]Attribute{},
			map[string]interface{}{"id": "ID", "type": "T"}},
		{map[string]Attribute{"a": {Value: "V"}},
			map[string]interface{}{"id": "ID", "type": "T", "a": "V"}},
		{map[string]Attribute{"a": {Value: "V", Type: "t"}},
			map[string]interface{}{"id": "ID", "type": "T", "a": "V"}},
		{map[string]Attribute{"a": {Value: 12}},
			map[string]interface{}{"id": "ID", "type": "T", "a": 12}},
	}
	for i, c := range cases {
		e := Entity{ID: EntityID{ID: "ID", Type: "T"}, Attrs: c.attrs}
		got := e.ToKeyValues()
		if !equalObjects(got, c.wanted) {
			t.Error(gotWanted(got, c.wanted) + fmt.Sprintf(" (%d)", i))
		}
	}
}

func TestEntityToValues(t *testing.T) {
	var e = Entity{
		ID: EntityID{ID: "id", Type: "type", Service: "s", ServicePath: "sp"},
		Attrs: map[string]Attribute{
			"a": Attribute{Value: "A", Type: "AT"},
			"b": Attribute{Value: "B", Type: "BT"},
			"c": Attribute{Value: 12.34, Type: "CT"},
		},
	}

	values, err := e.ToValues([]string{"a"})
	if err != nil {
		t.Fatal(unexpected(err))
	}
	wanted := []interface{}{"A"}
	if !reflect.DeepEqual(values, wanted) {
		t.Error(gotWanted(values, wanted))
	}

	values, err = e.ToValues([]string{"b", "c"})
	if err != nil {
		t.Fatal(unexpected(err))
	}
	wanted = []interface{}{"B", 12.34}
	if !reflect.DeepEqual(values, wanted) {
		t.Error(gotWanted(values, wanted))
	}

	values, err = e.ToValues([]string{"x", "c"})
	if err != ErrNotFoundAttr {
		t.Error(gotWanted(err, ErrNotFoundAttr))
	}
}

func TestFromKeyValues(t *testing.T) {
	var cases = []struct {
		key   string
		value interface{}
		class string
	}{
		{"temperature", 12.34, "Number"},
		{"status", "ON", "Text"},
		{"closed", true, "Boolean"},
		{"location", map[string]float64{"lon": 34.45, "lat": 78.12}, "StructuredValue"},
		{"scores", []int{1, 2, 3}, "StructuredValue"},
		{"empty", nil, "None"},
	}

	ei := EntityID{ID: "i", Type: "t"}

	for i, c := range cases {
		o := map[string]interface{}{"id": ei.ID, "type": ei.Type}
		o[c.key] = c.value
		entity, err := FromKeyValues(o)
		if err != nil {
			t.Error((unexpected(err)))
		}
		if got, wanted := entity.ID.ID, ei.ID; got != wanted {
			t.Error(gotWanted(got, wanted))
		}
		if got, wanted := entity.ID.Type, ei.Type; got != wanted {
			t.Error(gotWanted(got, wanted))
		}
		if got, wanted := entity.Attrs[c.key].Value, c.value; !reflect.DeepEqual(got, wanted) {
			t.Error(gotWanted(got, wanted) + fmt.Sprintf(" (%d)", i))
		}
		if got, wanted := entity.Attrs[c.key].Type, c.class; got != wanted {
			t.Error(gotWanted(got, wanted) + fmt.Sprintf(" (%d)", i))
		}
	}
}

func TestFromKeyValues_InvalidID(t *testing.T) {
	obj := object{idField: 12}
	_, err := FromKeyValues(obj)

	if err != ErrIdNotAString {
		t.Fatal(gotWanted(err, ErrIdNotAString))
	}
}

func TestFromKeyValues_InvalidType(t *testing.T) {
	obj := object{idField: "id", typeField: true}
	_, err := FromKeyValues(obj)

	if err != ErrTypeNotAString {
		t.Error(gotWanted(err, ErrTypeNotAString))
	}
}

func TestFromKeyValues_MissingID(t *testing.T) {
	obj := object{typeField: true}
	_, err := FromKeyValues(obj)

	if err != ErrMissingEntityId {
		t.Error(gotWanted(err, ErrMissingEntityId))
	}
}

func TestFromKeyValues_DefaultType(t *testing.T) {
	obj := object{idField: "id"}
	ent, err := FromKeyValues(obj)

	if err != nil {
		t.Fatal(unexpected(err))
	}

	if ent.ID.Type != defaultEntityType {
		t.Error(gotWanted(ent.ID.Type, defaultEntityType))
	}
}

func TestGorrionError(t *testing.T) {
	if got, wanted := ErrNotFoundAttr.Error(), "not found attr"; got != wanted {
		t.Error(gotWanted(got, wanted))
	}
	if got, wanted := ErrExistentAttr.Error(), "existent attribute"; got != wanted {
		t.Error(gotWanted(got, wanted))
	}
}

func TestEntity_ToObject(t *testing.T) {
	e := NewEntity(EntityID{ID: "id", Type: "type", Service: "S", ServicePath: "SP"})
	attrs := map[string]Attribute{
		"temperature": {Value: 12.5, Type: "celsisu", Md: map[string]interface{}{"MD1": "1", "MD2": 2}},
		"status":      {Value: "ON", Type: "text", Md: map[string]interface{}{"MD3": "3", "MD4": 4}},
	}
	e.Attrs = attrs

	obj := e.ToObject()

	if obj[idField] != e.ID.ID {
		t.Error(gotWanted(obj[idField], e.ID.ID))
	}
	if obj[typeField] != e.ID.Type {
		t.Error(gotWanted(obj[typeField], e.ID.Type))
	}

	for f, v := range obj {
		if f == idField || f == typeField {
			continue
		}
		if !equalObjects(v, attrs[f]) {
			t.Error(gotWanted(v, attrs[f]))
		}
	}
}

func TestFromObject(t *testing.T) {
	obj := object{
		idField:   "id",
		typeField: "type",
		"temperature": map[string]interface{}{
			attrValueField: 12.3,
			attrTypeField:  "celsius",
			attrMDField:    map[string]interface{}{"timestamp": 12435435}},
		"status": map[string]interface{}{
			attrValueField: "ON",
			attrTypeField:  "text",
			attrMDField:    map[string]interface{}{"location": "San Clemente"}},
	}
	wanted := Entity{ID: EntityID{ID: "id", Type: "type"},
		Attrs: map[string]Attribute{
			"temperature": {
				Value: 12.3,
				Type:  "celsius",
				Md:    map[string]interface{}{"timestamp": 12435435}},
			"status": {
				Value: "ON",
				Type:  "text",
				Md:    map[string]interface{}{"location": "San Clemente"}},
		}}

	// id and type are removed from original object
	savedID, savedType := obj[idField], obj[typeField]
	ent, err := FromObject(obj)

	if err != nil {
		t.Fatal(unexpected(err))
	}

	if got, wanted := ent.ID.ID, savedID; got != wanted {
		t.Error(gotWanted(got, wanted))
	}
	if got, wanted := ent.ID.Type, savedType; got != wanted {
		t.Error(gotWanted(got, wanted))
	}

	if !equalObjects(ent, wanted) {
		t.Error(gotWanted(ent, wanted))
	}

	// TODO: think this test again !!
}

func TestFromObject_InvalidID(t *testing.T) {
	obj := object{idField: 12}
	_, err := FromObject(obj)

	if err != ErrIdNotAString {
		t.Fatal(gotWanted(err, ErrIdNotAString))
	}
}

func TestFromObject_InvalidType(t *testing.T) {
	obj := object{idField: "id", typeField: true}
	_, err := FromObject(obj)

	if err != ErrTypeNotAString {
		t.Error(gotWanted(err, ErrTypeNotAString))
	}
}

func TestFromObject_MissingID(t *testing.T) {
	obj := object{typeField: true}
	_, err := FromObject(obj)

	if err != ErrMissingEntityId {
		t.Error(gotWanted(err, ErrMissingEntityId))
	}
}

func TestFromObject_DefaultType(t *testing.T) {
	obj := object{idField: "id"}
	ent, err := FromObject(obj)

	if err != nil {
		t.Fatal(unexpected(err))
	}

	if ent.ID.Type != defaultEntityType {
		t.Error(gotWanted(ent.ID.Type, defaultEntityType))
	}
}

func TestFromObject_AttrNotAnObject(t *testing.T) {
	obj := object{idField: "id", "temperature": 12.34}
	_, err := FromObject(obj)

	if err != ErrAttrNotAnObject {
		t.Error(gotWanted(err, ErrAttrNotAnObject))
	}

}

func TestFromObject_MissingValueField(t *testing.T) {
	obj := object{idField: "id", "temperature": map[string]interface{}{"valor": 12.34}}
	_, err := FromObject(obj)

	if err != ErrMissingValueField {
		t.Error(gotWanted(err, ErrMissingValueField))
	}
}

func TestFromObject_AttrTypeNotString(t *testing.T) {
	obj := object{idField: "id", "temperature": map[string]interface{}{attrValueField: 12.34, attrTypeField: nil}}
	_, err := FromObject(obj)

	if err != ErrAttrTypeNotAString {
		t.Error(gotWanted(err, ErrAttrTypeNotAString))
	}
}

func TestFromObject_MdNotObject(t *testing.T) {
	obj := object{idField: "id", "temperature": map[string]interface{}{
		attrValueField: 12.34, attrTypeField: "t",
		attrMDField: "x",
	}}
	_, err := FromObject(obj)

	if err != ErrMDNotAnObject {
		t.Error(gotWanted(err, ErrMDNotAnObject))
	}
}

func TestAttrsMapFromEntityObject(t *testing.T) {
	o := object{
		idField:   "id",
		typeField: "type",
		"temperature": map[string]interface{}{
			attrValueField: 12.4,
			typeField:      "celsius",
			attrMDField: map[string]interface{}{
				"md1": "md1 value",
			},
		},
	}
	attrs, err := AttrsMapFromEntityObject(o)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if l := len(attrs); l != 1 {
		t.Errorf("got %d attrs, expected length ==1", l)
	}

	// Md field in Attribute is shown as 'metadata' in JSON, so the equality check pass.
	if got, wanted := attrs["temperature"], o["temperature"]; !equalObjects(got, wanted) {
		t.Error(gotWanted(got, wanted))
	}
}

func TestAttrsMapFromObject(t *testing.T) {
	const attrname = "status"
	o := object{
		attrname: map[string]interface{}{
			attrValueField: "ON",
			attrMDField: map[string]interface{}{
				"location": map[string]float64{"lon": -2.4360722, "lat": 39.4314637},
			},
		},
	}
	attrs, err := AttrsMapFromObject(o)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if l := len(attrs); l != 1 {
		t.Errorf("got %d attrs, expected length == 1", l)
	}

	// Md field in Attribute is shown as 'metadata' in JSON, so the equality check pass.
	if got, wanted := attrs[attrname], o[attrname]; !equalObjects(got, wanted) {
		t.Error(gotWanted(got, wanted))
	}

}

func TestAttrsMapFromObject_IDAsAttr(t *testing.T) {
	o := object{
		idField: "id",
		"temperature": map[string]interface{}{
			attrValueField: 12.4,
			typeField:      "celsius",
			attrMDField: map[string]interface{}{
				"md1": "md1 value",
			},
		},
	}
	_, err := AttrsMapFromObject(o)
	if got, wanted := err, ErrInvalidAttrID; !equalObjects(got, wanted) {
		t.Error(gotWanted(got, wanted))
	}
}

func TestAttrsMapFromObject_TypeAsAttr(t *testing.T) {
	o := object{
		typeField: "type",
		"Empty": map[string]interface{}{
			attrValueField: 12.4,
			typeField:      "celsius",
			attrMDField: map[string]interface{}{
				"md1": "md1 value",
			},
		},
	}
	_, err := AttrsMapFromObject(o)
	if got, wanted := err, ErrInvalidAttrType; !equalObjects(got, wanted) {
		t.Error(gotWanted(got, wanted))
	}
}

func TestAttrValueFromObject(t *testing.T) {
	o := map[string]interface{}{
		attrValueField: "value",
		typeField:      "type",
		attrMDField: map[string]interface{}{
			"md": "md val",
		},
	}
	a, err := AttrValueFromObject(o)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	// Md field in Attribute is shown as 'metadata' in JSON, so the equality check pass.
	if got, wanted := a, o; !equalObjects(got, wanted) {
		t.Error(gotWanted(got, wanted))
	}
}

func TestAttrValueFromObject_AttrNotObject(t *testing.T) {

	_, err := AttrValueFromObject("not an object")
	if got, wanted := err, ErrAttrNotAnObject; got != wanted {
		t.Error(gotWanted(got, wanted))
	}

}

func TestAttrValueFromObject_MissingValue(t *testing.T) {
	o := map[string]interface{}{
		typeField: "type",
		attrMDField: map[string]interface{}{
			"md": "md val",
		},
	}
	_, err := AttrValueFromObject(o)
	if got, wanted := err, ErrMissingValueField; got != wanted {
		t.Error(gotWanted(got, wanted))
	}
}

func TestAttrValueFromObject_AttrTypeNotAString(t *testing.T) {
	o := map[string]interface{}{
		attrValueField: "value",
		typeField:      12,
		attrMDField: map[string]interface{}{
			"md": "md val",
		},
	}
	_, err := AttrValueFromObject(o)
	if got, wanted := err, ErrAttrTypeNotAString; got != wanted {
		t.Error(gotWanted(got, wanted))
	}
}

func TestAttrValueFromObject_MDNotAnObject(t *testing.T) {
	o := map[string]interface{}{
		attrValueField: "value",
		typeField:      "type",
		attrMDField:    true,
	}
	_, err := AttrValueFromObject(o)
	if got, wanted := err, ErrMDNotAnObject; got != wanted {
		t.Error(gotWanted(got, wanted))
	}
}
