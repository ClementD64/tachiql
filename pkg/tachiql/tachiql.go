package tachiql

import (
	"context"

	"github.com/clementd64/tachiql/pkg/backup"
	"github.com/clementd64/tachiql/pkg/graph"
)

type Tachiql struct {
	plugins Plugins
	Graph   *graph.Graph
	Backup  *backup.Backup

	context    context.Context
	StopWorker context.CancelFunc
}

func New(g *graph.Graph, plugins Plugins) *Tachiql {
	ctx, cancel := context.WithCancel(context.Background())

	t := &Tachiql{
		plugins:    plugins,
		Graph:      g,
		context:    ctx,
		StopWorker: cancel,
	}

	t.plugins.Schema(g)
	return t
}

func (t *Tachiql) SetBackup(b *backup.Backup) error {
	if err := t.plugins.Backup(t, b); err != nil {
		return err
	}

	t.Backup = b
	t.plugins.Clean()
	return nil
}

func (t *Tachiql) StartWorker() error {
	t.plugins.Worker(t.context, t)
	return nil
}
