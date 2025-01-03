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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/functions/framework"

	pgtransform "github.com/dolthub/doltgresql/server/transform"
)

// OptimizeFunctions replaces all functions that fit specific criteria with their optimized variants.
func OptimizeFunctions(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// This is supposed to be one of the last rules to run. Subqueries break that assumption, so we skip this rule in such cases.
	if scope != nil && scope.CurrentNodeIsFromSubqueryExpression {
		return node, transform.SameTree, nil
	}
	return pgtransform.NodeExprsWithNodeWithOpaque(node, func(node sql.Node, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		if compiledFunction, ok := expr.(*framework.CompiledFunction); ok {
			if quickFunction := compiledFunction.GetQuickFunction(); quickFunction != nil {
				return quickFunction, transform.NewTree, nil
			}
		}
		return expr, transform.SameTree, nil
	})
}
