package graph

import "github.com/clementd64/tachiql/pkg/tachiql"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Indexer *tachiql.Indexer
}

func Float(f32 *float32) *float64 {
	if f32 == nil {
		return nil
	}

	f64 := float64(*f32)
	return &f64
}
