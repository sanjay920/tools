package cmd

import (
	"context"
	"database/sql"
	"fmt"
)

// ExecDatabaseStatement executes a SQL statement (e.g., INSERT, UPDATE, DELETE, CREATE) and returns a status message.
func ExecDatabaseStatement(ctx context.Context, db *sql.DB, stmt string) (string, error) {
	_, err := db.ExecContext(ctx, stmt)
	if err != nil {
		return "", fmt.Errorf("error executing SQL: %w", err)
	}
	return "Command executed successfully.", nil
}
