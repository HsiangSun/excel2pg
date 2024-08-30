package models

type Config struct {
	Id          int    `json:"id"`
	TableName   string `json:"table_name"`
	ColumnName  string `json:"column_name"`
	ColumnOrder int    `json:"column_order"`
}
