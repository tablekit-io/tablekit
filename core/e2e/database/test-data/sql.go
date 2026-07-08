package main

import _ "embed"

// schemaSQL is the pg_dump preamble + CREATE TABLE block, carried verbatim.
//go:embed schema.sql
var schemaSQL string

// constraintsSQL is the trailing PK/UNIQUE constraint block, carried verbatim.
//go:embed constraints.sql
var constraintsSQL string
