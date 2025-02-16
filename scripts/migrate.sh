#!/bin/bash
PGPASSWORD=$DB_PASSWORD psql -h localhost -U $DB_USER -d $DB_NAME -f internal/db/migrations/001_initial_schema.sql
PGPASSWORD=$DB_PASSWORD psql -h localhost -U $DB_USER -d $DB_NAME -f internal/db/migrations/002_add_is_active.sql 