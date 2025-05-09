-- Создание таблицы для кошельков
CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0,
    currency_code TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    operation_type VARCHAR(10) NOT NULL CHECK (operation_type IN ('DEPOSIT', 'WITHDRAW')),
    amount BIGINT NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'COMPLETED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE currencies (
    code CHAR(3) PRIMARY KEY, -- например: 'USD', 'RUB', 'EUR'
    name TEXT NOT NULL,       -- Полное название, например "US Dollar"
    minor_units SMALLINT NOT NULL DEFAULT 2, -- Кол-во дробных знаков (например: 2 → 1 доллар = 100 центов)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Функция для обновления поля updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для обновления updated_at для таблицы wallets
CREATE TRIGGER update_wallets_updated_at
BEFORE UPDATE ON wallets
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Триггер для обновления updated_at для таблицы transactions
CREATE TRIGGER update_transactions_updated_at
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

INSERT INTO currencies (code, name, minor_units)
VALUES 
  ('USD', 'US Dollar', 2),
  ('EUR', 'Euro', 2),
  ('RUB', 'Russian Ruble', 2);


INSERT INTO wallets (id, balance, currency_code)
VALUES 
  ('33333333-3333-3333-3333-333333333333', 75000, 'RUB');

