package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/entity"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

type ProxyRepository struct {
	Connection port.IDatabase
	Context    *core.AppContext
	Config     *config.ModuleConfig
	Memory     port.ICacheMemory
}

type TransactiocPayload struct {
}

func NewProxyRepository(ctx *core.AppContext, cfg *config.ModuleConfig, conn port.IDatabase, mem port.ICacheMemory) *ProxyRepository {
	return &ProxyRepository{
		Connection: conn,
		Context:    ctx,
		Config:     cfg,
		Memory:     mem,
	}
}

func (r *ProxyRepository) SaveTransactionError(transaction *entity.Transactions) error {
	_, err := r.Connection.InsertOne(r.Context.Context, transaction.TableName(), transaction)

	return err
}

func (r *ProxyRepository) GetPendingTransactions(token *string, limit int64) ([]entity.Transactions, error) {
	results := make([]entity.Transactions, 0)

	// Filter menggunakan DbExpression sesuai interface
	filter := []port.DbExpression{
		{
			Expr: "(status = ? OR status = ?)",
			Args: []any{"PENDING", "RETRY"},
		},
	}

	if token != nil {
		filter = append(filter, port.DbExpression{
			Expr: "id = ?",
			Args: []any{token},
		})
	} else {
		filter = append(filter, port.DbExpression{
			Expr: "resource_type != ?",
			Args: []any{"credential"},
		})
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
	batchSize := 100

	if bunDB, ok := conn.(*bun.DB); ok {
		err := bunDB.RunInTx(r.Context.Context, nil, func(ctx context.Context, tx bun.Tx) error {
			for i := 0; i < len(transactions); i += batchSize {
				end := i + batchSize
				if end > len(transactions) {
					end = len(transactions)
				}

				batch := transactions[i:end]

				_, err := tx.NewUpdate().
					Model(&batch).
					Column("status", "updated_at", "retry_count", "error_message").
					Bulk().
					Exec(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		})
		return err
	} else {
		return fmt.Errorf("Connection bukan bun.DB")
	}
}

// DeleteOldTransactions menghapus data yang sudah sukses (opsional untuk cleanup)
func (r *ProxyRepository) DeleteOldTransactions(status string, credOnly bool) (int64, error) {

	filter := []port.DbExpression{
		{
			Expr: "status",
			Args: []any{
				status,
			},
		},
	}

	if credOnly {
		filter = append(filter, port.DbExpression{
			Expr: "resource_type",
			Args: []any{
				"credential",
			},
		})
	}
	return r.Connection.Delete(r.Context.Context, entity.Transactions{}.TableName(), filter)
}

func (r *ProxyRepository) DeleteOldCredTransactions(token string, clientId string) (int64, error) {
	filter := []port.DbExpression{
		{
			Expr: "id != ? AND resource_type = ? AND client = ?",
			Args: []any{
				token,
				"credential",
				clientId,
			},
		},
	}
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
		On("CONFLICT (client_id) DO UPDATE").
		Set("access_token = EXCLUDED.access_token").
		Set("expired_at = EXCLUDED.expired_at").
		Set("env = EXCLUDED.env").
		Set("updated_at = CURRENT_TIMESTAMP").
		Returning("NULL").
		Exec(ctx)

	if err != nil {
		logger.Error(fmt.Sprintf("Gagal simpan/update client credentials %s ke database", clientCredential.ClientID), err)
		return err
	}

	logger.Debug(fmt.Sprintf("Berhasil sinkronisasi client credentials %s ke database.", clientCredential.ClientID))
	return nil
}

func (r *ProxyRepository) GetToken(token *string) []types.Token {
	// tokens := map[string]string{}
	tokens := make([]types.Token, 0)
	var memkey string
	var credential []entity.ClinetCredential

	filter := []port.DbExpression{
		// {
		// 	Expr: "(client_id = ? OR client_id = ?)",
		// 	Args: []any{r.Config.SatSetDev.ClientID, r.Config.SatSetProd.ClientID},
		// },
	}

	if token != nil {
		logger.Info(fmt.Sprintf("Ambil token: %s", *token))

		memkey = "token_" + *token
		// Cek cache
		ok := r.Memory.Get(memkey, &tokens)
		if ok {
			return tokens
		}

		filter = append(filter, port.DbExpression{
			Expr: "access_token = ?",
			Args: []any{token},
		})
	}
	sort := map[string]int{"created_at": 1}
	err := r.Connection.Find(r.Context.Context, &credential, entity.ClinetCredential{}.TableName(), []string{"env", "access_token", "client_id"}, filter, sort, 0, 0)

	if err != nil {
		logger.DebugJson("Gagal Ambil Token", err.Error())
		return tokens
	}
	for _, cred := range credential {
		tokens = append(tokens, types.Token{
			AccessToken: cred.AccessToken,
			ClientId:    cred.ClientID,
			Env:         cred.Env,
		})
	}

	// Cache token yang didapat dari GetToken - jika tujuannya untuk mendapatkan client_id
	if token != nil {
		r.Memory.Set(memkey, tokens, 5*time.Hour)
	}

	return tokens
}
func (r *ProxyRepository) GetPendingCredentialTransactions(limit int64) ([]entity.Transactions, error) {
	conn := r.Connection.GetConnection()
	bunDB, ok := conn.(*bun.DB)
	if !ok {
		return nil, fmt.Errorf("connection bukan *bun.DB")
	}

	var results []entity.Transactions
	query := bunDB.NewSelect().
		Model(&results).
		ColumnExpr("tr.*").
		Join("JOIN client_credentials AS cc ON cc.access_token = tr.id").
		Where("tr.status IN (?)", bun.List([]string{"PENDING", "RETRY"})).
		Where("tr.resource_type = ?", "credential").
		OrderExpr("tr.created_at ASC")

	if limit > 0 {
		query = query.Limit(int(limit))
	}

	err := query.Scan(r.Context.Context)
	return results, err
}

func (d *ProxyRepository) StartMigration(DB interface{}, dialect string, service string, command string, dir string, args []string) error {
	bunDB, ok := DB.(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: objek yang dikirim bukan merupakan *bun.DB")
		return fmt.Errorf("invalid *bun.DB instance")
	}

	// Set dialek wajib di sini sebelum eksekusi run
	if dialect == "sqlite" {
		goose.SetDialect("sqlite3")
	}

	if service != "" {
		goose.SetTableName("__migration_" + service + "_logs")
	} else {
		goose.SetTableName("__migration_webcore_logs")
	}

	logger.Info(fmt.Sprintf("Mengeksekusi Goose %s pada folder: %s", command, dir))

	if err := goose.RunContext(d.Context.Context, command, bunDB.DB, dir, args...); err != nil {
		logger.Error(fmt.Sprintf("goose run %s: %v", command, err))
		return err
	}

	return nil
}

// FindIdempotency mengembalikan nil kalau tidak ada atau sudah kedaluwarsa.
func (r *ProxyRepository) FindIdempotency(fingerprint string) (*entity.RequestIdempotency, error) {
	var records []entity.RequestIdempotency

	filter := []port.DbExpression{
		{
			Expr: "fingerprint = ? AND expired_at > ?",
			Args: []any{fingerprint, time.Now()},
		},
	}

	err := r.Connection.Find(r.Context.Context, &records, entity.RequestIdempotency{}.TableName(),
		[]string{"fingerprint", "client", "env", "method", "resource_type", "url", "resource_id", "response_body", "created_at", "expired_at"},
		filter, map[string]int{}, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return &records[0], nil
}

// SaveIdempotency menyimpan respons sukses.
func (r *ProxyRepository) SaveIdempotency(record *entity.RequestIdempotency) error {
	_, err := r.Connection.InsertOne(r.Context.Context, record.TableName(), record)

	return err
}

// DeleteExpiredIdempotency membersihkan cache yang sudah lewat masa berlaku.
func (r *ProxyRepository) DeleteExpiredIdempotency() (int64, error) {
	filter := []port.DbExpression{
		{
			Expr: "expired_at <= ?",
			Args: []any{time.Now()},
		},
	}

	return r.Connection.Delete(r.Context.Context, entity.RequestIdempotency{}.TableName(), filter)
}
