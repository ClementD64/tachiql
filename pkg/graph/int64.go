package graph

import (
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

var Int64 = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Int64",
	Description: "64 bit integer",
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case int64:
			return value
		case *int64:
			if value == nil {
				return nil
			}
			return *value
		default:
			return nil
		}
	},
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case int64:
			return value
		case *int64:
			if value == nil {
				return nil
			}
			return *value
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.IntValue:
			v, err := strconv.ParseInt(valueAST.Value, 10, 64)
			if err != nil {
				return nil
			}
			return v
		default:
			return nil
		}
	},
})
