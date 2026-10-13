package test

import (
	. "adcompro-backend"
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

func TestReportMigration020Applies(t *testing.T) {
	if os.Getenv("RUN_AUCTION_INTEGRATION") != "1" {
		t.Skip("set RUN_AUCTION_INTEGRATION=1 to run database integration test")
	}

	ConnectDatabase()
	t.Cleanup(func() { DB.Close() })
	ctx := context.Background()

	var exists bool
	if err := DB.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conname = 'reports_exactly_one_target_check'
		)
	`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		return
	}

	raw, err := os.ReadFile("../db/migrations/020_reports_moderation_constraints.sql")
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf})
	sql := strings.TrimSpace(string(raw))
	sql = strings.TrimSpace(strings.TrimPrefix(sql, "BEGIN;"))
	sql = strings.TrimSpace(strings.TrimSuffix(sql, "COMMIT;"))

	tx, err := DB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, sql); err != nil {
		t.Fatalf("migration 020 cannot be applied cleanly: %v", err)
	}
}
