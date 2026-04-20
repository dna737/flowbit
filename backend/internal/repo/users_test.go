package repo

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

// When no users row exists, GetCategories must seed defaults from the dispatcher config
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

	mock.ExpectQuery(`SELECT allowed_job_types`).
		WithArgs(dispatcherConfigSingletonID).
		WillReturnRows(mock.NewRows([]string{"allowed_job_types"}).
			AddRow([]byte(`["echo","email","image_resize","url_scrape","fail"]`)))

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("alice", pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	cfg := &DispatcherConfigRepo{pool: mock}
	r := NewUsersRepo(mock, cfg)

	got, err := r.GetCategories(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	want := []string{"echo", "email", "image_resize", "url_scrape", "fail"}
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
