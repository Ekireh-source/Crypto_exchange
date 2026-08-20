-- 006_transactions_unique_idx.sql
-- Replace the overly restrictive unique index on tx_hash with one that allows
-- multiple entries with different types/users for the same hash (e.g. withdrawal + deposit).

DROP INDEX IF EXISTS idx_transactions_tx_hash_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_tx_hash_type_user_unique 
ON transactions(tx_hash, type, user_id) 
WHERE tx_hash IS NOT NULL;
