# Crypto Exchange — Feature-by-Feature Implementation Plan

## Stack Summary

| Layer | Technology |
|---|---|
| **Backend API** | Go + Echo v4 |
| **Blockchain (EVM/BSC)** | go-ethereum (already in go.mod ✅) |
| **Blockchain (TRON)** | TronGrid REST API (raw HTTP + gotron-sdk) |
| **Database** | PostgreSQL (pgx driver) + Redis (caching, sessions) |
| **Web Frontend** | Next.js 14 (App Router) + TypeScript |
| **Mobile App** | React Native (Expo) + TypeScript |
| **Auth** | JWT (access + refresh tokens) |
| **Real-time** | WebSocket (Echo) for balance updates, P2P chat |
| **Price Feed** | CoinGecko API (free tier) |

---

## Monorepo Folder Layout

```
crypto_exchange/                     ← REPO ROOT
│
├── backend/                         ← Go API server
│   ├── main.go
│   ├── go.mod / go.sum
│   ├── Makefile
│   ├── .env.example
│   │
│   ├── config/
│   │   └── config.go                # Viper/env loader
│   │
│   ├── orderbook/                   # ✅ Already implemented
│   │   ├── orderbook.go
│   │   └── orderbook_test.go
│   │
│   └── internal/
│       ├── models/                  # Domain structs (no DB tags here)
│       │   ├── user.go
│       │   ├── asset.go
│       │   ├── wallet.go
│       │   ├── transaction.go
│       │   └── p2p_order.go
│       │
│       ├── db/
│       │   ├── db.go                # pgx pool init
│       │   ├── migrations/
│       │   │   ├── 001_users.sql
│       │   │   ├── 002_assets.sql
│       │   │   ├── 003_wallets.sql
│       │   │   ├── 004_transactions.sql
│       │   │   └── 005_p2p_orders.sql
│       │   └── repository/
│       │       ├── user_repo.go
│       │       ├── wallet_repo.go
│       │       ├── transaction_repo.go
│       │       └── p2p_repo.go
│       │
│       ├── blockchain/
│       │   ├── interface.go         # BlockchainAdapter interface
│       │   ├── bsc/
│       │   │   ├── client.go
│       │   │   ├── wallet.go
│       │   │   ├── transfer.go
│       │   │   └── monitor.go
│       │   └── tron/
│       │       ├── client.go
│       │       ├── wallet.go
│       │       ├── transfer.go
│       │       └── monitor.go
│       │
│       ├── services/
│       │   ├── auth_service.go
│       │   ├── wallet_service.go
│       │   ├── transfer_service.go
│       │   ├── swap_service.go
│       │   ├── p2p_service.go
│       │   ├── price_service.go
│       │   └── referral_service.go
│       │
│       └── api/
│           ├── middleware/
│           │   ├── auth.go
│           │   └── ratelimit.go
│           ├── handlers/
│           │   ├── auth_handler.go
│           │   ├── wallet_handler.go
│           │   ├── transfer_handler.go
│           │   ├── swap_handler.go
│           │   ├── p2p_handler.go
│           │   ├── orderbook_handler.go
│           │   └── profile_handler.go
│           └── router.go
│
├── web/                             ← Next.js 14 App Router
│   ├── package.json
│   ├── next.config.ts
│   ├── tsconfig.json
│   └── src/
│       ├── app/
│       │   ├── layout.tsx           # Root layout (fonts, providers)
│       │   ├── page.tsx             # Landing / marketing page
│       │   ├── (auth)/
│       │   │   ├── login/page.tsx
│       │   │   └── register/page.tsx
│       │   └── (dashboard)/
│       │       ├── layout.tsx       # Sidebar + nav shell
│       │       ├── home/page.tsx    # Wallet dashboard (My Assets)
│       │       ├── send/page.tsx
│       │       ├── receive/page.tsx
│       │       ├── swap/page.tsx
│       │       ├── p2p/page.tsx
│       │       └── profile/page.tsx
│       ├── components/
│       │   ├── ui/                  # Shared UI primitives
│       │   │   ├── Button.tsx
│       │   │   ├── Modal.tsx
│       │   │   ├── CoinSelector.tsx
│       │   │   ├── AmountInput.tsx
│       │   │   └── QRCode.tsx
│       │   ├── wallet/
│       │   │   ├── AssetCard.tsx
│       │   │   ├── BalanceHeader.tsx
│       │   │   └── TransactionList.tsx
│       │   ├── swap/
│       │   │   ├── SwapForm.tsx
│       │   │   └── SwapQuote.tsx
│       │   └── p2p/
│       │       ├── OrderCard.tsx
│       │       └── TradeModal.tsx
│       ├── lib/
│       │   ├── api.ts               # Axios/fetch wrapper + interceptors
│       │   ├── auth.ts              # JWT storage, refresh logic
│       │   └── format.ts            # Number, currency formatters
│       ├── hooks/
│       │   ├── useWallet.ts
│       │   ├── useSwapQuote.ts
│       │   └── useP2POrders.ts
│       └── store/
│           └── authStore.ts         # Zustand auth state
│
└── mobile/                          ← React Native (Expo)
    ├── package.json
    ├── app.json
    ├── tsconfig.json
    └── src/
        ├── app/                     # Expo Router file-based routing
        │   ├── _layout.tsx          # Root nav layout
        │   ├── (auth)/
        │   │   ├── login.tsx
        │   │   └── register.tsx
        │   └── (tabs)/
        │       ├── _layout.tsx      # Bottom tab bar
        │       ├── index.tsx        # Home / wallet
        │       ├── swap.tsx
        │       ├── p2p.tsx
        │       └── profile.tsx
        ├── components/
        │   ├── ui/
        │   │   ├── Button.tsx
        │   │   ├── BottomSheet.tsx
        │   │   ├── CoinPicker.tsx
        │   │   └── QRScanner.tsx
        │   ├── wallet/
        │   │   ├── AssetRow.tsx
        │   │   └── BalanceBanner.tsx
        │   └── p2p/
        │       └── TraderCard.tsx
        ├── lib/
        │   ├── api.ts
        │   └── secureStorage.ts     # expo-secure-store for JWT
        ├── hooks/
        │   ├── useWallet.ts
        │   └── useSwapQuote.ts
        └── store/
            └── authStore.ts
```

---

## Feature 1 — Foundation & Infrastructure

> **Goal**: All the plumbing that every other feature depends on.

### Backend

#### [MODIFY] `backend/main.go`
- Remove hardcoded private key and ETH transaction from main()
- Initialize: config → DB → Redis → blockchain adapters → services → router
- Start deposit monitor goroutines

#### [NEW] `backend/config/config.go`
```go
type Config struct {
    Port              string
    DatabaseURL       string
    RedisURL          string
    JWTSecret         string
    MasterSeedPhrase  string  // BIP39 mnemonic for HD wallets
    BSCMainnetRPC     string
    BSCTestnetRPC     string
    TronGridAPIKey    string
    TronGridBaseURL   string
    CoinGeckoAPIKey   string
    ExchangeFeeRate   float64  // e.g. 0.001 = 0.1%
}
```

#### [NEW] `backend/internal/db/db.go`
- `pgxpool.Connect()` with connection pooling
- Expose `Pool *pgxpool.Pool` singleton

#### [NEW] `backend/internal/db/migrations/`
All 5 SQL migration files (run in order):

```sql
-- 001_users.sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  phone TEXT,
  password_hash TEXT NOT NULL,
  referral_code TEXT UNIQUE NOT NULL,
  referred_by UUID REFERENCES users(id),
  kyc_status TEXT DEFAULT 'pending',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 002_assets.sql
CREATE TABLE assets (
  id SERIAL PRIMARY KEY,
  symbol TEXT NOT NULL,         -- 'USDT', 'BNB', 'TRX'
  name TEXT NOT NULL,           -- 'Tether USD'
  network TEXT NOT NULL,        -- 'BSC', 'TRON'
  standard TEXT,                -- 'BEP20', 'TRC20', 'NATIVE'
  contract_address TEXT,        -- null for native coins
  decimals INT DEFAULT 18,
  is_active BOOLEAN DEFAULT true,
  UNIQUE(symbol, network)
);

-- 003_wallets.sql
CREATE TABLE deposit_addresses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  asset_id INT REFERENCES assets(id),
  address TEXT NOT NULL,
  encrypted_private_key TEXT NOT NULL,  -- AES-256 encrypted
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, asset_id)
);

CREATE TABLE balances (
  user_id UUID REFERENCES users(id),
  asset_id INT REFERENCES assets(id),
  available NUMERIC(36,18) DEFAULT 0,
  locked NUMERIC(36,18) DEFAULT 0,  -- locked in P2P escrow or open orders
  PRIMARY KEY (user_id, asset_id)
);

-- 004_transactions.sql
CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  asset_id INT REFERENCES assets(id),
  type TEXT NOT NULL,           -- 'deposit', 'withdrawal', 'swap', 'p2p_buy', 'p2p_sell'
  status TEXT DEFAULT 'pending',-- 'pending', 'confirmed', 'failed'
  amount NUMERIC(36,18) NOT NULL,
  fee NUMERIC(36,18) DEFAULT 0,
  tx_hash TEXT,                 -- on-chain tx hash (null for internal swaps)
  from_address TEXT,
  to_address TEXT,
  note TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  confirmed_at TIMESTAMPTZ
);

-- 005_p2p_orders.sql
CREATE TABLE p2p_orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seller_id UUID REFERENCES users(id),
  asset_id INT REFERENCES assets(id),
  fiat_currency TEXT DEFAULT 'USD',
  rate NUMERIC(18,6) NOT NULL,   -- price per coin in fiat
  min_amount NUMERIC(36,18) NOT NULL,
  max_amount NUMERIC(36,18) NOT NULL,
  available_amount NUMERIC(36,18) NOT NULL,
  payment_method TEXT,
  status TEXT DEFAULT 'active',  -- 'active', 'paused', 'completed'
  completion_rate NUMERIC(5,2) DEFAULT 0,
  total_orders INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE p2p_trades (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID REFERENCES p2p_orders(id),
  buyer_id UUID REFERENCES users(id),
  seller_id UUID REFERENCES users(id),
  asset_id INT REFERENCES assets(id),
  amount NUMERIC(36,18) NOT NULL,
  fiat_amount NUMERIC(18,2) NOT NULL,
  rate NUMERIC(18,6) NOT NULL,
  status TEXT DEFAULT 'waiting_payment', -- 'waiting_payment','paid','released','disputed','cancelled'
  escrow_locked BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  paid_at TIMESTAMPTZ,
  released_at TIMESTAMPTZ
);
```

#### [NEW] `backend/internal/blockchain/interface.go`
```go
type BlockchainAdapter interface {
    Network() string
    GenerateWallet() (address, encryptedPrivKey string, err error)
    GetNativeBalance(address string) (*big.Float, error)
    GetTokenBalance(address, contractAddress string, decimals int) (*big.Float, error)
    SendNative(fromPrivKey, toAddress string, amount *big.Float) (txHash string, err error)
    SendToken(fromPrivKey, toAddress, contractAddress string, amount *big.Float, decimals int) (txHash string, err error)
    GetIncomingTransactions(address string, fromBlock int64) ([]IncomingTx, error)
    EstimateFee(asset Asset) (*big.Float, error)
}
```

### Web Frontend (Next.js)

#### [NEW] `web/` — scaffold
```bash
npx create-next-app@latest web --typescript --app --tailwind=false --src-dir
```

#### [NEW] `web/src/lib/api.ts`
- Axios instance with `baseURL = process.env.NEXT_PUBLIC_API_URL`
- Request interceptor: attach `Authorization: Bearer <token>`
- Response interceptor: auto-refresh on 401

### Mobile (React Native)

#### [NEW] `mobile/` — scaffold
```bash
npx create-expo-app@latest mobile --template tabs
```

#### [NEW] `mobile/src/lib/api.ts`
- Same Axios pattern as web
- Use `expo-secure-store` for JWT (not localStorage)

---

## Feature 2 — Authentication

> **Goal**: Register, login, JWT tokens, secure session management.

### Backend

#### [NEW] `backend/internal/models/user.go`
```go
type User struct {
    ID           uuid.UUID
    Email        string
    Phone        string
    PasswordHash string
    ReferralCode string
    ReferredBy   *uuid.UUID
    KYCStatus    string
    CreatedAt    time.Time
}
```

#### [NEW] `backend/internal/services/auth_service.go`
- `Register(email, password, referralCode)` → hash password (bcrypt), generate referral code, insert user
- `Login(email, password)` → verify hash, issue JWT access (15min) + refresh (7d) tokens
- `RefreshToken(refreshToken)` → validate, issue new access token
- `ChangePassword(userID, old, new)`

#### [NEW] `backend/internal/api/handlers/auth_handler.go`
```
POST /auth/register   { email, password, referral_code? }
POST /auth/login      { email, password }
POST /auth/refresh    { refresh_token }
POST /auth/logout
```

#### [NEW] `backend/internal/api/middleware/auth.go`
- `RequireAuth` Echo middleware: parse Bearer JWT, inject `userID` into context

### Web Frontend

#### [NEW] `web/src/app/(auth)/login/page.tsx`
- Email + password form
- On success: store tokens in `httpOnly` cookie (via `/api/auth` Next.js route handler)
- Redirect to `/home`

#### [NEW] `web/src/app/(auth)/register/page.tsx`
- Email, password, confirm password, optional referral code
- Shows referral code benefit callout

#### [NEW] `web/src/store/authStore.ts`
- Zustand store: `{ user, isAuthenticated, login(), logout(), refreshToken() }`

### Mobile

#### [NEW] `mobile/src/app/(auth)/login.tsx`
- Same fields as web
- Store JWT in `expo-secure-store`

#### [NEW] `mobile/src/app/(auth)/register.tsx`

---

## Feature 3 — Wallet & Receive (My Assets + Deposit Addresses)

> **Goal**: Show all asset balances, generate per-network deposit address with QR code, monitor deposits.

### Backend

#### [NEW] `backend/internal/blockchain/bsc/wallet.go`
```go
// HD wallet: derive child key from master seed + user index
func DeriveWallet(masterSeed []byte, userIndex uint32) (address, privKey string)

// Or: generate standalone keypair
func GenerateWallet() (address, privKey string)
```

#### [NEW] `backend/internal/blockchain/tron/wallet.go`
```go
// Same secp256k1 key, different address encoding
func GenerateWallet() (tronAddress, privKey string)
// tronAddress = Base58Check(0x41 + keccak256(pubKey)[12:])
```

#### [NEW] `backend/internal/blockchain/bsc/monitor.go`
```go
// Background goroutine: poll eth_getLogs every 12s
// Filter: Transfer events to any of our deposit addresses
// On match: call ProcessDeposit(txHash, address, amount)
func StartDepositMonitor(client *ethclient.Client, repo WalletRepo)
```

#### [NEW] `backend/internal/blockchain/tron/monitor.go`
```go
// Background goroutine: poll TronGrid /v1/accounts/{addr}/transactions/trc20
// For each deposit address, check every 5s
func StartDepositMonitor(apiKey string, repo WalletRepo)
```

#### [NEW] `backend/internal/services/wallet_service.go`
```go
func GetPortfolio(userID uuid.UUID) ([]AssetBalance, totalUSD float64, error)
// → fetches balances from `balances` table + live prices from PriceSvc

func GetDepositAddress(userID uuid.UUID, assetID int) (DepositAddress, error)
// → looks up existing address in deposit_addresses
// → if not found: calls blockchain adapter GenerateWallet(), saves encrypted key, returns address

func ProcessDeposit(address, txHash string, amount *big.Float, assetID int) error
// → credits balances table, inserts confirmed transaction record
```

#### [NEW] `backend/internal/api/handlers/wallet_handler.go`
```
GET  /wallet                          → portfolio with USD values
GET  /wallet/deposit-address/:assetID → { address, network, qr_data }
GET  /wallet/transactions             → paginated tx history
     ?page=1&limit=20&type=deposit
```

#### [NEW] `backend/internal/services/price_service.go`
```go
// Fetch from CoinGecko, cache in Redis for 60s
func GetPrices(symbols []string) (map[string]float64, error)
```

### Web Frontend

#### [NEW] `web/src/app/(dashboard)/home/page.tsx`
Mirror of screenshot 1 — "My Assets" list:
- `BalanceHeader` at top: "Total Balance $X.XX"
- 4 quick-action buttons: Send / Receive / Swap / P2P
- `AssetCard` for each asset (icon, name, network label, balance, USD value)
- Refresh button

#### [NEW] `web/src/components/wallet/AssetCard.tsx`
```tsx
// Props: symbol, name, network, balance, usdValue, logoUrl
// Shows: coin icon, name (bold), network subtitle, amount right-aligned
```

#### [NEW] `web/src/app/(dashboard)/receive/page.tsx`
- Coin selector dropdown (like screenshot 3)
- Shows generated deposit address as text
- `QRCode` component (use `qrcode.react`)
- Copy address button
- Network warning banner: "Only send [ASSET] on [NETWORK] to this address"

### Mobile

#### [NEW] `mobile/src/app/(tabs)/index.tsx`
- `BalanceBanner`: large total balance display
- Horizontal quick-action row
- `FlatList` of `AssetRow` components

#### [NEW] `mobile/src/components/wallet/AssetRow.tsx`
- Same data as web `AssetCard`, adapted for native list

#### [NEW] `mobile/src/components/ui/QRScanner.tsx`
- Uses `expo-camera` to scan QR codes on Receive/Send screens

---

## Feature 4 — Send / Transfer

> **Goal**: User enters recipient address + amount, selects coin+network, confirms and broadcasts on-chain.

### Backend

#### [NEW] `backend/internal/blockchain/bsc/transfer.go`
```go
func SendBNB(fromPrivKey, toAddress string, amountWei *big.Int, client *ethclient.Client) (string, error)

func SendBEP20(fromPrivKey, toAddress, contractAddress string, amount *big.Int, client *ethclient.Client) (string, error)
// ABI encodes: transfer(address recipient, uint256 amount)
// Sends to contract address with encoded data
```

#### [NEW] `backend/internal/blockchain/tron/transfer.go`
```go
func SendTRX(fromPrivKey, toAddress string, amountSun int64) (string, error)
// POST /wallet/createtransaction → sign → POST /wallet/broadcasttransaction

func SendTRC20(fromPrivKey, toAddress, contractAddress string, amount int64) (string, error)
// POST /wallet/triggersmartcontract
// function_selector: "transfer(address,uint256)"
// then sign + broadcast
```

#### [NEW] `backend/internal/services/transfer_service.go`
```go
func EstimateFee(assetID int, toAddress string) (FeeEstimate, error)
// → calls adapter.EstimateFee(), returns in native gas coin + USD

func SendCrypto(userID uuid.UUID, req SendRequest) (Transaction, error)
// 1. Validate: balance >= amount + fee
// 2. Lock amount in balances.locked
// 3. Fetch user's deposit address private key (decrypt AES-256)
// 4. Call adapter.SendNative or SendToken
// 5. Insert transaction with status=pending, tx_hash
// 6. Background: poll for confirmation, then deduct balance
```

#### [NEW] `backend/internal/api/handlers/transfer_handler.go`
```
GET  /transfer/fee-estimate   ?assetID=&amount=&toAddress=
POST /transfer/send           { asset_id, to_address, amount, note? }
GET  /transfer/:txID          → single transaction detail
```

### Web Frontend

#### [NEW] `web/src/app/(dashboard)/send/page.tsx`
Flow (3 steps):
1. **Select Coin** — same `CoinSelector` modal as screenshot 3
2. **Enter Details** — recipient address (text + QR scan), amount input, shows fee estimate + total
3. **Confirm** — summary card, "Confirm & Send" button, success/error state

#### [NEW] `web/src/components/ui/AmountInput.tsx`
- Number input with MAX button
- Shows available balance below
- Live USD equivalent as user types

#### [NEW] `web/src/hooks/useTransfer.ts`
```ts
// Wraps fee estimation + send mutation
// Returns: { estimateFee, sendCrypto, isLoading, txHash }
```

### Mobile

#### [NEW] `mobile/src/app/send.tsx`
- Step-based bottom sheet flow (react-native-bottom-sheet)
- `QRScanner` for recipient address input
- Biometric confirmation before sending (expo-local-authentication)

---

## Feature 5 — Swap

> **Goal**: Internal CEX swap between any two supported assets. No on-chain tx needed — pure ledger update.

### Backend

#### [NEW] `backend/internal/services/swap_service.go`
```go
type SwapQuote struct {
    FromAsset   Asset
    ToAsset     Asset
    FromAmount  *big.Float
    ToAmount    *big.Float   // after fee
    Rate        float64      // e.g. 1 BNB = 310.4 USDT
    Fee         *big.Float
    FeeRate     float64      // e.g. 0.001
    ExpiresAt   time.Time    // quote valid for 10s
}

func GetSwapQuote(fromAssetID, toAssetID int, amount *big.Float) (SwapQuote, error)
// 1. Fetch live prices for both assets from PriceSvc
// 2. Calculate: toAmount = fromAmount * (fromPrice/toPrice) * (1 - feeRate)
// 3. Cache quote with UUID for 10s

func ExecuteSwap(userID uuid.UUID, quoteID string) (Transaction, error)
// 1. Re-validate quote (not expired)
// 2. Check fromAsset balance >= fromAmount
// 3. BEGIN tx: debit from, credit to, insert swap transaction record, COMMIT
// (all in-database, no blockchain call)
```

#### [NEW] `backend/internal/api/handlers/swap_handler.go`
```
GET  /swap/quote    ?from=assetID&to=assetID&amount=
POST /swap/execute  { quote_id }
GET  /swap/pairs    → all valid swap pairs with liquidity info
```

### Web Frontend

#### [NEW] `web/src/app/(dashboard)/swap/page.tsx`
Mirror of screenshot 2:
- "From" dropdown (shows coin + network + balance)
- Arrow/flip button to reverse swap direction
- "To" dropdown
- Amount input
- Live quote display: rate, fee, you receive
- 10-second countdown timer (re-fetches quote automatically)
- "Swap" CTA button → confirmation modal

#### [NEW] `web/src/components/swap/SwapForm.tsx`
```tsx
// State machine: idle → quoting → quoted → confirming → success/error
// Auto-fetches new quote every 10s when amount > 0
```

#### [NEW] `web/src/components/swap/SwapQuote.tsx`
```tsx
// Shows: 1 BNB = 310.4 USDT | Fee: 0.1% | You receive: 309.79 USDT
// 10s countdown with thin progress bar
```

#### [NEW] `web/src/hooks/useSwapQuote.ts`
```ts
// Debounces amount input (500ms), fetches quote, auto-refreshes every 10s
```

### Mobile

#### [NEW] `mobile/src/app/(tabs)/swap.tsx`
- Same UI adapted to mobile (vertical card layout)
- `CoinPicker` bottom sheet for From/To selection (screenshot 3 equivalent)
- Haptic feedback on swap confirmation

---

## Feature 6 — P2P Marketplace

> **Goal**: Users list crypto for sale at a custom rate. Buyers initiate trades. Seller confirms release. Escrow protection.

### Backend

#### [NEW] `backend/internal/services/p2p_service.go`
```go
func ListOrders(assetID int, side string, page int) ([]P2POrder, error)
// Returns active sell orders sorted by rate, completion rate

func CreateSellOrder(sellerID uuid.UUID, req CreateOrderReq) (P2POrder, error)
// 1. Validate seller has enough balance
// 2. Lock amount in balances.locked
// 3. Insert p2p_orders record

func InitiateBuy(buyerID uuid.UUID, orderID uuid.UUID, amount *big.Float) (P2PTrade, error)
// 1. Create p2p_trades record, status=waiting_payment
// 2. Lock crypto in escrow (already locked from CreateSellOrder)
// 3. Return trade with payment instructions

func ConfirmPaymentSent(buyerID uuid.UUID, tradeID uuid.UUID) error
// Buyer marks fiat payment as sent → status=paid

func ReleaseCrypto(sellerID uuid.UUID, tradeID uuid.UUID) error
// Seller confirms fiat received:
// 1. Transfer from locked balance to buyer's available balance
// 2. Update trade status=released
// 3. Insert transaction records for both parties
// 4. Update seller completion_rate

func DisputeTrade(userID uuid.UUID, tradeID uuid.UUID, reason string) error
// Creates dispute record, notifies admin

func CancelTrade(userID uuid.UUID, tradeID uuid.UUID) error
// Only before status=paid. Unlocks seller escrow.
```

#### [NEW] P2P WebSocket for real-time trade chat
```
GET /ws/p2p/:tradeID   → WebSocket connection
// Messages: { type: "message"|"status_update", content, sender_id, timestamp }
```

#### [NEW] `backend/internal/api/handlers/p2p_handler.go`
```
GET    /p2p/orders                  → list active sell orders ?assetID=&page=
POST   /p2p/orders                  → create sell order
DELETE /p2p/orders/:id              → cancel/remove listing

POST   /p2p/trades                  → initiate buy { order_id, amount }
GET    /p2p/trades/:id              → trade detail + chat history
POST   /p2p/trades/:id/pay          → buyer: mark fiat sent
POST   /p2p/trades/:id/release      → seller: release crypto
POST   /p2p/trades/:id/dispute      → raise dispute
POST   /p2p/trades/:id/cancel       → cancel trade

GET    /p2p/my-orders               → seller's own listings
GET    /p2p/my-trades               → buyer/seller trade history
```

### Web Frontend

#### [NEW] `web/src/app/(dashboard)/p2p/page.tsx`
Mirror of screenshot 4 (P2P marketplace):
- Buy / Sell tab toggle
- Asset + amount filter row
- `OrderCard` list: avatar, name, completion %, order count, rate, Buy button

#### [NEW] `web/src/components/p2p/OrderCard.tsx`
```tsx
// Props: seller, rate, limit (min-max), completionRate, totalOrders
// Buy button → opens TradeModal
```

#### [NEW] `web/src/components/p2p/TradeModal.tsx`
Trade flow with steps:
1. Enter amount to buy → see fiat total
2. Payment instructions screen (seller's payment method)
3. "I have paid" button → WebSocket chat opens
4. Awaiting seller confirmation
5. Success: crypto credited to wallet

#### [NEW] `web/src/app/(dashboard)/p2p/create/page.tsx`
Seller create listing form:
- Select asset, set rate, set min/max limits, select payment method

### Mobile

#### [NEW] `mobile/src/app/(tabs)/p2p.tsx`
- `FlatList` of `TraderCard` components
- Bottom sheet for trade flow
- Push notification on trade status change (expo-notifications)

---

## Feature 7 — Profile, Referral & Transaction History

> **Goal**: User profile page with referral link, team stats, trading history tabs. Mirror of screenshot 5.

### Backend

#### [NEW] `backend/internal/services/referral_service.go`
```go
func GetReferralStats(userID uuid.UUID) (ReferralStats, error)
// Returns: referral link, code, team count, fee earnings, volume

func ClaimReferralEarnings(userID uuid.UUID) error
// Moves earnings to available balance

// Triggered on: swap fee, P2P trade fee
func CreditReferralFee(userID uuid.UUID, feeAmount *big.Float, assetID int) error
// Finds referrer, credits % of fee to referrer's referral_earnings
```

#### [NEW] `backend/internal/api/handlers/profile_handler.go`
```
GET  /profile                  → user info + KYC status
PUT  /profile                  → update display name, avatar
GET  /profile/referral         → referral link, code, stats (screenshot 5)
GET  /profile/referral/team    → list of referred users
GET  /profile/transactions     → paginated, all types
     ?type=deposit|withdrawal|swap|p2p
     &from=2024-01-01&to=2024-12-31
GET  /profile/trading-stats    → volume, # trades, P&L summary
```

### Web Frontend

#### [NEW] `web/src/app/(dashboard)/profile/page.tsx`
Mirror of screenshot 5:
- User avatar / initials circle + referral ID
- 3-tab navigation: Transactions | Referrals | Trading
- **Referrals tab**: referral link card, copy + share buttons, stats grid (Team, Fee Earnings, Volume)
- "Your Team" list of referred users

#### [NEW] `web/src/components/wallet/TransactionList.tsx`
```tsx
// Groups by date, shows type icon, amount (+/-), status badge, asset
// Tap to expand: tx hash, from/to addresses, fee, timestamp
```

### Mobile

#### [NEW] `mobile/src/app/(tabs)/profile.tsx`
- Same tabs as web
- Share referral via `expo-sharing`
- Copy referral link via `expo-clipboard`

---

## Feature 8 — Admin Panel (Web Only)

> **Goal**: Internal dashboard to manage users, watch deposits, handle P2P disputes, view system health.

### Backend

#### [NEW] Admin-specific routes (separate middleware: `RequireAdmin`)
```
GET  /admin/users              → list users with KYC status
PUT  /admin/users/:id/kyc     → approve/reject KYC
GET  /admin/transactions       → all transactions, filter by status
GET  /admin/p2p/disputes       → open disputes
POST /admin/p2p/disputes/:id/resolve
GET  /admin/balances/hot-wallet → exchange hot wallet balances
POST /admin/assets/toggle      → enable/disable an asset
```

### Web Frontend

#### [NEW] `web/src/app/admin/` (separate layout, role-gated)
- Users table with KYC approve/reject
- Live transaction feed
- Dispute management queue
- Hot wallet balance monitor

---

## API Contract Summary

```
BASE URL: https://api.yourexchange.com/v1

── Auth ──────────────────────────────────────────────────────
POST /auth/register      POST /auth/login       POST /auth/refresh

── Wallet ────────────────────────────────────────────────────
GET  /wallet             GET  /wallet/deposit-address/:assetID
GET  /wallet/transactions

── Transfer ──────────────────────────────────────────────────
GET  /transfer/fee-estimate    POST /transfer/send

── Swap ──────────────────────────────────────────────────────
GET  /swap/quote    GET  /swap/pairs    POST /swap/execute

── P2P ───────────────────────────────────────────────────────
GET  /p2p/orders       POST /p2p/orders        DELETE /p2p/orders/:id
POST /p2p/trades       GET  /p2p/trades/:id
POST /p2p/trades/:id/pay   POST /p2p/trades/:id/release
POST /p2p/trades/:id/dispute
GET  /p2p/my-orders    GET  /p2p/my-trades

── Profile ───────────────────────────────────────────────────
GET  /profile    PUT  /profile    GET  /profile/referral
GET  /profile/referral/team    GET  /profile/transactions
GET  /profile/trading-stats

── WebSocket ─────────────────────────────────────────────────
WS   /ws/p2p/:tradeID           (P2P trade chat)
WS   /ws/portfolio              (live balance updates)
```

---

## Execution Order (Build Sequence)

```
[1] Foundation          → config, DB, migrations, blockchain interface
[2] Auth                → register/login, JWT, middleware
[3] Wallet + Receive    → addresses, balances, deposit monitoring
[4] Send / Transfer     → fee estimate, broadcast, confirmation
[5] Swap                → quote engine, internal ledger swap
[6] P2P                 → listings, escrow, trade flow, WebSocket chat
[7] Profile + Referral  → stats, referral tracking, tx history
[8] Admin Panel         → dispute resolution, KYC, monitoring
```

Each feature is vertically sliced:
**Backend service → API handler → Web UI → Mobile UI**

---

## Key Technical Decisions

| Decision | Choice | Reason |
|---|---|---|
| Wallet custody | Per-user HD wallet (BIP44) from master seed | Unique deposit address per user, all keys derived deterministically |
| Private key storage | AES-256 encrypted in DB, master key in env/Vault | Industry standard for hot wallets |
| Swap execution | Internal ledger (no on-chain tx) | Speed, zero gas cost, standard CEX model |
| TRON SDK | `github.com/fbsobreira/gotron-sdk` | Best Go wrapper for TronGrid |
| Real-time | Echo WebSocket | Already using Echo, lightweight |
| State management | Zustand (web) + Zustand (mobile) | Minimal boilerplate, works in both |
| P2P escrow | DB-level lock on `balances.locked` | Simple, auditable, no smart contract needed |
| Fee model | 0.1% swap fee, 0% deposit, flat withdrawal fee | Standard CEX model |

---

## Environment Variables (`.env.example`)

```bash
# Server
PORT=3000
JWT_SECRET=your-256-bit-secret
EXCHANGE_FEE_RATE=0.001

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/exchange

# Redis
REDIS_URL=redis://localhost:6379

# HD Wallet Master Seed (NEVER commit the real value)
MASTER_SEED_PHRASE=word1 word2 ... word24

# BSC
BSC_MAINNET_RPC=https://bsc-dataseed.binance.org/
BSC_TESTNET_RPC=https://data-seed-prebsc-1-s1.binance.org:8545/
BSC_CHAIN_ID=56
USDT_BSC_CONTRACT=0x55d398326f99059fF775485246999027B3197955

# TRON
TRON_GRID_API_KEY=your-trongrid-api-key
TRON_GRID_BASE_URL=https://api.trongrid.io
USDT_TRC20_CONTRACT=TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t

# Prices
COINGECKO_API_KEY=your-key
```
