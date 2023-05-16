-- Replace gas_quota with gas_lock which checks network balance threshold 
DROP TRIGGER IF EXISTS update_gas_quota_timestamp ON gas_quota;
DROP TABLE IF EXISTS gas_quota_meta;
DROP TABLE IF EXISTS gas_quota;
DROP TRIGGER IF EXISTS insert_gas_quota ON keystore;
DROP FUNCTION IF EXISTS insert_gas_quota;

-- Gas lock table
-- A gas_locked account indicates gas balance is below threshold awaiting next available top up
CREATE TABLE IF NOT EXISTS gas_lock (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    key_id INT REFERENCES keystore(id) NOT NULL,
    lock BOOLEAN DEFAULT true,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create function insert_gas_lock()
    returns trigger
as $$
begin
    insert into gas_lock (key_id) values (new.id);
    return new;
end;
$$ language plpgsql;

create trigger insert_gas_lock
    after insert on keystore
for each row
execute procedure insert_gas_lock();

create trigger update_gas_lock_timestamp
    before update on gas_lock
for each row
execute procedure update_timestamp();