-- Gas quota meta table
CREATE TABLE IF NOT EXISTS gas_quota_meta (
    default_quota INT NOT NULL
);
INSERT INTO gas_quota_meta (default_quota) VALUES (25);

-- Gas quota table
CREATE TABLE IF NOT EXISTS gas_quota (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    key_id INT REFERENCES keystore(id) NOT NULL,
    quota INT NOT NULL DEFAULT 0
);

-- Gas quota trigger on keystore insert to default 0 quota
-- We wait for the event handler to correctly set the chain quota
create function insert_gas_quota()
    returns trigger
as $$
begin
    insert into gas_quota (key_id) values (new.id);
    return new;
end;
$$ language plpgsql;

create trigger insert_gas_quota
    after insert on keystore
for each row
execute procedure insert_gas_quota()