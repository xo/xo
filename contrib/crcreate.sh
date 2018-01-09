#!/bin/bash
#
# Creation script for CockroachDB. 
# Can be used for setting up for the examples/booktest/cockroachdb tests.
# First install CockRoachDB: https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html
#
# Setup the instance and initialize the DB:
cockroach start --insecure --store=booktest --host=localhost --background

cockroach user set booktest --insecure

cockroach sql --insecure -e 'CREATE DATABASE public'

cockroach sql --insecure -e 'GRANT ALL ON DATABASE public TO booktest'