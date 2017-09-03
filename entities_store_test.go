package gorrion

import (
	"testing"
)

func TestCreateEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID", Type: "Type"}
		e   = NewEntity(id)
		e2  = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	e.Attrs["temperature"] = Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	e2, err = GetEntity(id)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	if !equalObjects(e, e2) {
		t.Error(gotWanted(e2, e))
	}
}

func TestCreateEntity_Dup(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	// An already existent entity should be I1/T1/S/SP
	e := NewEntity(EntityID{ID: "I1", Type: "T1", Service: "S", ServicePath: "SP"})
	e.Attrs["x"] = Attribute{Value: 12.34, Type: "float"}
	err := CreateEntity(e)
	if err != ErrExistentEntity {
		t.Error(gotWanted(err, ErrExistentEntity))
	}

}

func TestCreateEntity_EmptyID(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	e := NewEntity(EntityID{ID: "", Type: "T1", Service: "S", ServicePath: "SP"})
	e.Attrs["x"] = Attribute{Value: 12.34, Type: "float"}
	err := CreateEntity(e)
	if err != ErrEmptyEntityID {
		t.Error(gotWanted(err, ErrEmptyEntityID))
	}
}

func TestCreateEntity_IDAsAttr(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	e := NewEntity(EntityID{ID: "id", Type: "t", Service: "S", ServicePath: "SP"})
	e.Attrs[idField] = Attribute{Value: 12.34, Type: "float"}
	err := CreateEntity(e)
	if err != ErrInvalidAttrID {
		t.Error(gotWanted(err, ErrInvalidAttrID))
	}
}

func TestCreateEntity_TypeAsAttr(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	e := NewEntity(EntityID{ID: "id", Type: "t", Service: "S", ServicePath: "SP"})
	e.Attrs[typeField] = Attribute{Value: 12.34, Type: "float"}
	err := CreateEntity(e)
	if err != ErrInvalidAttrType {
		t.Error(gotWanted(err, ErrInvalidAttrType))
	}
}

func TestCreateEntity_EmptyType(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	e := NewEntity(EntityID{ID: "id", Type: "", Service: "S", ServicePath: "SP"})
	e.Attrs["x"] = Attribute{Value: 12.34, Type: "float"}
	err := CreateEntity(e)
	if err != ErrEmptyEntityType {
		t.Error(gotWanted(err, ErrEmptyEntityType))
	}
}

func TestFetchEntityNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID", Type: "Type"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = GetEntity(id)
	if err != ErrNotFoundEntity {
		t.Error(gotWanted(err, ErrNotFoundEntity))
	}
}

func TestDeleteEntity(t *testing.T) {
	var (
		e   = NewEntity(EntityID{ID: "ID_to_delete", Type: "Type"})
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	e.Attrs["temperature"] = Attribute{Value: 32.0, Type: "celsius", Md: nil}
	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	err = DeleteEntity(e.ID)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	e, err = GetEntity(e.ID)
	if err != ErrNotFoundEntity {
		t.Error(gotWanted(err, ErrNotFoundEntity))
	}
}

func TestDeleteEntityNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID", Type: "Type"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = DeleteEntity(id)
	if err != ErrNotFoundEntity {
		t.Error(gotWanted(ErrNotFoundEntity, err))
	}
}

func TestDeleteAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_DeleteAttr", Type: "Type"}
		e   = NewEntity(id)
		e2  = NewEntity(id)
		old *Entity
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	e.Attrs["temperature"] = Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["pressure"] = Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	old, err = DeleteAttr(id, "temperature")
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !equalObjects(e, old) {
		t.Error(gotWanted(old, e))
	}

	// And check if we removed it really
	e2, err = GetEntity(id)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	// Now, remove the attribute from "by hand"
	delete(e.Attrs, "temperature")
	// and the object in DB must be e without the attribute
	if !equalObjects(e2, e) {
		t.Error(gotWanted(e2, e))
	}

	// delete a nonexistent attribute
	old, err = DeleteAttr(id, "speed")
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !equalObjects(e, old) {
		t.Error(gotWanted(old, e))
	}
	// And check the entity is the same in DB
	e3, err := GetEntity(id)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !equalObjects(e3, e) {
		t.Error(gotWanted(e3, e))
	}
}

func TestDeleteAttrNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = DeleteAttr(id, "temperature")
	if err != ErrNotFoundEntity {
		t.Error(gotWanted(err, ErrNotFoundEntity))
	}
}

func TestGetAttr(t *testing.T) {
	var (
		id   = EntityID{ID: "ID_GetAttrr", Type: "Type"}
		e    = NewEntity(id)
		attr Attribute
		err  error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature
	pressure := Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	e.Attrs["pressure"] = pressure
	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attr, err = GetAttr(id, "temperature")
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !equalObjects(attr, temperature) {
		t.Error(gotWanted(attr, temperature))
	}

	attr, err = GetAttr(id, "pressure")
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !equalObjects(attr, pressure) {
		t.Error(gotWanted(attr, pressure))
	}
}

func TestGetAttrNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = GetAttr(id, "temperature")
	if err != ErrNotFoundEntity {
		t.Error(gotWanted(err, ErrNotFoundEntity))
	}
}

func TestGetAttrNoAttr(t *testing.T) {

	var (
		id  = EntityID{ID: "ID_GetAttrr", Type: "Type"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 23.45, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature
	pressure := Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	e.Attrs["pressure"] = pressure
	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	_, err = GetAttr(id, "inexistent_attribute")
	if err != ErrNotFoundAttr {
		t.Error(gotWanted(err, ErrNotFoundAttr))
	}

}

func TestSetAttr(t *testing.T) {
	var (
		id   = EntityID{ID: "ID_SetAttrr", Type: "Type", Service: "Valencia", ServicePath: "/Alumbrado"}
		e    = NewEntity(id)
		attr Attribute
		err  error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	pressure := Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	old, err := SetAttr(id, "pressure", &pressure)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	// temperature remains the same
	attr, err = GetAttr(id, "temperature")
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !equalObjects(old, e) {
		t.Error(gotWanted(old, e))
	}

	// pressure is stored properly
	attr, err = GetAttr(id, "pressure")
	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !equalObjects(attr, pressure) {
		t.Errorf(gotWanted(attr, pressure))
	}
}

func TestSetAttrIDAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_SAA", Type: "T", Service: "S", ServicePath: "/SP"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	_, err = SetAttr(id, idField, &Attribute{})

	if err != ErrInvalidAttrID {
		t.Error(gotWanted(err, ErrInvalidAttrID))
	}
}

func TestSetAttrTypeAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_SAA", Type: "T", Service: "S", ServicePath: "/SP"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	_, err = SetAttr(id, typeField, &Attribute{})

	if err != ErrInvalidAttrType {
		t.Error(gotWanted(err, ErrInvalidAttrType))
	}
}

func TestSetAttrNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = SetAttr(id, "temperature", &Attribute{})
	if err != ErrNotFoundEntity {
		t.Errorf(gotWanted(err, ErrNotFoundEntity))
	}
}

func TestAddAttrs(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_AddAttrs", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	pressure := Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	attrs := map[string]Attribute{"pressure": pressure}
	_, err = AddAttrs(id, attrs)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	newAttrs, err := GetAllAttrs(id)
	if err != nil {
		t.Fatal(unexpected(err))
	}
	// the whole set
	both := map[string]Attribute{
		"temperature": temperature,
		"pressure":    pressure,
	}
	if !equalObjects(newAttrs, both) {
		t.Errorf(gotWanted(newAttrs, both))
	}
}

func TestAddAttrsExistent(t *testing.T) {
	var (
		id  = EntityID{ID: "xxx", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	temperature.Value = 1945
	attrs := map[string]Attribute{"temperature": temperature}

	_, err = AddAttrs(id, attrs)
	if err != ErrExistentAttr {
		t.Error(gotWanted(err, ErrExistentAttr))
	}

}

func TestAddAttrsNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = AddAttrs(id, map[string]Attribute{"x": {}})
	if err != ErrNotFoundEntity {
		t.Error(gotWanted(err, ErrNotFoundEntity))
	}
}

func TestAddAttrsIDAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "xxx", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"id": {}}

	_, err = AddAttrs(id, attrs)
	if err != ErrInvalidAttrID {
		t.Error(gotWanted(err, ErrInvalidAttrID))
	}
}

func TestAddAttrsTypeAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "xxx", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"type": {}}

	_, err = AddAttrs(id, attrs)
	if err != ErrInvalidAttrType {
		t.Error(gotWanted(err, ErrInvalidAttrType))
	}
}

func TestUpdateAttrs(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_UpdateAttrs", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	temperature.Value = 1921
	attrs := map[string]Attribute{"temperature": temperature}
	_, err = UpdateAttrs(id, attrs)
	if err != nil {
		t.Fatalf(unexpected(err), err)
	}

	newAttrs, err := GetAllAttrs(id)
	if err != nil {
		t.Fatalf(unexpected(err), err)
	}

	if !equalObjects(newAttrs, attrs) {
		t.Error(gotWanted(newAttrs, attrs))
	}
}

func TestUpdateAttrsNotExist(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_UpdateAttrsNotExist", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature

	err = CreateEntity(e)
	if err != nil {
		t.Fatalf(unexpected(err), err)
	}

	pressure := Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	attrs := map[string]Attribute{"pressure": pressure}
	_, err = UpdateAttrs(id, attrs)
	if err != ErrNotFoundAttr {
		t.Errorf("wanted %#v, got %#v", ErrNotFoundAttr, err)
	}
}

func TestUpdateAttrsNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = UpdateAttrs(id, map[string]Attribute{"x": {}})
	if err != ErrNotFoundEntity {
		t.Errorf("wanted %#v, got %#v", ErrNotFoundEntity, err)
	}
}

func TestUpdateAttrsIDAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "xxx", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	// This test works because the attr name is checked before trying to perform the
	// database action, so it's not necessary that the attr exists before the update

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"id": {}}

	_, err = UpdateAttrs(id, attrs)
	if err != ErrInvalidAttrID {
		t.Error(gotWanted(err, ErrInvalidAttrID))
	}
}

func TestUpdateAttrsTypeAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "xxx", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	// This test works because the attr name is checked before trying to perform the
	// database action, so it's not necessary that the attr exists before the update

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"type": {}}

	_, err = UpdateAttrs(id, attrs)
	if err != ErrInvalidAttrType {
		t.Error(gotWanted(err, ErrInvalidAttrType))
	}
}

func TestAddOrUpdateAttrs(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_AddOrUpdateAttrs", Type: "Type", Service: "Málaga", ServicePath: "/Diputacion"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature
	open := Attribute{Value: true, Type: "alarm", Md: map[string]interface{}{}}
	e.Attrs["open"] = open
	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	pressure := Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	temperature.Value = -10
	attrs := map[string]Attribute{"pressure": pressure, "temperature": temperature}
	_, err = AddOrUpdateAttrs(id, attrs)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	newAttrs, err := GetAllAttrs(id)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	// Add temperature to attrs to get the whole set
	three := map[string]Attribute{
		"temperature": temperature,
		"pressure":    pressure,
		"open":        open,
	}

	if !equalObjects(newAttrs, three) {
		t.Errorf("wanted %#v, got %#v", three, newAttrs)
	}
}

func TestAddOrUpdateAttrsNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = AddOrUpdateAttrs(id, map[string]Attribute{"x": {}})
	if err != ErrNotFoundEntity {
		t.Errorf("wanted %#v, got %#v", ErrNotFoundEntity, err)
	}
}

func TestAddOrUpdateAttrsIDAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "xxx", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"id": {}}

	_, err = AddOrUpdateAttrs(id, attrs)
	if err != ErrInvalidAttrID {
		t.Error(gotWanted(err, ErrInvalidAttrID))
	}
}

func TestAddOrUpdateAttrsTypeAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "xxx", Type: "Type", Service: "Logroño", ServicePath: "/MedioAmbiente"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"type": {}}

	_, err = AddAttrs(id, attrs)
	if err != ErrInvalidAttrType {
		t.Error(gotWanted(err, ErrInvalidAttrType))
	}
}

func TestSetAllAttrs(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_SAA", Type: "T", Service: "S", ServicePath: "/SP"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	temperature := Attribute{Value: 32.0, Type: "celsius", Md: map[string]interface{}{}}
	e.Attrs["temperature"] = temperature
	pressure := Attribute{Value: 10022, Type: "millibar", Md: map[string]interface{}{}}
	e.Attrs["pressure"] = pressure
	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	lon := Attribute{Value: -2.4360722, Type: "Number", Md: map[string]interface{}{}}
	lat := Attribute{Value: 39.4314637, Type: "Number", Md: map[string]interface{}{}}
	attrs := map[string]Attribute{"lon": lon, "lat": lat}
	_, err = SetAllAttrs(id, attrs)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	newAttrs, err := GetAllAttrs(id)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	if !equalObjects(newAttrs, attrs) {
		t.Errorf(gotWanted(newAttrs, attrs))
	}
}

func TestSetAllAttrsNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = SetAllAttrs(id, map[string]Attribute{"x": {}})
	if err != ErrNotFoundEntity {
		t.Errorf(gotWanted(err, ErrNotFoundEntity))
	}
}

func TestSetAllAttrsIDAsAttr(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_SAA", Type: "T", Service: "S", ServicePath: "/SP"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"id": {}, "another": {}}
	_, err = SetAllAttrs(id, attrs)

	if err != ErrInvalidAttrID {
		t.Error(gotWanted(err, ErrInvalidAttrID))
	}
}

func TestSetAllAttrsIDAsType(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_SAA", Type: "T", Service: "S", ServicePath: "/SP"}
		e   = NewEntity(id)
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	err = CreateEntity(e)
	if err != nil {
		t.Fatal(unexpected(err))
	}

	attrs := map[string]Attribute{"type": {}, "another": {}}
	_, err = SetAllAttrs(id, attrs)

	if err != ErrInvalidAttrType {
		t.Error(gotWanted(err, ErrInvalidAttrType))
	}
}

func TestGetAllAttrsNoEntity(t *testing.T) {
	var (
		id  = EntityID{ID: "ID_Not_Exist", Type: "T"}
		err error
	)

	setupTestDB(t)
	defer teardownTestDB(t)

	_, err = GetAllAttrs(id)
	if err != ErrNotFoundEntity {
		t.Errorf(gotWanted(err, ErrNotFoundEntity))
	}
}
