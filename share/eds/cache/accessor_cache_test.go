package cache

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/filecoin-project/dagstore"
	"github.com/filecoin-project/dagstore/shard"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

func TestAccessorCache(t *testing.T) {
	t.Run("add / get item from cache", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		cache, err := NewAccessorCache("test", 1)
		require.NoError(t, err)

		// add accessor to the cache
		key := shard.KeyFromString("key")
		mock := &mockAccessor{
			data: []byte("test_data"),
		}
		loaded, err := cache.GetOrLoad(ctx, key, func(ctx context.Context, key shard.Key) (Accessor, error) {
			return mock, nil
		})
		require.NoError(t, err)

		// check if item exists
		got, err := cache.Get(key)
		require.NoError(t, err)

		l, err := io.ReadAll(loaded.Reader())
		require.NoError(t, err)
		require.Equal(t, mock.data, l)
		g, err := io.ReadAll(got.Reader())
		require.NoError(t, err)
		require.Equal(t, mock.data, g)
	})

	t.Run("get blockstore from accessor", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		cache, err := NewAccessorCache("test", 1)
		require.NoError(t, err)

		// add accessor to the cache
		key := shard.KeyFromString("key")
		mock := &mockAccessor{}
		accessor, err := cache.GetOrLoad(ctx, key, func(ctx context.Context, key shard.Key) (Accessor, error) {
			return mock, nil
		})
		require.NoError(t, err)

		// check if item exists
		_, err = cache.Get(key)
		require.NoError(t, err)

		// blockstore should be created only after first request
		require.Equal(t, 0, mock.returnedBs)

		// try to get blockstore
		_, err = accessor.Blockstore()
		require.NoError(t, err)

		// second call to blockstore should return same blockstore
		_, err = accessor.Blockstore()
		require.NoError(t, err)
		require.Equal(t, 1, mock.returnedBs)
	})

	t.Run("remove an item", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		cache, err := NewAccessorCache("test", 1)
		require.NoError(t, err)

		// add accessor to the cache
		key := shard.KeyFromString("key")
		mock := &mockAccessor{}
		ac, err := cache.GetOrLoad(ctx, key, func(ctx context.Context, key shard.Key) (Accessor, error) {
			return mock, nil
		})
		require.NoError(t, err)
		err = ac.Close()
		require.NoError(t, err)

		err = cache.Remove(key)
		require.NoError(t, err)

		// accessor should be closed on removal
		mock.checkClosed(t, true)

		// check if item exists
		_, err = cache.Get(key)
		require.ErrorIs(t, err, errCacheMiss)
	})

	t.Run("successive reads should read the same data", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		cache, err := NewAccessorCache("test", 1)
		require.NoError(t, err)

		// add accessor to the cache
		key := shard.KeyFromString("key")
		mock := &mockAccessor{data: []byte("test")}
		accessor, err := cache.GetOrLoad(ctx, key, func(ctx context.Context, key shard.Key) (Accessor, error) {
			return mock, nil
		})
		require.NoError(t, err)

		loaded, err := io.ReadAll(accessor.Reader())
		require.NoError(t, err)
		require.Equal(t, mock.data, loaded)

		for i := 0; i < 2; i++ {
			accessor, err = cache.Get(key)
			require.NoError(t, err)
			got, err := io.ReadAll(accessor.Reader())
			require.NoError(t, err)
			require.Equal(t, mock.data, got)
		}
	})

	t.Run("removed by eviction", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		cache, err := NewAccessorCache("test", 1)
		require.NoError(t, err)

		// add accessor to the cache
		key := shard.KeyFromString("key")
		mock := &mockAccessor{}
		ac1, err := cache.GetOrLoad(ctx, key, func(ctx context.Context, key shard.Key) (Accessor, error) {
			return mock, nil
		})
		require.NoError(t, err)
		err = ac1.Close()
		require.NoError(t, err)

		// add second item
		key2 := shard.KeyFromString("key2")
		ac2, err := cache.GetOrLoad(ctx, key2, func(ctx context.Context, key shard.Key) (Accessor, error) {
			return mock, nil
		})
		require.NoError(t, err)
		err = ac2.Close()
		require.NoError(t, err)

		// accessor should be closed on removal by eviction
		mock.checkClosed(t, true)

		// check if item evicted
		_, err = cache.Get(key)
		require.ErrorIs(t, err, errCacheMiss)
	})

	t.Run("close on accessor is noop", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		cache, err := NewAccessorCache("test", 1)
		require.NoError(t, err)

		// add accessor to the cache
		key := shard.KeyFromString("key")
		mock := &mockAccessor{}
		_, err = cache.GetOrLoad(ctx, key, func(ctx context.Context, key shard.Key) (Accessor, error) {
			return mock, nil
		})
		require.NoError(t, err)

		// check if item exists
		accessor, err := cache.Get(key)
		require.NoError(t, err)
		require.NotNil(t, accessor)

		// close on returned accessor should not close inner reader
		err = accessor.Close()
		require.NoError(t, err)

		// check that close was not performed on inner accessor
		mock.checkClosed(t, false)
	})
}

type mockAccessor struct {
	data       []byte
	isClosed   bool
	returnedBs int
}

func (m *mockAccessor) Reader() io.Reader {
	return bytes.NewBuffer(m.data)
}

func (m *mockAccessor) Blockstore() (dagstore.ReadBlockstore, error) {
	if m.returnedBs > 0 {
		return nil, errors.New("blockstore already returned")
	}
	m.returnedBs++
	return rbsMock{}, nil
}

func (m *mockAccessor) Close() error {
	if m.isClosed {
		return errors.New("already closed")
	}
	m.isClosed = true
	return nil
}

func (m *mockAccessor) checkClosed(t *testing.T, expected bool) {
	// item will be removed in background, so give it some time to settle
	time.Sleep(time.Millisecond * 100)
	require.Equal(t, expected, m.isClosed)
}

// rbsMock is a dagstore.ReadBlockstore mock
type rbsMock struct{}

func (r rbsMock) Has(context.Context, cid.Cid) (bool, error) {
	panic("implement me")
}

func (r rbsMock) Get(_ context.Context, _ cid.Cid) (blocks.Block, error) {
	panic("implement me")
}

func (r rbsMock) GetSize(context.Context, cid.Cid) (int, error) {
	panic("implement me")
}

func (r rbsMock) AllKeysChan(context.Context) (<-chan cid.Cid, error) {
	panic("implement me")
}

func (r rbsMock) HashOnRead(bool) {
	panic("implement me")
}
