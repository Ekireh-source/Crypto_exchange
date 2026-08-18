package models

// Network identifies a blockchain network.
type Network string

const (
	NetworkBSC  Network = "BSC"
	NetworkTRON Network = "TRON"
)

// TokenStandard describes how a token is implemented on its chain.
type TokenStandard string

const (
	StandardNative TokenStandard = "NATIVE" // e.g. BNB on BSC, TRX on TRON
	StandardBEP20  TokenStandard = "BEP20"
	StandardTRC20  TokenStandard = "TRC20"
)

// Asset represents a unique coin+network combination.
// USDT on BSC and USDT on TRON are two distinct assets even though they share a
// ticker — they live on different blockchains and require different adapters.
type Asset struct {
	ID              int           `db:"id"               json:"id"`
	Symbol          string        `db:"symbol"           json:"symbol"`          // e.g. "USDT"
	Name            string        `db:"name"             json:"name"`            // e.g. "Tether USD"
	Network         Network       `db:"network"          json:"network"`         // e.g. "BSC"
	Standard        TokenStandard `db:"standard"         json:"standard"`        // e.g. "BEP20"
	ContractAddress *string       `db:"contract_address" json:"contract_address"` // nil for native coins
	Decimals        int           `db:"decimals"         json:"decimals"`        // e.g. 18 for BEP20, 6 for TRC20 USDT
	IsActive        bool          `db:"is_active"        json:"is_active"`
	LogoURL         string        `db:"logo_url"         json:"logo_url"`
}

// IsNative returns true when the asset is the chain's gas/native coin.
func (a Asset) IsNative() bool {
	return a.Standard == StandardNative
}

// SeedAssets returns the default set of supported assets.
// These are inserted during the initial DB seed/migration.
func SeedAssets() []Asset {
	usdtBSCContract := "0x55d398326f99059fF775485246999027B3197955"
	usdtTRXContract := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"

	return []Asset{
		{
			Symbol:   "BNB",
			Name:     "BNB Smart Chain",
			Network:  NetworkBSC,
			Standard: StandardNative,
			Decimals: 18,
			IsActive: true,
		},
		{
			Symbol:          "USDT",
			Name:            "Tether USD (BEP20)",
			Network:         NetworkBSC,
			Standard:        StandardBEP20,
			ContractAddress: &usdtBSCContract,
			Decimals:        18,
			IsActive:        true,
		},
		{
			Symbol:   "TRX",
			Name:     "TRON",
			Network:  NetworkTRON,
			Standard: StandardNative,
			Decimals: 6,
			IsActive: true,
		},
		{
			Symbol:          "USDT",
			Name:            "Tether USD (TRC20)",
			Network:         NetworkTRON,
			Standard:        StandardTRC20,
			ContractAddress: &usdtTRXContract,
			Decimals:        6,
			IsActive:        true,
		},
	}
}
