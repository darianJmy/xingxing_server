package main

import (
	"fmt"
	"xingxing_server/cmd/dbstone"
)

func main() {
	db := dbstone.DB
	fmt.Println(db)
}