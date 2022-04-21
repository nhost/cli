package hasura

import "encoding/json"

type QualifiedTable struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type TableEntry struct {
	IsEnum *bool          `json:"is_enum,omitempty"`
	Table  QualifiedTable `json:"table"`
}

type Source struct {
	Name   string       `json:"name"`
	Tables []TableEntry `json:"tables"`
}

type MetadataV3 struct {
	Sources []Source `json:"sources"`
}

//    objRelUsing, err := UnmarshalObjRelUsing(bytes)
//    bytes, err = objRelUsing.Marshal()
//
//    objRelUsingManualMapping, err := UnmarshalObjRelUsingManualMapping(bytes)
//    bytes, err = objRelUsingManualMapping.Marshal()

func UnmarshalObjRelUsing(data []byte) (ObjRelUsing, error) {
	var r ObjRelUsing
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ObjRelUsing) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalObjRelUsingManualMapping(data []byte) (ObjRelUsingManualMapping, error) {
	var r ObjRelUsingManualMapping
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ObjRelUsingManualMapping) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

//
// https://hasura.io/docs/latest/graphql/core/api-reference/schema-metadata-api/relationship.html#args-syntax
type ObjectRelationship struct {
	Comment *string     `json:"comment,omitempty"` // Comment
	Name    string      `json:"name"`              // Name of the new relationship
	Using   ObjRelUsing `json:"using"`             // Use one of the available ways to define an object relationship
}

type ForeignKeyConstraintOn struct {
	Table  string
	Column string
	wire   struct {
		Table  string `json:"table"`
		Column string `json:"column"`
	}
}

// Use one of the available ways to define an object relationship
//
// Use one of the available ways to define an object relationship
//
// https://hasura.io/docs/latest/graphql/core/api-reference/schema-metadata-api/relationship.html#objrelusing
type ObjRelUsing struct {
	ForeignKeyConstraintOn *ForeignKeyConstraintOn   `json:"foreign_key_constraint_on,omitempty"` // The column with foreign key constraint
	ManualConfiguration    *ObjRelUsingManualMapping `json:"manual_configuration,omitempty"`      // Manual mapping of table and columns
}

// Manual mapping of table and columns
//
// Manual mapping of table and columns
//
// https://hasura.io/docs/latest/graphql/core/api-reference/schema-metadata-api/relationship.html#objrelusingmanualmapping
type ObjRelUsingManualMapping struct {
	ColumnMapping map[string]string `json:"column_mapping"` // Mapping of columns from current table to remote table
	RemoteTable   *TableName        `json:"remote_table"`   // The table to which the relationship has to be established
}

func UnmarshalTableName(data []byte) (TableName, error) {
	var r TableName
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TableName) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type TableName struct {
	QualifiedTable *QualifiedTable
	String         *string
}
