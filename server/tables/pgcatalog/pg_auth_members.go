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

package pgcatalog

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgAuthMembersName is a constant to the pg_auth_members name.
const PgAuthMembersName = "pg_auth_members"

// InitPgAuthMembers handles registration of the pg_auth_members handler.
func InitPgAuthMembers() {
	tables.AddHandler(PgCatalogName, PgAuthMembersName, PgAuthMembersHandler{})
}

// PgAuthMembersHandler is the handler for the pg_auth_members table.
type PgAuthMembersHandler struct{}

var _ tables.Handler = PgAuthMembersHandler{}

// Name implements the interface tables.Handler.
func (p PgAuthMembersHandler) Name() string {
	return PgAuthMembersName
}

// RowIter implements the interface tables.Handler.
func (p PgAuthMembersHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_auth_members row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAuthMembersHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAuthMembersSchema,
		PkOrdinals: nil,
	}
}

// pgAuthMembersSchema is the schema for pg_auth_members.
var pgAuthMembersSchema = sql.Schema{
	{Name: "roleid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAuthMembersName},
	{Name: "member", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAuthMembersName},
	{Name: "grantor", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAuthMembersName},
	{Name: "admin_option", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthMembersName},
}

// pgAuthMembersRowIter is the sql.RowIter for the pg_auth_members table.
type pgAuthMembersRowIter struct {
}

var _ sql.RowIter = (*pgAuthMembersRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAuthMembersRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAuthMembersRowIter) Close(ctx *sql.Context) error {
	return nil
}