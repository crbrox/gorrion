package gorrion

import (
	"testing"
)

func TestQuery_Get_SimpleOne(t *testing.T) {

	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	var (
		id = EntityID{
			ID:          population[2].ID.ID,
			Type:        population[2].ID.Type,
			Service:     population[2].ID.Service,
			ServicePath: population[2].ID.ServicePath,
		}
		e2  = Entity{}
		q   = &Query{ID: []string{id.ID}, Type: []string{id.Type}}
		err error
	)

	ei, err := q.Get("S", "SP")

	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !ei.Next(&e2) {
		t.Fatalf("wanted true, got false (entity iter empty)")
	}
	if !equalObjects(population[2], e2) {
		t.Errorf("wanted %#v, got %#v", population[2], e2)
	}
	if ei.Next(&e2) {
		t.Errorf("wanted true, got false (entity iter not empty)")
	}
	if ei.Err() != nil {
		t.Fatal(unexpected(err))
	}
}

func TestQuery_Get_IDPattern(t *testing.T) {

	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	var (
		result = [3]Entity{}
		q      = &Query{IDPattern: "^E.*$", Type: []string{"T2"}}
		err    error
	)

	eIter, err := q.Get("S", "SP")
	if err != nil {
		t.Fatal(unexpected(err))
	}

	originSet := make([]*Entity, 3)
	copy(originSet, population[3:6])

result_loop:
	for i, ent := range result {
		if !eIter.Next(&ent) {
			t.Fatalf("wanted true, got false (entity iter empty) (%d) %v", i, eIter.Err())
		}
		for j, orig := range originSet {
			if equalObjects(ent, orig) {
				// cannot match again. 3 times the same entity is not we want
				originSet[j] = nil
				continue result_loop
			}
		}
		t.Fatalf("not found %+v (%d)", ent, i)
	}

	if eIter.Next(nil) {
		t.Errorf("wanted true, got false (entity iter not empty)")
	}
	if eIter.Err() != nil {
		t.Fatal(unexpected(err))
	}
}

func TestQuery_Get_TypePattern(t *testing.T) {

	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	var (
		result = [3]Entity{}
		q      = &Query{TypePattern: "^.*1$"}
		err    error
	)

	eIter, err := q.Get("S", "SP")
	if err != nil {
		t.Fatal(unexpected(err))
	}

	originSet := make([]*Entity, 3)
	copy(originSet, population[0:3])

result_loop:
	for i, ent := range result {
		if !eIter.Next(&ent) {
			t.Fatalf("wanted true, got false (entity iter empty) (%d) %v", i, eIter.Err())
		}
		for j, orig := range originSet {
			if equalObjects(ent, orig) {
				// cannot match again. 3 times the same entity is not we want
				originSet[j] = nil
				continue result_loop
			}
		}
		t.Fatalf("not found %+v (%d)", ent, i)
	}

	if eIter.Next(nil) {
		t.Errorf("wanted true, got false (entity iter not empty)")
	}
	if eIter.Err() != nil {
		t.Fatal(unexpected(err))
	}
}

func TestQuery_Get_Limit(t *testing.T) {

	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	var (
		result Entity
		q      = &Query{}
	)

	for limit := 1; limit <= len(population); limit++ {
		q.Limit = limit
		ei, err := q.Get("S", "SP")
		if err != nil {
			t.Fatal(unexpected(err))
		}

		var got = 0
		for ei.Next(&result) {
			got++
		}

		if ei.Err() != nil {
			t.Fatal(unexpected(err))
		}

		if got != limit {
			t.Fatalf("wanted %d, got %d", limit, got)
		}
	}

	// now limit greater than size of population
	for limit := len(population); limit <= len(population)*2; limit++ {
		q.Limit = limit
		ei, err := q.Get("S", "SP")
		if err != nil {
			t.Fatal(unexpected(err))
		}

		var got = 0
		for ei.Next(&result) {
			got++
		}

		if ei.Err() != nil {
			t.Fatal(unexpected(err))
		}

		if got != len(population) {
			t.Fatalf("wanted %d, got %d", len(population), got)
		}
	}
}

func TestQuery_Get_Offset(t *testing.T) {

	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	const limit = 3
	var q = &Query{Limit: limit,
		OrderBy: []string{"temperature"}, // same order as in population
	}

	for offset := 0; offset <= len(population); offset++ {
		var result = []*Entity{}

		q.Offset = offset
		ei, err := q.Get("S", "SP")
		if err != nil {
			t.Fatal(unexpected(err))
		}

		// Take all elements in cursor
		for ent := (&Entity{}); ei.Next(ent); ent = (&Entity{}) {
			result = append(result, ent)
		}
		if ei.Err() != nil {
			t.Fatalf("unexpected error %v", ei.Err())
		}

		{
			var expectedSize int
			if offset+limit < len(population) {
				expectedSize = limit
			} else {
				expectedSize = len(population) - offset
			}
			if !equalObjects(population[offset:offset+expectedSize], result) {
				t.Errorf("wanted %#v, got %#v", population[offset:offset+expectedSize], result)
			}
		}
	}
}

func TestQuery_Get_ReverseOrder(t *testing.T) {

	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	const limit = 3
	var q = &Query{
		OrderBy: []string{"!temperature"}, // reversed order in population
	}
	var result = []*Entity{}

	ei, err := q.Get("S", "SP")
	if err != nil {
		t.Fatal(unexpected(err))
	}

	// Take all elements in cursor
	for ent := (&Entity{}); ei.Next(ent); ent = (&Entity{}) {
		result = append(result, ent)
	}
	if ei.Err() != nil {
		t.Fatalf("unexpected error %v", ei.Err())
	}

	reversed := make([]*Entity, 0, len(population))
	for i := len(population); i > 0; i-- {
		reversed = append(reversed, population[i-1])
	}
	if !equalObjects(reversed, result) {
		t.Errorf("wanted %#v, got %#v", reversed, result)
	}

}

func TestQuery_Get_Attrs(t *testing.T) {

	setupTestDB(t)
	defer teardownTestDB(t)

	populateDB(t)

	const attribute = "status"
	var (
		id = EntityID{
			ID:          population[2].ID.ID,
			Type:        population[2].ID.Type,
			Service:     population[2].ID.Service,
			ServicePath: population[2].ID.ServicePath,
		}
		e2  = Entity{}
		q   = &Query{ID: []string{id.ID}, Type: []string{id.Type}, Attrs: []string{attribute}}
		err error
	)

	eIter, err := q.Get("S", "SP")

	if err != nil {
		t.Fatal(unexpected(err))
	}
	if !eIter.Next(&e2) {
		t.Fatalf("wanted true, got false (entity iter empty)")
	}
	if eIter.Err() != nil {
		t.Fatal(unexpected(err))
	}

	if len(e2.Attrs) != 1 {
		t.Errorf("wanted 1 attribute, got %d", len(e2.Attrs))
	}
	if status, ok := e2.Attrs[attribute]; !ok {
		t.Errorf("missing attribute %s", attribute)
	} else if !equalObjects(status, population[2].Attrs[attribute]) {
		t.Errorf("wanted %#v, got %#v", status, population[2].Attrs[attribute])
	}
}
