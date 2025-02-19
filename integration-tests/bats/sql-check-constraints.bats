#!/usr/bin/env bats                                                             
load $BATS_TEST_DIRNAME/helper/common.bash

setup() {
    setup_common
}

teardown() {
    teardown_common
}


@test "sql-check-constraints: basic tests for check constraints" {
    dolt sql <<SQL
CREATE table t1 (
       a INTEGER PRIMARY KEY check (a > 3),
       b INTEGER check (b > a)
);
SQL

    dolt sql -q "insert into t1 values (5, 6)"

    run dolt sql -q "insert into t1 values (3, 4)"
    [ $status -eq 1 ]
    [[ "$output" =~ "constraint" ]] || false

    run dolt sql -q "insert into t1 values (4, 2)"
    [ $status -eq 1 ]
    [[ "$output" =~ "constraint" ]] || false

    dolt sql <<SQL
CREATE table t2 (
       a INTEGER PRIMARY KEY,
       b INTEGER
);
ALTER TABLE t2 ADD CONSTRAINT chk1 CHECK (a > 3);
ALTER TABLE t2 ADD CONSTRAINT chk2 CHECK (b > a);
SQL

    dolt sql -q "insert into t2 values (5, 6)"
    dolt sql -q "insert into t2 values (6, NULL)"

    run dolt sql -q "insert into t2 values (3, 4)"
    [ $status -eq 1 ]
    [[ "$output" =~ "constraint" ]] || false

    run dolt sql -q "insert into t2 values (4, 2)"
    [ $status -eq 1 ]
    [[ "$output" =~ "constraint" ]] || false

    dolt sql -q "ALTER TABLE t2 DROP CONSTRAINT chk1;"
    dolt sql -q "insert into t2 values (3, 4)"
    
    run dolt sql -q "insert into t2 values (4, 2)"
    [ $status -eq 1 ]
    [[ "$output" =~ "constraint" ]] || false

    dolt sql -q "ALTER TABLE t2 DROP CONSTRAINT chk2;"    
    dolt sql -q "insert into t2 values (4, 2)"

    # t1 should still have its constraints
    run dolt sql -q "insert into t1 values (4, 2)"
    [ $status -eq 1 ]
    [[ "$output" =~ "constraint" ]] || false
}

@test "sql-check-constraints: check constraints survive adding a primary key" {
    dolt sql <<SQL 
create table foo (
       pk int,
       c1 int
       CHECK (c1 > 3)
);      
ALTER TABLE foo ADD PRIMARY KEY(pk);
SQL
    skip "Alter tables kill all constraints now"
    run dolt schema show
    [ $status -eq 0 ]
    [[ "$output" =~ "CHECK" ]] || false
    [[ "$output" =~ "c1 > c3" ]] || false
    
}

