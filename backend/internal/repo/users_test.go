package repo

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

// When no users row exists, GetCategories must seed DefaultDispatchCategories
// and return them so the Settings dialog shows the defaults on first open.
func TestUsersRepo_GetCategories_seedsDefaultsOnErrNoRows(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectQuery(`SELECT dispatch_categories`).
		WithArgs("alice").
		WillReturnError(pgx.ErrNoRows)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("alice", pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	r := NewUsersRepo(mock)

	got, err := r.GetCategories(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	want := DefaultDispatchCategories
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("idx %d: want %q, got %q", i, want[i], got[i])
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
