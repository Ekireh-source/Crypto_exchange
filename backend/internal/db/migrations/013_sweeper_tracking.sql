-- 003_sweeper_tracking.sql
-- Add sweep tracking to deposits and create logs for background operations.

-- We use TEXT to avoid complicating schema migrations if we need to add states later.
ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS sweep_status TEXT NOT NULL DEFAULT 'NOT_NEEDED';

-- Create sweep logs table to track internal bot movements
CREATE TABLE IF NOT EXISTS sweep_logs (
    id SERIAL PRIMARY KEY,
    transaction_id UUID REFERENCES transactions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id INT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    type TEXT NOT NULL,         -- e.g., 'GAS_FUNDING', 'TOKEN_SWEEP', 'NATIVE_SWEEP'
    status TEXT NOT NULL,       -- e.g., 'PENDING', 'SUCCESS', 'FAILED'
    amount TEXT NOT NULL,
    tx_hash TEXT,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for querying unresolved sweeps faster
CREATE INDEX IF NOT EXISTS idx_transactions_sweep_status 
ON transactions(sweep_status) 
WHERE sweep_status != 'COMPLETED' AND sweep_status != 'NOT_NEEDED';

CREATE INDEX IF NOT EXISTS idx_sweep_logs_tx_id 
ON sweep_logs(transaction_id);
