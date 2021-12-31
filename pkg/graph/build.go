package graph

import (
	"errors"
	"log"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
)

type Graph struct {
	Types  map[string]*graphql.Object
	Schema graphql.Schema
}

func New(obj interface{}) (g Graph, err error) {
	g.Types = map[string]*graphql.Object{}
	g.Schema, err = g.genSchema(obj)
	return
}

func (g *Graph) genSchema(obj interface{}) (graphql.Schema, error) {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return graphql.Schema{}, errors.New("schema root must be a struct")
	}

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: g.gen(t).(*graphql.Object),
	})
}

func (g *Graph) gen(t reflect.Type) graphql.Type {
	switch t.Kind() {
	case reflect.Ptr:
		return g.gen(t.Elem())
	case reflect.Slice:
		return graphql.NewList(g.gen(t.Elem()))
	case reflect.Struct:
		return g.genStruct(t)
	case reflect.String:
		return graphql.String
	case reflect.Int32:
		return graphql.Int
	case reflect.Int64:
		return Int64
	case reflect.Float32:
		return graphql.Float
	case reflect.Float64:
		return graphql.Float
	case reflect.Bool:
		return graphql.Boolean
	default:
		log.Print("Unknow ", t.Kind())
		return nil
	}
}

func (g *Graph) genStruct(t reflect.Type) graphql.Type {
	fields := g.genStructFields(t)

	obj := graphql.NewObject(graphql.ObjectConfig{
		Name:   t.Name(),
		Fields: fields,
	})

	g.Types[t.Name()] = obj
	return obj
}

func (g *Graph) genStructFields(t reflect.Type) graphql.Fields {
	fields := graphql.Fields{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name := getName(field); name != "" {
			fields[name] = &graphql.Field{
				Name: name,
				Type: g.gen(field.Type),
			}
		}
	}

	return fields
}

func getName(field reflect.StructField) string {
	json, ok := field.Tag.Lookup("json")
	if !ok {
		return ""
	}
	return strings.TrimSuffix(json, ",omitempty")
}

func ToMap(obj interface{}) map[string]interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	m := map[string]interface{}{}
	for i := 0; i < v.NumField(); i++ {
		if name := getName(v.Type().Field(i)); name != "" {
			m[name] = v.Field(i).Interface()
		}
	}

	return m
}
