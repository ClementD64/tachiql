package tachiql

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/clementd64/tachiql/pkg/backup"
	"github.com/clementd64/tachiql/pkg/graph"
)

type Plugin struct {
	Schema func(*graph.Graph) error              `plugin:""`
	Backup func(*Tachiql, *backup.Backup) error  `plugin:""`
	Clean  func()                                `plugin:""`
	Worker func(context.Context, *Tachiql) error `plugin:""`
}

func WrapPlugins(plugins []interface{}) (Plugins, error) {
	p := Plugins{}
	for _, plugin := range plugins {
		wrapper := Plugin{}
		if err := WrapPlugin(plugin, &wrapper); err != nil {
			return nil, err
		}
		p = append(p, wrapper)
	}
	return p, nil
}

type Plugins []Plugin

func (p *Plugins) Schema(g *graph.Graph) error {
	for _, plugin := range *p {
		if err := plugin.Schema(g); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugins) Backup(t *Tachiql, b *backup.Backup) error {
	for _, plugin := range *p {
		if err := plugin.Backup(t, b); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugins) Clean() {
	for _, plugin := range *p {
		plugin.Clean()
	}
}

func (p *Plugins) Worker(ctx context.Context, cancel context.CancelFunc, t *Tachiql) {
	wg := sync.WaitGroup{}
	for _, plugin := range *p {
		wg.Add(1)
		go func(plugin Plugin) {
			defer wg.Done()
			if err := plugin.Worker(ctx, t); err != nil {
				log.Print(err)
				cancel()
			}
		}(plugin)
	}
	wg.Wait()
}

func WrapPlugin(plugin interface{}, wrapper interface{}) error {
	p := reflect.ValueOf(plugin)

	if p.Kind() != reflect.Ptr && p.Elem().Kind() != reflect.Struct {
		return nil
	}

	w := reflect.ValueOf(wrapper).Elem()
	for i := 0; i < w.Type().NumField(); i++ {
		if flag, ok := w.Type().Field(i).Tag.Lookup("plugin"); ok {
			method, ok := method(p, w.Type().Field(i).Name)
			if !ok && flag == "required" {
				return errors.New("method " + w.Type().Field(i).Name + " is required")
			}
			if !ok {
				method = defaultMethod(w.Type().Field(i).Type)
			}
			if !method.Type().AssignableTo(w.Field(i).Type()) {
				return fmt.Errorf("invalid method %s (need %s, found %s)", w.Type().Field(i).Name, w.Field(i).Type(), method.Type())
			}
			w.Field(i).Set(method)
		}
	}

	return nil
}

func method(plugin reflect.Value, name string) (reflect.Value, bool) {
	method := plugin.MethodByName(name)
	if method.Kind() == reflect.Func {
		return method, true
	}

	method = plugin.Elem().FieldByName(name)
	if method.Kind() == reflect.Func {
		return method, true
	}

	return reflect.Value{}, false
}

func defaultMethod(t reflect.Type) reflect.Value {
	return reflect.MakeFunc(t, func(args []reflect.Value) (results []reflect.Value) {
		out := []reflect.Value{}
		for i := 0; i < t.NumOut(); i++ {
			out = append(out, reflect.New(t.Out(i)).Elem())
		}
		return out
	})
}
