CREATE TABLE IF NOT EXISTS transaction_analytics (
    transaction_id UUID,
    user_id UUID,
    amount DECIMAL(19,4),
    currency String,
    status String,
    transaction_type String, -- 'deposit', 'transfer', 'withdrawal'
    created_at DateTime,
    year UInt16,
    month UInt8,
    day UInt8
) ENGINE = MergeTree()
PARTITION BY (year, month)
ORDER BY (user_id, created_at, transaction_type)
SETTINGS index_granularity = 8192;