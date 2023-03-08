package internal

type UserID int

type Token string

type OrderNumber string

type OrderStatus string

const (
	New        OrderStatus = "NEW"
	Processing OrderStatus = "PROCESSING"
	Invalid    OrderStatus = "INVALID"
	Processed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     OrderNumber `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    *float64    `json:"accrual,omitempty"`
	UploadedAt string      `json:"uploaded_at"`
}

type ProcessingOrder struct {
	Number  OrderNumber
	Status  OrderStatus
	Accrual *float64
	UserID  UserID
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawReq struct {
	Number OrderNumber `json:"order"`
	Sum    float64     `json:"sum"`
	UserID UserID
}

type Withdrawal struct {
	Number      OrderNumber `json:"order"`
	Sum         float64     `json:"sum"`
	ProcessedAt string      `json:"processed_at"`
}
