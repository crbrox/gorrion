// models.go
package gorrion

const (
	idField           = "id"
	typeField         = "type"
	defaultEntityType = "Thing"
	attrValueField    = "value"
	attrTypeField     = "type"
	attrMDField       = "metadata"
)

type Entity struct {
	ID    EntityID             `bson:"_id"`
	Attrs map[string]Attribute `bson:"attrs"`
}

type EntityID struct {
	ID          string
	Type        string
	Service     string
	ServicePath string
}

type Attribute struct {
	// Fields in alphabetical order, by their JSON names
	Md    map[string]interface{} `json:"metadata,omitempty"`
	Type  string                 `json:"type,omitempty"`
	Value interface{}            `json:"value"`
}

func NewEntity(id EntityID) *Entity {
	return &Entity{ID: id, Attrs: map[string]Attribute{}}
}

func (e Entity) ToKeyValues() object {
	kvs := object{}
	for k, attr := range e.Attrs {
		kvs[k] = attr.Value
	}
	kvs[idField] = e.ID.ID
	kvs[typeField] = e.ID.Type
	return kvs
}

func (e *Entity) ToValues(attrs []string) (result []interface{}, err error) {
	for _, attr := range attrs {
		if entAttr, ok := e.Attrs[attr]; ok {
			result = append(result, entAttr.Value)
		} else {
			err = ErrNotFoundAttr
			break
		}
	}
	return result, err
}

func FromKeyValues(kv object) (*Entity, error) {
	entity := NewEntity(EntityID{})

	if id, ok := kv[idField]; ok {
		if id, ok := id.(string); ok {
			entity.ID.ID = id
		} else {
			return nil, ErrIdNotAString
		}
	} else {
		return nil, ErrMissingEntityId
	}

	if t, ok := kv[typeField]; ok {
		if t, ok := t.(string); ok {
			entity.ID.Type = t
		} else {
			return nil, ErrTypeNotAString
		}
	} else {
		entity.ID.Type = defaultEntityType
	}

	entity.Attrs = AttrsFromKeyValue(kv)

	return entity, nil
}

func AttrsFromKeyValue(kv object) map[string]Attribute {
	m := map[string]Attribute{}
	for k, v := range kv {
		if k == idField || k == typeField {
			continue
		}
		m[k] = *AttributeFromKeyValue(v)
	}
	return m
}

func AttributeFromKeyValue(v interface{}) *Attribute {
	attr := &Attribute{Value: v}
	switch v.(type) {
	case string:
		attr.Type = "Text"
	case int, float32, float64:
		attr.Type = "Number"
	case bool:
		attr.Type = "Boolean"
	case nil:
		attr.Type = "None"
	default: // TODO: finer grain?
		attr.Type = "StructuredValue"
	}
	// TODO: think about this, specification says it must be an empty object when undefined
	// but may be different implementations (only when rendering, not to store, f.e)
	attr.Md = map[string]interface{}{}

	return attr
}

/*
func (e *Entity) GetField(path []string) (interface{}, error) {
	return getMapField(e.Attrs, path)
}

func getMapField(m map[string]interface{}, path []string) (interface{}, error) {
	if m == nil || len(path) == 0 {
		return nil, FieldNotFound
	}
	if v, ok := m[path[0]]; ok {
		if len(path) == 1 {
			return v, nil
		} else if submap, ok := v.(map[string]interface{}); ok {
			return getMapField(submap, path[1:])
		}
	}
	return nil, FieldNotFound
}
*/

func (e *Entity) ToObject() (o object) {
	o = object{}
	o[idField] = e.ID.ID
	o[typeField] = e.ID.Type
	for k, v := range e.Attrs {
		o[k] = v
	}
	return o
}

func FromObject(o object) (e *Entity, err error) {

	eID := EntityID{}

	if id, ok := o[idField]; ok {
		if id, ok := id.(string); ok {
			eID.ID = id
		} else {
			return nil, ErrIdNotAString
		}
	} else {
		return nil, ErrMissingEntityId
	}

	if t, ok := o[typeField]; ok {
		if t, ok := t.(string); ok {
			eID.Type = t
		} else {
			return nil, ErrTypeNotAString
		}
	} else {
		eID.Type = defaultEntityType
	}

	e = NewEntity(eID)

	attrs, err := AttrsMapFromEntityObject(o)

	if err != nil {
		return nil, err
	}
	e.Attrs = attrs

	return e, nil
}

func AttrsMapFromEntityObject(o object) (m map[string]Attribute, err error) {
	return attrsMapFromObjectAux(o, true) // ignore id, type
}

func AttrsMapFromObject(o object) (m map[string]Attribute, err error) {
	return attrsMapFromObjectAux(o, false) // id or type as attr in an error
}

func attrsMapFromObjectAux(o object, ignoreIDnType bool) (m map[string]Attribute, err error) {
	m = map[string]Attribute{}
	for k, v := range o {
		if k == idField {
			if ignoreIDnType {
				continue
			}
			return m, ErrInvalidAttrID
		}
		if k == typeField {
			if ignoreIDnType {
				continue
			}
			return m, ErrInvalidAttrType
		}
		a, err := AttrValueFromObject(v)
		if err != nil {
			return nil, err
		}
		m[k] = *a
	}
	return m, err
}

func AttrValueFromObject(m interface{}) (a *Attribute, err error) {
	a = &Attribute{}

	v, ok := m.(map[string]interface{})
	if !ok {
		return nil, ErrAttrNotAnObject
	}

	if value, ok := v[attrValueField]; ok {
		a.Value = value
	} else {
		return nil, ErrMissingValueField
	}

	if t, ok := v[attrTypeField]; ok {
		if t, ok := t.(string); ok {
			a.Type = t
		} else {
			return nil, ErrAttrTypeNotAString
		}
	}

	if md, ok := v[attrMDField]; ok {
		if md, ok := md.(map[string]interface{}); ok {
			a.Md = md
		} else {
			return nil, ErrMDNotAnObject
		}
	}

	return a, nil
}

func ValidateEntity(e *Entity) error {
	if len(e.ID.ID) == 0 {
		return ErrEmptyEntityID
	}
	if len(e.ID.Type) == 0 {
		return ErrEmptyEntityType
	}
	if err := ValidateAttrsMap(e.Attrs); err != nil {
		return err
	}
	return nil
}

func ValidateAttrsMap(m map[string]Attribute) error {
	for name, attr := range m {
		err := ValidateAttribute(name, &attr)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateAttribute(name string, a *Attribute) error {
	if name == idField {
		return ErrInvalidAttrID
	}
	if name == typeField {
		return ErrInvalidAttrType
	}
	return nil
}
