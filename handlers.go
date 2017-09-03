package gorrion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type object map[string]interface{}

const (
	paramType    = "type"
	paramOptions = "options"
	paramAttrs   = "attrs"
)

func AddHandlers() http.Handler {
	const (
		entitiesPrefix = "/v2/entities" // root router
		entity         = "/{id}"
		attributes     = entity + "/attrs"
		attribute      = attributes + "/{name}"
		attributeValue = attribute + "/value"
	)
	r := mux.NewRouter()
	r.StrictSlash(true)
	entR := r.PathPrefix(entitiesPrefix).Subrouter()

	// entities
	entR.HandleFunc("/", cH(getEntitiesHandleF)).Methods("GET")
	entR.HandleFunc("/", cH(postEntitiesHandleF)).Methods("POST")

	// entity
	entR.HandleFunc(entity, cH(getEntityHandleF)).Methods("GET")
	entR.HandleFunc(entity, cH(deleteEntityHandleF)).Methods("DELETE")

	// attrs
	entR.HandleFunc(attributes, cH(getAttrsHandleF)).Methods("GET")
	entR.HandleFunc(attributes, cH(postAttrsHandleF)).Methods("POST")
	entR.HandleFunc(attributes, cH(patchAttrsHandleF)).Methods("PATCH")
	entR.HandleFunc(attributes, cH(putAttrsHandleF)).Methods("PUT")

	// attr
	entR.HandleFunc(attribute, cH(getAttrHandleF)).Methods("GET")
	entR.HandleFunc(attribute, cH(deleteAttrHandleF)).Methods("DELETE")
	entR.HandleFunc(attribute, cH(putAttrHandleF)).Methods("PUT")

	// attrValue
	entR.HandleFunc(attributeValue, cH(getAttrValueHandleF)).Methods("GET")
	entR.HandleFunc(attributeValue, cH(putAttrValueHandleF)).Methods("PUT")

	return entR

}

func respondErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err, ok := err.(gorrionErr); ok {
		w.WriteHeader(err.Status())
	} else {
		w.WriteHeader(500)
	}
	fmt.Fprint(w, ErrToJSON(err))
}

type handlerArgs struct {
	ID      EntityID
	vars    map[string]string
	options OptionSet
	attrs   []string
	obj     object
	any     interface{}
	w       http.ResponseWriter
	req     *http.Request
}

func cH(f func(ctx context.Context, args handlerArgs) (interface{}, error)) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		args := handlerArgs{
			vars: mux.Vars(req),
			req:  req,
			w:    w,
		}

		ctx := req.Context()

		// options param
		optParam := req.FormValue(paramOptions)
		args.options = ParseOptSet(optParam)

		// attrs param
		attrsParam := req.FormValue(paramAttrs)
		args.attrs = strings.Split(attrsParam, ",")

		// incomming object
		if req.ContentLength > 0 {
			// use prefix, allow "; charset=utf-8", be liberal with input
			if !strings.HasPrefix(req.Header.Get("content-type"), "application/json") {
				respondErr(w, ErrContentTypeNotJSON)
				return
			}
			var any interface{}
			err := json.NewDecoder(req.Body).Decode(&any)
			if err != nil {
				respondErr(w, ErrParsingJSON)
				return
			}
			if obj, ok := any.(map[string]interface{}); ok {
				args.obj = obj
			}
			args.any = any
		}

		t := req.FormValue(paramType)
		if len(t) == 0 {
			args.vars[paramType] = defaultEntityType
		} else {
			args.vars[paramType] = t
		}

		args.ID = EntityID{ID: args.vars["id"], Type: args.vars[paramType]}

		result, err := f(ctx, args)
		if err != nil {
			respondErr(w, err)
			return
		}
		if result != nil {
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)
			if req.FormValue("pretty") == "on" {
				encoder.SetIndent("", "\t")
			}
			errJ := encoder.Encode(result)
			if errJ != nil {
				// loggear error, esto sí es más grave
			}
			return
		}
	}
}

func getEntitiesHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	args.w.Write([]byte("GET entities"))
	return nil, nil
}

func postEntitiesHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	var (
		e   *Entity
		err error
	)

	if args.obj == nil {
		return nil, ErrEmptyObject
	}

	if args.options.Get(OptKeyValues) {
		e, err = FromKeyValues(args.obj)
	} else {
		e, err = FromObject(args.obj)
	}
	if err != nil {
		return nil, err
	}
	if err := CreateEntity(e); err != nil {
		return nil, err
	}
	args.w.WriteHeader(201)
	return nil, nil
}

func getEntityHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	var result interface{} // object or []interface{}
	entity, err := GetEntityAttrs(args.ID, args.attrs)
	if err != nil {
		return nil, err
	}

	if args.options.Get(OptKeyValues) {
		result = entity.ToKeyValues() // map
	} else if args.options.Get(OptValues) {
		result, err = entity.ToValues(args.attrs) // slice
		if err != nil {
			return nil, err
		}
	} else {
		result = entity.ToObject()
	}

	return result, nil
}

func deleteEntityHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	err := DeleteEntity(args.ID)
	if err != nil {
		return nil, err
	}
	args.w.WriteHeader(204)
	return nil, nil
}

func getAttrsHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	return GetAllAttrs(args.ID)
}

func postAttrsHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	var (
		m   map[string]Attribute
		err error
	)
	if args.options.Get(OptKeyValues) {
		m = AttrsFromKeyValue(args.obj)
	} else {
		m, err = AttrsMapFromObject(args.obj)
	}
	if err != nil {
		return nil, err
	}
	if args.options.Get(OptAppend) {
		// strict append
		_, err = AddAttrs(args.ID, m)
	} else {
		_, err = AddOrUpdateAttrs(args.ID, m)
	}
	return nil, err
}

func patchAttrsHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	var (
		m   map[string]Attribute
		err error
	)
	if args.options.Get(OptKeyValues) {
		m = AttrsFromKeyValue(args.obj)
	} else {
		m, err = AttrsMapFromObject(args.obj)
	}
	if err != nil {
		return nil, err
	}
	_, err = UpdateAttrs(args.ID, m)
	return nil, err
}

func putAttrsHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	var (
		m   map[string]Attribute
		err error
	)
	if args.options.Get(OptKeyValues) {
		m = AttrsFromKeyValue(args.obj)
	} else {
		m, err = AttrsMapFromObject(args.obj)
	}
	if err != nil {
		return nil, err
	}
	_, err = SetAllAttrs(args.ID, m)
	return nil, err
}

func getAttrHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	name := args.vars["name"]
	attr, err := GetAttr(args.ID, name)
	if err != nil {
		return nil, err
	}
	return map[string]Attribute{name: attr}, nil
}

func putAttrHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	name := args.vars["name"]
	attr, err := AttrValueFromObject(args.obj)
	if err != nil {
		return nil, err
	}
	_, err = SetAttr(args.ID, name, attr)
	return nil, err
}

func deleteAttrHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	name := args.vars["name"]
	_, err := DeleteAttr(args.ID, name)
	return nil, err
}

func getAttrValueHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	name := args.vars["name"]
	attr, err := GetAttr(args.ID, name)
	if err != nil {
		return nil, err
	}
	return attr.Value, nil
}

func putAttrValueHandleF(ctx context.Context, args handlerArgs) (interface{}, error) {
	name := args.vars["name"]
	_, err := SetAttr(args.ID, name, &Attribute{Value: args.any})
	return nil, err
}
