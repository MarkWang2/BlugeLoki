package downloads

import (
	"context"
	"fmt"
	"github.com/MarkWang2/loki/pkg/storage/stores/shipper/bluge_db"
	"github.com/MarkWang2/loki/pkg/storage/stores/shipper/util"
	segment "github.com/blugelabs/bluge_segment_api"
	"github.com/cortexproject/cortex/pkg/util/spanlogger"
	"os"
	"sync"
	"time"

	chunk_util "github.com/cortexproject/cortex/pkg/chunk/util"
	pkg_util "github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const cacheCleanupInterval = time.Hour

type Config struct {
	CacheDir     string
	SyncInterval time.Duration
	CacheTTL     time.Duration
}

type TableManager struct {
	cfg           Config
	storageClient StorageClient

	tables    map[string]*Table
	tablesMtx sync.RWMutex
	metrics   *metrics

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// 将object store 数据定时缓存到本地，提供查询接口供查询
// 根据 最后被使用时间（被查询），计算ttl实现缓存的自动过期。
func NewTableManager(cfg Config, storageClient StorageClient, registerer prometheus.Registerer) (*TableManager, error) {
	// cleanup existing directory and re-create it since we do not use existing files in it.
	if err := os.RemoveAll(cfg.CacheDir); err != nil {
		return nil, err
	}

	if err := chunk_util.EnsureDirectory(cfg.CacheDir); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	tm := &TableManager{
		cfg:           cfg,
		storageClient: storageClient,
		tables:        make(map[string]*Table),
		metrics:       newMetrics(registerer),
		ctx:           ctx,
		cancel:        cancel,
	}

	// tm.getOrCreateTable(context.Background(), "snpseg")
	go tm.loop()
	return tm, nil
}

func (tm *TableManager) loop() {
	tm.wg.Add(1)
	defer tm.wg.Done()

	syncTicker := time.NewTicker(tm.cfg.SyncInterval)
	defer syncTicker.Stop()

	cacheCleanupTicker := time.NewTicker(cacheCleanupInterval)
	defer cacheCleanupTicker.Stop()

	for {
		select {
		case <-syncTicker.C:
			err := tm.syncTables(tm.ctx)
			if err != nil {
				level.Error(pkg_util.Logger).Log("msg", "error syncing local boltdb files with storage", "err", err)
			}
		case <-cacheCleanupTicker.C:
			err := tm.cleanupCache()
			if err != nil {
				level.Error(pkg_util.Logger).Log("msg", "error cleaning up expired tables", "err", err)
			}
		case <-tm.ctx.Done():
			return
		}
	}
}

func (tm *TableManager) Stop() {
	tm.cancel()
	tm.wg.Wait()

	tm.tablesMtx.Lock()
	defer tm.tablesMtx.Unlock()

	for _, table := range tm.tables {
		table.Close()
	}
}

func (tm *TableManager) QueryPages(ctx context.Context, queries []bluge_db.IndexQuery, callback segment.StoredFieldVisitor) error {
	queriesByTable := util.QueriesByTable(queries)
	for tableName, queries := range queriesByTable {
		err := tm.query(ctx, tableName, queries, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tm *TableManager) query(ctx context.Context, tableName string, queries []bluge_db.IndexQuery, callback segment.StoredFieldVisitor) error {
	log, ctx := spanlogger.New(ctx, "Shipper.Downloads.Query")
	defer log.Span.Finish()

	level.Debug(log).Log("table-name", tableName)

	table := tm.getOrCreateTable(ctx, tableName)

	err := util.DoParallelQueries(ctx, table, queries, callback)
	if err != nil {
		if table.Err() != nil {
			// table is in invalid state, remove the table so that next queries re-create it.
			tm.tablesMtx.Lock()
			defer tm.tablesMtx.Unlock()

			level.Error(pkg_util.Logger).Log("msg", fmt.Sprintf("table %s has some problem, cleaning it up", tableName), "err", table.Err())

			delete(tm.tables, tableName)
			return table.Err()
		}
	}

	return err
}

func (tm *TableManager) getOrCreateTable(spanCtx context.Context, tableName string) *Table {
	// if table is already there, use it.
	tm.tablesMtx.RLock()
	table, ok := tm.tables[tableName]
	tm.tablesMtx.RUnlock()

	if !ok {
		tm.tablesMtx.Lock()
		// check if some other competing goroutine got the lock before us and created the table, use it if so.
		table, ok = tm.tables[tableName]
		if !ok {
			// table not found, creating one.
			level.Info(pkg_util.Logger).Log("msg", fmt.Sprintf("downloading all files for table %s", tableName))

			table = NewTable(spanCtx, tableName, tm.cfg.CacheDir, tm.storageClient, tm.metrics)
			tm.tables[tableName] = table
		}
		tm.tablesMtx.Unlock()
	}

	return table
}

func (tm *TableManager) syncTables(ctx context.Context) error {
	tm.tablesMtx.RLock()
	defer tm.tablesMtx.RUnlock()

	level.Info(pkg_util.Logger).Log("msg", "syncing tables")

	for _, table := range tm.tables {
		err := table.Sync(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tm *TableManager) cleanupCache() error {
	tm.tablesMtx.Lock()
	defer tm.tablesMtx.Unlock()

	level.Info(pkg_util.Logger).Log("msg", "cleaning tables cache")

	for name, table := range tm.tables {
		lastUsedAt := table.LastUsedAt()
		if lastUsedAt.Add(tm.cfg.CacheTTL).Before(time.Now()) {
			level.Info(pkg_util.Logger).Log("msg", fmt.Sprintf("cleaning up expired table %s", name))
			err := table.CleanupAllDBs()
			if err != nil {
				return err
			}

			delete(tm.tables, name)
		}
	}

	return nil
}
