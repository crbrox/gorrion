package gorrion

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Query struct {
	ID          []string
	IDPattern   string
	Type        []string
	TypePattern string
	Limit       int
	Offset      int
	Attrs       []string
	OrderBy     []string
	Options     []Option
	//mainly for debugging
	condition bson.M
	attrs     bson.M
	sort      []string
}

type EntityIter struct {
	iter *mgo.Iter
}

func (ei *EntityIter) Next(e *Entity) bool {
	return ei.iter.Next(e)
}

func (ei *EntityIter) Err() error {
	return ei.iter.Err()
}

func (q *Query) Get(service, servicepath string) (eIter *EntityIter, err error) {

	// Build
	var conditions = []bson.M{{"_id.service": service}, {"_id.servicepath": servicepath}}

	if len(q.ID) > 0 {
		conditions = append(conditions, bson.M{"_id.id": bson.M{"$in": q.ID}})
	}
	if q.IDPattern != "" {
		conditions = append(conditions, bson.M{"_id.id": bson.M{"$regex": q.IDPattern}})
	}

	if len(q.Type) > 0 {
		conditions = append(conditions, bson.M{"_id.type": bson.M{"$in": q.Type}})
	}

	if q.TypePattern != "" {
		conditions = append(conditions, bson.M{"_id.type": bson.M{"$regex": q.TypePattern}})
	}

	q.condition = bson.M{"$and": conditions}

	// Select attributes asked for
	q.attrs = bson.M{}
	for _, s := range q.Attrs {
		q.attrs["attrs."+s] = 1
	}

	// Change ! to - in sort fields
	for _, s := range q.OrderBy {
		if s[0] == '!' {
			// desc order
			q.sort = append(q.sort, "-attrs."+s[1:]+".value")
		} else {
			q.sort = append(q.sort, "attrs."+s+".value")
		}
	}

	//  Get iterator
	col := initialSession.DB(db).C(getCol(EntityID{Service: service, ServicePath: servicepath}))
	mgoQ := col.Find(q.condition)

	if q.Limit > 0 {
		mgoQ = mgoQ.Limit(q.Limit)
	}

	if q.Offset > 0 {
		mgoQ = mgoQ.Skip(q.Offset)
	}

	if len(q.attrs) > 0 {
		mgoQ = mgoQ.Select(q.attrs)
	}
	if len(q.sort) > 0 {
		mgoQ = mgoQ.Sort(q.sort...)
	}

	eIter = &EntityIter{iter: mgoQ.Iter()}

	return eIter, nil
}

// mainly for debugging
func (q *Query) ToJSON() string {
	var obj = map[string]interface{}{
		"query":     q,
		"condition": q.condition,
		"attrs":     q.attrs,
		"sort":      q.sort,
	}
	d, _ := json.Marshal(obj)
	return string(d)
}
