// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unary

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/duration"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '-' AND o.oprleft = 0 ORDER BY o.oprcode::varchar;

// initUnaryMinus registers the functions to the catalog.
func initUnaryMinus() {
	framework.RegisterUnaryFunction(framework.Operator_UnaryMinus, float4um)
	framework.RegisterUnaryFunction(framework.Operator_UnaryMinus, float8um)
	framework.RegisterUnaryFunction(framework.Operator_UnaryMinus, int2um)
	framework.RegisterUnaryFunction(framework.Operator_UnaryMinus, int4um)
	framework.RegisterUnaryFunction(framework.Operator_UnaryMinus, int8um)
	framework.RegisterUnaryFunction(framework.Operator_UnaryMinus, interval_um)
	framework.RegisterUnaryFunction(framework.Operator_UnaryMinus, numeric_uminus)
}

// float4um represents the PostgreSQL function of the same name, taking the same parameters.
var float4um = framework.Function1{
	Name:       "float4um",
	Return:     pgtypes.Float32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return -(val1.(float32)), nil
	},
}

// float8um represents the PostgreSQL function of the same name, taking the same parameters.
var float8um = framework.Function1{
	Name:       "float8um",
	Return:     pgtypes.Float64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return -(val1.(float64)), nil
	},
}

// int2um represents the PostgreSQL function of the same name, taking the same parameters.
var int2um = framework.Function1{
	Name:       "int2um",
	Return:     pgtypes.Int16,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return -(val1.(int16)), nil
	},
}

// int4um represents the PostgreSQL function of the same name, taking the same parameters.
var int4um = framework.Function1{
	Name:       "int4um",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return -(val1.(int32)), nil
	},
}

// int8um represents the PostgreSQL function of the same name, taking the same parameters.
var int8um = framework.Function1{
	Name:       "int8um",
	Return:     pgtypes.Int64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return -(val1.(int64)), nil
	},
}

// interval_um represents the PostgreSQL function of the same name, taking the same parameters.
var interval_um = framework.Function1{
	Name:       "interval_um",
	Return:     pgtypes.Interval,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		dur := val1.(duration.Duration)
		return dur.Mul(-1), nil
	},
}

// numeric_uminus represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_uminus = framework.Function1{
	Name:       "numeric_uminus",
	Return:     pgtypes.Numeric,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return val1.(decimal.Decimal).Neg(), nil
	},
}
