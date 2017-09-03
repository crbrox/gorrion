package gorrion

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	url          = "localhost"
	db           = "gorrion"
	entitiesColl = "ent"
)

var (
	initialSession *mgo.Session
)

func StartStore() (err error) {
	initialSession, err = mgo.Dial(url)
	return err
}

func StopStore() (err error) {
	initialSession.Close()
	return nil
}

func getCol(ei EntityID) string {
	return entitiesColl
}

func GetEntity(ei EntityID) (e *Entity, err error) {
	return GetEntityAttrs(ei, nil)
}

func GetEntityAttrs(ei EntityID, attrs []string) (e *Entity, err error) {
	e = &Entity{}
	query := initialSession.DB(db).C(getCol(ei)).FindId(ei)
	if len(attrs) != 0 {
		attrsFilter := bson.M{}
		for _, a := range attrs {
			attrsFilter["attrs."+a] = true
		}
		query = query.Select(attrsFilter)
	}
	err = query.One(&e)
	if err == mgo.ErrNotFound {
		return nil, ErrNotFoundEntity
	}
	return e, err
}

func DeleteEntity(ei EntityID) error {
	err := initialSession.DB(db).C(getCol(ei)).RemoveId(ei)
	if err == mgo.ErrNotFound {
		return ErrNotFoundEntity
	}
	return err
}

func CreateEntity(e *Entity) error {
	err := ValidateEntity(e)
	if err != nil {
		return err
	}
	err = initialSession.DB(db).C(getCol(e.ID)).Insert(e)
	if mgo.IsDup(err) {
		return ErrExistentEntity
	}
	return err
}

func DeleteAttr(ei EntityID, name string) (old *Entity, err error) {
	old = &Entity{}
	col := initialSession.DB(db).C(getCol(ei))
	change := mgo.Change{
		Update:    bson.M{"$unset": bson.M{"attrs." + name: true}},
		ReturnNew: false,
	}
	_, err = col.Find(bson.M{"_id": ei}).Apply(change, old)
	if err == mgo.ErrNotFound {
		return nil, ErrNotFoundEntity
	}

	return old, err
}

func SetAttr(ei EntityID, name string, attr *Attribute) (old *Entity, err error) {
	err = ValidateAttribute(name, attr)
	if err != nil {
		return nil, err
	}
	old = &Entity{}
	col := initialSession.DB(db).C(getCol(ei))
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"attrs." + name: attr}},
		ReturnNew: false,
	}
	_, err = col.Find(bson.M{"_id": ei}).Apply(change, old)
	if err == mgo.ErrNotFound {
		return nil, ErrNotFoundEntity
	}
	return old, err
}

func GetAttr(ei EntityID, name string) (attr Attribute, err error) {

	/*
		var result struct {
			A Attribute `bson:"A"`
		}


		col := initialSession.DB(db).C(getCol(ei))

		err = col.Pipe([]bson.M{
			{"$match": bson.M{"_id": ei}},
			{"$project": bson.M{"A": "$attrs." + name, "_id": false}},
		}).One(&result)
		if err == mgo.ErrNotFound {
			err = ErrNotFoundEntity
		}
	*/

	e, err := GetEntityAttrs(ei, []string{name})
	if err != nil {
		return attr, err
	}
	attr, ok := e.Attrs[name]
	if !ok {
		return attr, ErrNotFoundAttr
	}
	return e.Attrs[name], err

}

func SetAllAttrs(ei EntityID, attrs map[string]Attribute) (old *Entity, err error) {
	err = ValidateAttrsMap(attrs)
	if err != nil {
		return nil, err
	}
	old = &Entity{}
	col := initialSession.DB(db).C(getCol(ei))
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"attrs": attrs}},
		ReturnNew: false,
	}
	_, err = col.Find(bson.M{"_id": ei}).Apply(change, old)
	if err == mgo.ErrNotFound {
		return nil, ErrNotFoundEntity
	}
	return old, err
}

func GetAllAttrs(ei EntityID) (attrs map[string]Attribute, err error) {
	e, err := GetEntity(ei)
	// GetEntity returns ErrNotFoundEntity already, not check is necessary
	if err != nil {
		return nil, err
	}
	return e.Attrs, err
}

func AddAttrs(ei EntityID, attrs map[string]Attribute) (old *Entity, err error) {
	err = ValidateAttrsMap(attrs)
	if err != nil {
		return nil, err
	}
	// attributes must not exist
	condition := bson.M{"_id": ei}
	for name := range attrs {
		condition["attrs."+name] = bson.M{"$exists": false}
	}

	update := bson.M{}
	for name, attr := range attrs {
		update["attrs."+name] = attr
	}

	old = &Entity{}
	col := initialSession.DB(db).C(getCol(ei))
	change := mgo.Change{
		Update:    bson.M{"$set": update},
		ReturnNew: false,
	}
	_, err = col.Find(condition).Apply(change, old)

	if err == mgo.ErrNotFound {
		err = col.FindId(ei).One(nil)
		if err == mgo.ErrNotFound {
			// the entity does not exist
			return nil, ErrNotFoundEntity
		}
		// some attr is in the entity already ...
		return nil, ErrExistentAttr
	}

	return old, err
}

func UpdateAttrs(ei EntityID, attrs map[string]Attribute) (old *Entity, err error) {
	err = ValidateAttrsMap(attrs)
	if err != nil {
		return nil, err
	}
	// attributes must exist
	condition := bson.M{"_id": ei}
	for name := range attrs {
		condition["attrs."+name] = bson.M{"$exists": true}
	}

	update := bson.M{}
	for name, attr := range attrs {
		update["attrs."+name] = attr
	}

	old = &Entity{}
	col := initialSession.DB(db).C(getCol(ei))
	change := mgo.Change{
		Update:    bson.M{"$set": update},
		ReturnNew: false,
	}
	_, err = col.Find(condition).Apply(change, old)

	if err == mgo.ErrNotFound {
		err = col.FindId(ei).One(nil)
		if err == mgo.ErrNotFound {
			// the entity does not exist
			return nil, ErrNotFoundEntity
		}
		// the entity does not have the attribute
		return nil, ErrNotFoundAttr
	}

	return old, err
}

func AddOrUpdateAttrs(ei EntityID, attrs map[string]Attribute) (old *Entity, err error) {
	// might make SetAttr redundant ...
	err = ValidateAttrsMap(attrs)
	if err != nil {
		return nil, err
	}
	// attributes may exist or not
	condition := bson.M{"_id": ei}
	update := bson.M{}
	for name, attr := range attrs {
		update["attrs."+name] = attr
	}

	old = &Entity{}
	col := initialSession.DB(db).C(getCol(ei))
	change := mgo.Change{
		Update:    bson.M{"$set": update},
		ReturnNew: false,
	}
	_, err = col.Find(condition).Apply(change, old)
	if err == mgo.ErrNotFound {
		err = ErrNotFoundEntity
	}
	return old, err
}
