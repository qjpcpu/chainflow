package db

func init() {
	RegisterModel(
		new(TokenTransfer),
		new(TokenBalance),
	)
}

type TokenTransfer struct {
	Id       uint64 `orm:"pk;auto"`
	Contract string
	From     string
	To       string
	Value    uint64
	Tx       string
	Block    uint64
}

func (t *TokenTransfer) TableName() string {
	return "token_transfer"
}

type TokenBalance struct {
	Id       uint64 `orm:"pk;auto"`
	Contract string
	User     string
	Amount   int64
	Block    uint64
}

func (t *TokenBalance) TableName() string {
	return "token_balance"
}
