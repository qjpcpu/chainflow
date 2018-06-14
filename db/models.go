package db

func init() {
	RegisterModel(
		new(TokenTransfer),
		new(TokenBalance),
	)
}

type TokenTransfer struct {
	Id       uint64 `orm:"pk;auto" json:"-"`
	Contract string `json:"contract"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    uint64 `json:"value"`
	Tx       string `json:"tx"`
	Block    uint64 `json:"block"`
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
