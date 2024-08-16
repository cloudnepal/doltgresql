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

package functions

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dprocedures"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initDoltProcedures() {
	for _, procDef := range dprocedures.DoltProcedures {
		p, err := resolveExternalStoredProcedure(nil, procDef)
		if err != nil {
			panic(err)
		}

		funcVal := reflect.ValueOf(procDef.Function)
		callable := callableForDoltProcedure(p, funcVal)

		framework.RegisterFunction(framework.Function1{
			Name:       procDef.Name,
			Return:     pgtypes.TextArray,
			Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray},
			Variadic:   true,
			Callable:   callable,
		})
	}
}

func callableForDoltProcedure(p *plan.ExternalProcedure, funcVal reflect.Value) func(ctx *sql.Context, paramsAndReturn [2]pgtypes.DoltgresType, val1 any) (any, error) {
	funcType := funcVal.Type()

	return func(ctx *sql.Context, paramsAndReturn [2]pgtypes.DoltgresType, val1 any) (any, error) {
		values, ok := val1.([]any)
		if !ok {
			return nil, sql.ErrExternalProcedureInvalidParamType.New(reflect.TypeOf(val1).String())
		}

		funcParams := make([]reflect.Value, len(values)+1)
		funcParams[0] = reflect.ValueOf(ctx)

		for i := range values {
			paramDefinition := p.ParamDefinitions[0]
			var funcParamType reflect.Type
			if paramDefinition.Variadic {
				funcParamType = funcType.In(funcType.NumIn() - 1).Elem()
			} else {
				// TODO: support non-variadic procedures
				return nil, sql.ErrExternalProcedureInvalidParamType.New(funcType.String())
			}

			// Grab the passed-in variable and convert it to the type we expect
			exprParamVal, _, err := paramDefinition.Type.Convert(values[i])
			if err != nil {
				return nil, err
			}

			funcParams[i+1], err = p.ProcessParam(ctx, funcParamType, exprParamVal)
			if err != nil {
				return nil, err
			}
		}

		out := funcVal.Call(funcParams)
		if err, ok := out[1].Interface().(error); ok { // Only evaluates to true when error is not nil
			return nil, err
		}

		var rowIter sql.RowIter
		if iter, ok := out[0].Interface().(sql.RowIter); ok {
			rowIter = iter
		} else {
			rowIter = sql.RowsToRowIter()
		}

		return drainRowIter(ctx, rowIter)
	}
}

func drainRowIter(ctx *sql.Context, rowIter sql.RowIter) (any, error) {
	defer rowIter.Close(ctx)

	row, err := rowIter.Next(ctx)
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// The conversion to []text needs []any, not sql.Row
	rowSlice := make([]any, len(row))
	for i := range row {
		fromType, err := typeForElement(row[i])
		if err != nil {
			return nil, err
		}

		castFn := framework.GetExplicitCast(fromType, pgtypes.Text.BaseID())
		textVal, err := castFn(ctx, row[i], pgtypes.Text)
		if err != nil {
			return nil, err
		}

		rowSlice[i] = textVal
	}
	return rowSlice, nil
}

func typeForElement(v any) (pgtypes.DoltgresTypeBaseID, error) {
	switch x := v.(type) {
	case int64:
		return pgtypes.Int64.BaseID(), nil
	case int32:
		return pgtypes.Int32.BaseID(), nil
	case int16, int8:
		return pgtypes.Int16.BaseID(), nil
	case string:
		return pgtypes.Text.BaseID(), nil
	default:
		return 0, fmt.Errorf("dolt_procedures: unsupported type %T", x)
	}
}

var (
	// ctxType is the reflect.Type of a *sql.Context.
	ctxType = reflect.TypeOf((*sql.Context)(nil))
	// ctxType is the reflect.Type of a sql.RowIter.
	rowIterType = reflect.TypeOf((*sql.RowIter)(nil)).Elem()
	// ctxType is the reflect.Type of an error.
	errorType = reflect.TypeOf((*error)(nil)).Elem()
	// externalStoredProcedurePointerTypes maps a non-pointer type to a sql.Type for external stored procedures.
	externalStoredProcedureTypes = map[reflect.Type]sql.Type{
		reflect.TypeOf(int(0)):      types.Int64,
		reflect.TypeOf(int8(0)):     types.Int8,
		reflect.TypeOf(int16(0)):    types.Int16,
		reflect.TypeOf(int32(0)):    types.Int32,
		reflect.TypeOf(int64(0)):    types.Int64,
		reflect.TypeOf(uint(0)):     types.Uint64,
		reflect.TypeOf(uint8(0)):    types.Uint8,
		reflect.TypeOf(uint16(0)):   types.Uint16,
		reflect.TypeOf(uint32(0)):   types.Uint32,
		reflect.TypeOf(uint64(0)):   types.Uint64,
		reflect.TypeOf(float32(0)):  types.Float32,
		reflect.TypeOf(float64(0)):  types.Float64,
		reflect.TypeOf(bool(false)): types.Int8,
		reflect.TypeOf(string("")):  types.LongText,
		reflect.TypeOf([]byte{}):    types.LongBlob,
		reflect.TypeOf(time.Time{}): types.DatetimeMaxPrecision,
	}
	// externalStoredProcedurePointerTypes maps a pointer type to a sql.Type for external stored procedures.
	externalStoredProcedurePointerTypes = map[reflect.Type]sql.Type{
		reflect.TypeOf((*int)(nil)):       types.Int64,
		reflect.TypeOf((*int8)(nil)):      types.Int8,
		reflect.TypeOf((*int16)(nil)):     types.Int16,
		reflect.TypeOf((*int32)(nil)):     types.Int32,
		reflect.TypeOf((*int64)(nil)):     types.Int64,
		reflect.TypeOf((*uint)(nil)):      types.Uint64,
		reflect.TypeOf((*uint8)(nil)):     types.Uint8,
		reflect.TypeOf((*uint16)(nil)):    types.Uint16,
		reflect.TypeOf((*uint32)(nil)):    types.Uint32,
		reflect.TypeOf((*uint64)(nil)):    types.Uint64,
		reflect.TypeOf((*float32)(nil)):   types.Float32,
		reflect.TypeOf((*float64)(nil)):   types.Float64,
		reflect.TypeOf((*bool)(nil)):      types.Int8,
		reflect.TypeOf((*string)(nil)):    types.LongText,
		reflect.TypeOf((*[]byte)(nil)):    types.LongBlob,
		reflect.TypeOf((*time.Time)(nil)): types.DatetimeMaxPrecision,
	}
)

func init() {
	if strconv.IntSize == 32 {
		externalStoredProcedureTypes[reflect.TypeOf(int(0))] = types.Int32
		externalStoredProcedureTypes[reflect.TypeOf(uint(0))] = types.Uint32
		externalStoredProcedurePointerTypes[reflect.TypeOf((*int)(nil))] = types.Int32
		externalStoredProcedurePointerTypes[reflect.TypeOf((*uint)(nil))] = types.Uint32
	}
}

func resolveExternalStoredProcedure(_ *sql.Context, externalProcedure sql.ExternalStoredProcedureDetails) (*plan.ExternalProcedure, error) {
	funcVal := reflect.ValueOf(externalProcedure.Function)
	funcType := funcVal.Type()
	if funcType.Kind() != reflect.Func {
		return nil, sql.ErrExternalProcedureNonFunction.New(externalProcedure.Function)
	}
	if funcType.NumIn() == 0 {
		return nil, sql.ErrExternalProcedureMissingContextParam.New()
	}
	if funcType.NumOut() != 2 {
		return nil, sql.ErrExternalProcedureReturnTypes.New()
	}
	if funcType.In(0) != ctxType {
		return nil, sql.ErrExternalProcedureMissingContextParam.New()
	}
	if funcType.Out(0) != rowIterType {
		return nil, sql.ErrExternalProcedureFirstReturn.New()
	}
	if funcType.Out(1) != errorType {
		return nil, sql.ErrExternalProcedureSecondReturn.New()
	}
	funcIsVariadic := funcType.IsVariadic()

	paramDefinitions := make([]plan.ProcedureParam, funcType.NumIn()-1)
	paramReferences := make([]*expression.ProcedureParam, len(paramDefinitions))
	for i := 0; i < len(paramDefinitions); i++ {
		funcParamType := funcType.In(i + 1)
		paramName := "A" + strconv.FormatInt(int64(i), 10)
		paramIsVariadic := false
		if funcIsVariadic && i == len(paramDefinitions)-1 {
			paramIsVariadic = true
			funcParamType = funcParamType.Elem()
			if funcParamType.Kind() == reflect.Ptr {
				return nil, sql.ErrExternalProcedurePointerVariadic.New()
			}
		}

		if sqlType, ok := externalStoredProcedureTypes[funcParamType]; ok {
			paramDefinitions[i] = plan.ProcedureParam{
				Direction: plan.ProcedureParamDirection_In,
				Name:      paramName,
				Type:      sqlType,
				Variadic:  paramIsVariadic,
			}
			paramReferences[i] = expression.NewProcedureParam(paramName, sqlType)
		} else if sqlType, ok = externalStoredProcedurePointerTypes[funcParamType]; ok {
			paramDefinitions[i] = plan.ProcedureParam{
				Direction: plan.ProcedureParamDirection_Inout,
				Name:      paramName,
				Type:      sqlType,
				Variadic:  paramIsVariadic,
			}
			paramReferences[i] = expression.NewProcedureParam(paramName, sqlType)
		} else {
			return nil, sql.ErrExternalProcedureInvalidParamType.New(funcParamType.String())
		}
	}

	return &plan.ExternalProcedure{
		ExternalStoredProcedureDetails: externalProcedure,
		ParamDefinitions:               paramDefinitions,
		Params:                         paramReferences,
	}, nil
}