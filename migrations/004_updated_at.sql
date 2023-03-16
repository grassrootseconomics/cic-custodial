ALTER TABLE keystore
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE otx_dispatch
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE gas_quota
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- updated_at function
create function update_timestamp()
    returns trigger
as $$
begin
    new.updated_at = current_timestamp;
    return new;
end;
$$ language plpgsql;

create trigger update_keystore_timestamp
    before update on keystore
for each row
execute procedure update_timestamp();

create trigger update_otx_dispatch_timestamp
    before update on otx_dispatch
for each row
execute procedure update_timestamp();

create trigger update_gas_quota_timestamp
    before update on gas_quota
for each row
execute procedure update_timestamp();