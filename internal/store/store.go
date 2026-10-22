package store

import (
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/logger"
	"context"
	"fmt"
)

type baseStore struct {
	client *db.Client
}

func newBaseStore(client *db.Client) *baseStore {
	return &baseStore{client: client}
}

type txKey struct{}

func txFromContext(ctx context.Context) (*db.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*db.Tx)
	return tx, ok
}

func (s *baseStore) Client(ctx context.Context) *db.Client {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Client()
	}
	return s.client
}

type Transactor struct {
	client *db.Client
}

func NewTransactor(client *db.Client) *Transactor {
	return &Transactor{client: client}
}

func (t *Transactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// 如果已经在事务上下文中，则直接复用当前事务
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}

	// 开启新事务
	tx, err := t.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}

	// 将事务注入到 Context 中
	txCtx := context.WithValue(ctx, txKey{}, tx)

	// 使用标志位来追踪事务是否正常结束，防止重复 Rollback 或漏掉 Rollback
	committed := false

	defer func() {
		if p := recover(); p != nil {
			if rerr := tx.Rollback(); rerr != nil {
				logger.Error("发生 panic 后回滚事务失败: %v (panic: %v)", rerr, p)
			}
			panic(p) // 重新抛出 panic
		} else if !committed {
			// 如果没有正常提交，说明中途退出了（比如业务返回了 err），执行回滚
			if rerr := tx.Rollback(); rerr != nil {
				logger.Error("回滚事务失败: %v", rerr)
			}
		}
	}()

	// 执行传入的业务逻辑
	if err := fn(txCtx); err != nil {
		return err // 错误会触发上面的 defer 执行 Rollback
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	committed = true // 标记已提交
	return nil
}

func paginate(total, page, perPage int) (offset, limit int) {
	if perPage <= 0 {
		perPage = 10
	}
	if page < 1 {
		page = 1
	}

	// 计算最大页码
	totalPages := (total + perPage - 1) / perPage
	if page > totalPages {
		// 指向最后一页的起始位置
		page = totalPages
	}

	return (page - 1) * perPage, perPage
}
