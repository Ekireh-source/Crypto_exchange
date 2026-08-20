# Wallet Architecture & Sweeper Implementation Plan

This document outlines the phased implementation plan for upgrading the exchange's wallet system from a "per-user custodial" model to a production-ready **Centralized Hot Wallet & Sweeper** architecture.

This architecture ensures the security of funds via isolated hot wallets, solves the "Gas Trap" for token transfers, and prevents the exchange from operating at a loss by using sweeping thresholds and withdrawal fees.

---

## Phase 1: Environment & Core Configuration
Before writing business logic, we must configure the master wallets and economic variables.

- [ ] **1.1. Master Wallet Keys:** Add environment variables for `HOT_WALLET_BSC_KEY`, `HOT_WALLET_TRON_KEY`, and `MASTER_GAS_WALLET_KEY`.
- [ ] **1.2. Economic Thresholds:** Add environment variables to control profitability:
  - `MIN_SWEEP_USD_THRESHOLD` (e.g., $50) - Only sweep user addresses if their token value exceeds this amount to save on fixed gas fees.
  - `WITHDRAWAL_FEE_USD` (e.g., $1.00) - The flat fee charged to users to subsidize sweeping and withdrawal gas costs.
- [ ] **1.3. Block Confirmations:** Add `BSC_MIN_CONFIRMATIONS` (e.g., 15) and `TRON_MIN_CONFIRMATIONS` (e.g., 20) to prevent double-spend attacks from blockchain reorgs.

## Phase 2: Database Schema Updates
Track the state of deposits to ensure they are swept properly.

- [ ] **2.1. Update Transactions Table:** Add a `sweep_status` column (`enum: NOT_NEEDED, PENDING_GAS, SWEEPING, COMPLETED`) to track the lifecycle of a user's deposit.
- [ ] **2.2. Create Sweep Logs Table:** Create a table to log internal background transfers (e.g., funding a user address with BNB, transferring the swept USDT to the Hot Wallet) for auditing and accounting.

## Phase 3: The Sweeper Bot (Background Service)
The core engine that secures user funds into the central liquidity pool.

- [ ] **3.1. Sweeper Daemon:** Create a background worker (`sweeper_service.go`) that continuously polls the database for confirmed deposits where `sweep_status != COMPLETED`.
- [ ] **3.2. Minimum Value Checks:** Implement logic to check if a pending deposit meets the `MIN_SWEEP_USD_THRESHOLD`. If it's too small, leave it in the user's deposit address until they deposit more.
- [ ] **3.3. Native Coin Sweeping:** For BNB/TRX deposits, calculate the maximum amount that can be sent minus the network gas fee, and sweep it directly to the Hot Wallet.
- [ ] **3.4. Token Sweeping (Gas Funding):** 
  - For USDT deposits, check the BNB/TRX balance of the user's deposit address.
  - If insufficient, broadcast a transaction from the `MASTER_GAS_WALLET` to send exactly enough gas to the user's address. Update status to `PENDING_GAS`.
- [ ] **3.5. Token Sweeping (Execution):** 
  - Once the gas arrives, broadcast the USDT transfer from the user's deposit address to the Hot Wallet. Update status to `COMPLETED`.

## Phase 4: Deposit Processing Upgrades
Ensure incoming funds are validated securely before crediting the user.

- [ ] **4.1. Confirmations:** Update `ProcessDeposit` to wait for the required block confirmations before officially marking the transaction as `TxConfirmed` and crediting the user's DB balance.
- [ ] **4.2. Sweep Trigger:** Upon confirming the deposit, insert it with `sweep_status = PENDING_SWEEP` so the Sweeper Bot picks it up.

## Phase 5: Withdrawal Processing Upgrades
Shift outbound transactions from the user's deposit address to the Central Hot Wallet.

- [ ] **5.1. Fee Deduction:** Update the withdrawal request handler to deduct the flat `WITHDRAWAL_FEE_USD` from the user's balance.
- [ ] **5.2. Hot Wallet Broadcasting:** Rewrite `SendCrypto` to completely ignore the user's deposit address. Instead, load the appropriate `HOT_WALLET_KEY` for the requested network and execute the outward transfer directly from the company's central liquidity pool.
- [ ] **5.3. Retry Mechanism:** Add a monitor for pending withdrawals. If a Hot Wallet transaction gets stuck due to network congestion, allow the system to re-broadcast it with a higher gas fee.
