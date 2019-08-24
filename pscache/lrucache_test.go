package pscache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgconn/pscache"

	"github.com/stretchr/testify/require"
)

func TestLRUCacheModePrepare(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgconn.Connect(ctx, os.Getenv("PGX_TEST_CONN_STRING"))
	require.NoError(t, err)
	defer conn.Close(ctx)

	cache := pscache.NewLRUCache(conn, pscache.ModePrepare, 2)
	require.EqualValues(t, 0, cache.Len())
	require.EqualValues(t, 2, cache.Cap())
	require.EqualValues(t, pscache.ModePrepare, cache.Mode())

	psd, err := cache.Get(ctx, "select 1")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 1, cache.Len())
	require.ElementsMatch(t, []string{"select 1"}, fetchServerStatements(t, ctx, conn))

	psd, err = cache.Get(ctx, "select 1")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 1, cache.Len())
	require.ElementsMatch(t, []string{"select 1"}, fetchServerStatements(t, ctx, conn))

	psd, err = cache.Get(ctx, "select 2")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 2, cache.Len())
	require.ElementsMatch(t, []string{"select 1", "select 2"}, fetchServerStatements(t, ctx, conn))

	psd, err = cache.Get(ctx, "select 3")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 2, cache.Len())
	require.ElementsMatch(t, []string{"select 2", "select 3"}, fetchServerStatements(t, ctx, conn))

	err = cache.Clear(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 0, cache.Len())
	require.Empty(t, fetchServerStatements(t, ctx, conn))
}

func TestLRUCacheModeDescribe(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgconn.Connect(ctx, os.Getenv("PGX_TEST_CONN_STRING"))
	require.NoError(t, err)
	defer conn.Close(ctx)

	cache := pscache.NewLRUCache(conn, pscache.ModeDescribe, 2)
	require.EqualValues(t, 0, cache.Len())
	require.EqualValues(t, 2, cache.Cap())
	require.EqualValues(t, pscache.ModeDescribe, cache.Mode())

	psd, err := cache.Get(ctx, "select 1")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 1, cache.Len())
	require.Empty(t, fetchServerStatements(t, ctx, conn))

	psd, err = cache.Get(ctx, "select 1")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 1, cache.Len())
	require.Empty(t, fetchServerStatements(t, ctx, conn))

	psd, err = cache.Get(ctx, "select 2")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 2, cache.Len())
	require.Empty(t, fetchServerStatements(t, ctx, conn))

	psd, err = cache.Get(ctx, "select 3")
	require.NoError(t, err)
	require.NotNil(t, psd)
	require.EqualValues(t, 2, cache.Len())
	require.Empty(t, fetchServerStatements(t, ctx, conn))

	err = cache.Clear(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 0, cache.Len())
	require.Empty(t, fetchServerStatements(t, ctx, conn))
}

func fetchServerStatements(t testing.TB, ctx context.Context, conn *pgconn.PgConn) []string {
	result := conn.ExecParams(ctx, `select statement from pg_prepared_statements`, nil, nil, nil, nil).Read()
	require.NoError(t, result.Err)
	var statements []string
	for _, r := range result.Rows {
		statements = append(statements, string(r[0]))
	}
	return statements
}
