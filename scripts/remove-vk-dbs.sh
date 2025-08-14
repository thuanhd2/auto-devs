#!/bin/bash
# check all databases that has name starts with auto_devs_vk and remove them all
PREFFIX="auto_devs_vk"
db_list=$(psql -d postgres -c "SELECT datname FROM pg_database WHERE datname LIKE '$PREFFIX%'")
for db in $db_list; do
    # do not process if db is not starts with PREFFIX
    if [[ "$db" == $PREFFIX* ]]; then
        echo "Removing database: $db"
        psql -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$db' AND pid <> pg_backend_pid();" && dropdb $db
    fi
done