package kafka

const (
	EventWalletCreated     = "Wallet_Created"
	EventWalletDeleted     = "Wallet_Deleted"
	EventWalletDeposited   = "Wallet_Deposited"
	EventWalletWithdrawn   = "Wallet_Withdrawn"
	EventWalletTransferred = "Wallet_Transfered"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WalletCreatedPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WalletDeletedPayload struct {
	ID string `json:"id"`
}

type WalletDepositedPayload struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type WalletWithdrawnPayload struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type WalletTransferredPayload struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	TransferTo string  `json:"transfer_to"`
	Amount     float64 `json:"amount"`
}
