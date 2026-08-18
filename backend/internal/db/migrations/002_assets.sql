-- 002_assets.sql
-- Assets: each row is a unique coin+network combination.
-- USDT on BSC and USDT on TRON are TWO separate assets because they live on
-- different chains, have different address formats, and require different
-- blockchain adapters to interact with.

CREATE TABLE IF NOT EXISTS assets (
    id               SERIAL      PRIMARY KEY,
    symbol           TEXT        NOT NULL,         -- ticker:  'USDT', 'BNB', 'TRX'
    name             TEXT        NOT NULL,         -- display: 'Tether USD (BEP20)'
    network          TEXT        NOT NULL,         -- chain:   'BSC', 'TRON'
    standard         TEXT        NOT NULL,         -- type:    'NATIVE', 'BEP20', 'TRC20'
    contract_address TEXT,                         -- NULL for native coins (BNB, TRX)
    decimals         INT         NOT NULL DEFAULT 18,
    logo_url         TEXT        NOT NULL DEFAULT '',
    is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
    UNIQUE (symbol, network)
);

-- Seed the four assets shown in the Rivo screenshot.
INSERT INTO assets (symbol, name, network, standard, contract_address, decimals, logo_url) VALUES
    ('BNB',  'BNB Smart Chain',      'BSC',  'NATIVE', NULL,                                                   18, '/icons/bnb.svg'),
    ('USDT', 'Tether USD (BEP20)',   'BSC',  'BEP20',  '0x55d398326f99059fF775485246999027B3197955',           18, '/icons/usdt.svg'),
    ('TRX',  'TRON',                 'TRON', 'NATIVE', NULL,                                                    6, '/icons/trx.svg'),
    ('USDT', 'Tether USD (TRC20)',   'TRON', 'TRC20',  'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',                  6, '/icons/usdt.svg')
ON CONFLICT (symbol, network) DO NOTHING;
