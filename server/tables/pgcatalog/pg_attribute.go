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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgAttributeName is a constant to the pg_attribute name.
const PgAttributeName = "pg_attribute"

// InitPgAttribute handles registration of the pg_attribute handler.
func InitPgAttribute() {
	tables.AddHandler(PgCatalogName, PgAttributeName, PgAttributeHandler{})
}

// PgAttributeHandler is the handler for the pg_attribute table.
type PgAttributeHandler struct{}

var _ tables.Handler = PgAttributeHandler{}

// Name implements the interface tables.Handler.
func (p PgAttributeHandler) Name() string {
	return PgAttributeName
}

// emptyRowIter implements the sql.RowIter for empty table.
func emptyRowIter() (sql.RowIter, error) {
	return sql.RowsToRowIter(), nil
}

// RowIter implements the interface tables.Handler.
func (p PgAttributeHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	doltSession := dsess.DSessFromSess(ctx.Session)
	c := sqle.NewDefault(doltSession.Provider()).Analyzer.Catalog

	var cols []*sql.Column

	err := currentDatabaseSchemaIter(ctx, c, func(db sql.Database) (bool, error) {
		err := sql.DBTableIter(ctx, db, func(t sql.Table) (cont bool, err error) {
			for _, col := range t.Schema() {
				cols = append(cols, col)
			}
			return true, nil
		})
		if err != nil {
			return false, err
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return &pgAttributeRowIter{
		cols: cols,
		idx:  0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgAttributeHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAttributeSchema,
		PkOrdinals: nil,
	}
}

// pgAttributeSchema is the schema for pg_attribute.
var pgAttributeSchema = sql.Schema{
	{Name: "attrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atttypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attlen", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attnum", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcacheoff", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atttypmod", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attndims", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attbyval", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attalign", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attstorage", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcompression", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attnotnull", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atthasdef", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atthasmissing", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attidentity", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attgenerated", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attisdropped", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attislocal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attinhcount", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attstattarget", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcollation", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName},        // TODO: type aclitem[]
	{Name: "attoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName},    // TODO: collation C
	{Name: "attfdwoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName}, // TODO: collation C
	{Name: "attmissingval", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgAttributeName},
}

// pgAttributeRowIter is the sql.RowIter for the pg_attribute table.
type pgAttributeRowIter struct {
	cols []*sql.Column
	idx  int
}

var _ sql.RowIter = (*pgAttributeRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAttributeRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.cols) {
		return nil, io.EOF
	}
	iter.idx++
	col := iter.cols[iter.idx-1]

	generated := ""
	if col.Generated != nil {
		generated = "s"
	}

	dimensions := 0
	s, ok := col.Type.(sql.SetType)
	if ok {
		dimensions = int(s.NumberOfElements())
	}

	hasDefault := col.Default != nil

	// TODO: Fill in the rest of the pg_attribute columns
	return sql.Row{
		uint32(0),         // attrelid
		col.Name,          // attname
		uint32(0),         // atttypid
		int16(0),          // attlen
		int16(iter.idx),   // attnum
		int32(-1),         // attcacheoff
		int32(-1),         // atttypmod
		int16(dimensions), // attndims
		false,             // attbyval
		"i",               // attalign
		"p",               // attstorage
		"",                // attcompression
		!col.Nullable,     // attnotnull
		hasDefault,        // atthasdef
		false,             // atthasmissing
		"",                // attidentity
		generated,         // attgenerated
		false,             // attisdropped
		true,              // attislocal
		int16(0),          // attinhcount
		int16(-1),         // attstattarget
		uint32(0),         // attcollation
		nil,               // attacl
		nil,               // attoptions
		nil,               // attfdwoptions
		nil,               // attmissingval
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAttributeRowIter) Close(ctx *sql.Context) error {
	return nil
}