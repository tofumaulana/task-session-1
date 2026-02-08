package models

type TopProduct struct {
	Nama       string `json:"nama"`
	QtyTerjual int    `json:"qty_terjual"`
}

type DailyReport struct {
	TotalRevenue   int        `json:"total_revenue"`
	TotalTransaksi int        `json:"total_transaksi"`
	ProdukTerlaris TopProduct `json:"produk_terlaris"`
}