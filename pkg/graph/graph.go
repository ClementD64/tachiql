package graph

import (
	"context"

	"github.com/graphql-go/graphql"
)

type Graph struct {
	Schema graphql.Schema
	Types  map[string]*graphql.Object
	Root   interface{}

	plugins    Plugins
	context    context.Context
	StopWorker context.CancelFunc
}

func New(obj interface{}, plugins Plugins) (*Graph, error) {
	schema, types, err := BuildGraph(obj)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	t := &Graph{
		Schema:     schema,
		Types:      types,
		plugins:    plugins,
		context:    ctx,
		StopWorker: cancel,
	}

	t.plugins.Schema(t)
	return t, nil
}

func (t *Graph) SetRoot(root interface{}) error {
	if err := t.plugins.Root(t, root); err != nil {
		return err
	}

	t.Root = root
	t.plugins.Clean()
	return nil
}

func (t *Graph) StartWorker() error {
	t.plugins.Worker(t.context, t.StopWorker, t)
	return nil
}
