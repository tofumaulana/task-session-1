package repositories

import (
	"database/sql"
	"fmt"
	"task-sesion-1/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) GetDailyReport() (*models.DailyReport, error) {
    report := &models.DailyReport{}

    // 1. Query Total Revenue & Total Transaksi Hari Ini
    // COALESCE digunakan agar jika tidak ada data, return 0 (bukan NULL error)
    queryTotals := `
        SELECT 
            COALESCE(SUM(total_amount), 0), 
            COUNT(id) 
        FROM transactions 
        WHERE DATE(created_at) = CURRENT_DATE
    `
    err := repo.db.QueryRow(queryTotals).Scan(&report.TotalRevenue, &report.TotalTransaksi)
    if err != nil {
        return nil, err
    }

    // 2. Query Produk Terlaris Hari Ini
    // Join transactions -> transaction_details -> products
    queryTopProduct := `
        SELECT 
            p.name, 
            SUM(td.quantity) as total_qty
        FROM transaction_details td
        JOIN transactions t ON td.transaction_id = t.id
        JOIN products p ON td.product_id = p.id
        WHERE DATE(t.created_at) = CURRENT_DATE
        GROUP BY p.name
        ORDER BY total_qty DESC
        LIMIT 1
    `
    
    err = repo.db.QueryRow(queryTopProduct).Scan(&report.ProdukTerlaris.Nama, &report.ProdukTerlaris.QtyTerjual)
    if err == sql.ErrNoRows {
        // Jika belum ada penjualan hari ini, set default string kosong atau "-"
        report.ProdukTerlaris.Nama = "-"
        report.ProdukTerlaris.QtyTerjual = 0
    } else if err != nil {
        return nil, err
    }

    return report, nil
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	var (
		res *models.Transaction
	)

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// inisialisasi subtotal -> jumlah total transaksi keseluruhan
	totalAmount := 0
	// inisialisasi modeling transactionDetails -> nanti kita insert ke db
	details := make([]models.TransactionDetail, 0)
	// loop setiap item
	for _, item := range items {
		var productName string
		var productID, price, stock int
		// get product dapet pricing
		err := tx.QueryRow("SELECT id, name, price, stock FROM products WHERE id=$1", item.ProductID).Scan(&productID, &productName, &price, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}

		if err != nil {
			return nil, err
		}

		// hitung current total = quantity * pricing
		// ditambahin ke dalam subtotal
		subtotal := item.Quantity * price
		totalAmount += subtotal

		// kurangi jumlah stok
		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, productID)
		if err != nil {
			return nil, err
		}

		// item nya dimasukkin ke transactionDetails
		details = append(details, models.TransactionDetail{
			ProductID:   productID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// insert transaction
	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING ID", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// insert transaction details

	qDetails := `INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4)`
    
    for i := range details {
        details[i].TransactionID = transactionID
        if _, err := tx.Exec(qDetails, transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal); err != nil {
            return nil, err
        }
    }

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	res = &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}

	return res, nil
}