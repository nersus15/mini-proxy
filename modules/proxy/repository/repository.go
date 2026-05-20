package repository

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/entity"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

type ProxyRepository struct {
	Connection port.IDatabase
	Context    *core.AppContext
	Config     *config.ModuleConfig
}

type TransactiocPayload struct {
}

func NewProxyRepository(ctx *core.AppContext, cfg *config.ModuleConfig, conn port.IDatabase) *ProxyRepository {
	return &ProxyRepository{
		Connection: conn,
		Context:    ctx,
		Config:     cfg,
	}
}

func (r *ProxyRepository) SaveTransactionError(transaction *entity.Transactions) error {
	logger.DebugJson("Simpan Error Transaction Ke DB", transaction)
	_, err := r.Connection.InsertOne(r.Context.Context, transaction.TableName(), transaction)

	return err
}

func (r *ProxyRepository) GetPendingTransactions(limit int64) ([]entity.Transactions, error) {
	results := make([]entity.Transactions, 0)

	// Filter menggunakan DbExpression sesuai interface
	filter := []port.DbExpression{
		{
			Expr: "(status = ? OR status = ?)",
			Args: []any{"PENDING", "RETRY"},
		},
	}

	// Menentukan urutan (Lama ke Baru)
	sort := map[string]int{"created_at": 1}
	logger.Info("TabelName", entity.Transactions{}.TableName())

	err := r.Connection.Find(
		r.Context.Context,
		&results,
		entity.Transactions{}.TableName(),
		[]string{"*"},
		filter,
		sort,
		limit,
		0,
	)

	return results, err
}

// UpdateTransaction memperbarui data transaksi (misal: update retry_count atau status)
func (r *ProxyRepository) UpdateTransaction(transaction *entity.Transactions) error {
	// Filter berdasarkan ID spesifik
	filter := []port.DbExpression{}

	// Menggunakan UpdateOne untuk efisiensi
	_, err := r.Connection.UpdateOne(
		r.Context.Context,
		transaction.TableName(),
		filter,
		transaction,
	)

	return err
}

func (r *ProxyRepository) UpdateBulkTransactions(transactions []entity.Transactions) error {
	conn := r.Connection.GetConnection()

	if bunDB, ok := conn.(*bun.DB); ok {

		_, err := bunDB.NewUpdate().
			Model(&transactions).
			Column("status", "updated_at", "retry_count", "error_message").
			Bulk().
			Exec(r.Context.Context)

		return err
	} else {
		return fmt.Errorf("Connection bukan bun.DB")
	}
}

// DeleteOldTransactions menghapus data yang sudah sukses (opsional untuk cleanup)
func (r *ProxyRepository) DeleteOldTransactions(status string) (int64, error) {
	filter := []port.DbExpression{}

	return r.Connection.Delete(r.Context.Context, entity.Transactions{}.TableName(), filter)
}

func (r *ProxyRepository) SaveClientCredentials(clientCredential *entity.ClinetCredential) error {
	conn := r.Connection.GetConnection()
	bunDB, ok := conn.(*bun.DB)
	if !ok {
		return fmt.Errorf("connection is not an instance of bun.DB")
	}

	logger.Debug(fmt.Sprintf("Memproses simpan/update client credentials %s ke database...", clientCredential.ClientID))

	// Gunakan context bawaan dari AppContext milik repository r.Context.Context
	ctx := r.Context.Context

	_, err := bunDB.NewInsert().
		Model(clientCredential).
		On("DUPLICATE KEY UPDATE").
		Set("access_token = VALUES(access_token)").
		Set("expired_at = VALUES(expired_at)").
		Set("env = VALUES(env)").
		Set("updated_at = NOW()").
		Exec(ctx)

	if err != nil {
		logger.Error(fmt.Sprintf("Gagal simpan/update client credentials %s ke database", clientCredential.ClientID), err)
		return err
	}

	logger.Debug(fmt.Sprintf("Berhasil sinkronisasi client credentials %s ke database.", clientCredential.ClientID))
	return nil
}
