#!/bin/bash

DB="mysql://dimas:dimas@tcp(localhost:3306)/inventory"

migrate -path migrations -database "$DB" down