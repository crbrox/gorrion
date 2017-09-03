package main

import (
	"fmt"
	"log"

	"github.com/crbrox/gorrion"
	"net/http"
)

func main() {
	var err error

	err = gorrion.StartStore()
	if err != nil {
		log.Fatal(err)
	}

	ei := gorrion.EntityID{ID: "E1", Type: "T", Service: "S", ServicePath: "SP"}
	e := gorrion.NewEntity(ei)
	e.Attrs["temperature"] = gorrion.Attribute{Value: 23.45, Type: "celsius", Md: map[string]interface{}{
		"location": "12,67",
		"date":     "12-12-1999",
	}}
	gorrion.CreateEntity(e)

	e2, err := gorrion.GetEntity(ei)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("e2 %#v\n", e2)

	log.Fatal(http.ListenAndServe(":9090", gorrion.AddHandlers()))

}
