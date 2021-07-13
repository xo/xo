package main

import (
	"context"
	"database/sql"
	"fmt"

	models "github.com/xo/xo/_examples/a_bit_of_everything/oracle"
)

func runOracle(ctx context.Context, db *sql.DB) error {
	var res1, res2 int64
	var err error
	// run procs
	// if err := models.A0In0Out(ctx, db); err != nil {
	// 	return fmt.Errorf("0 in 0 out: %v", err)
	// }
	// 1 in 0 out
	if err := models.A1In0Out(ctx, db, 10); err != nil {
		return fmt.Errorf("1 in 0 out: %v", err)
	}
	// 0 in 1 out
	if res1, err = models.A0In1Out(ctx, db); err != nil {
		return fmt.Errorf("0 in 1 out: %v", err)
	}
	fmt.Printf("a_0_in_1_out(): %d\n", res1)
	// 1 in 1 out
	if res1, err = models.A1In1Out(ctx, db, 10); err != nil {
		return fmt.Errorf("1 in 0 out: %v", err)
	}
	fmt.Printf("a_1_in_1_out(%d): %d\n", 10, res1)
	// 2 in 2 out
	if res1, res2, err = models.A2In2Out(ctx, db, 10, 20); err != nil {
		return fmt.Errorf("2 in 2 out: %v", err)
	}
	fmt.Printf("a_2_in_2_out(%d, %d): %d, %d\n", 10, 20, res1, res2)
	// run funcs
	// 0 in
	if res1, err = models.AFunc0In(ctx, db); err != nil {
		return fmt.Errorf("a func 0 in: %v", err)
	}
	fmt.Printf("a_func_0_in(): %d\n", res1)
	// 1 in
	if res1, err = models.AFunc1In(ctx, db, 10); err != nil {
		return fmt.Errorf("a func 1 in: %v", err)
	}
	fmt.Printf("a_func_1_in(%d): %d\n", 10, res1)
	// 1 in
	if res1, err = models.AFunc2In(ctx, db, 10, 20); err != nil {
		return fmt.Errorf("a func 2 in: %v", err)
	}
	fmt.Printf("a_func_2_in(%d, %d): %d\n", 10, 20, res1)
	return nil
}
