package main

import (
	"fmt"
	"os"

	"github.com/tecbot/gorocksdb"
)

func main() {
	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(false)

	db, err := gorocksdb.OpenDbForReadOnly(opts, "bin/rocksdb_storage", false)
	if err != nil {
		panic(err)
	}

	data, err := db.Get(gorocksdb.NewDefaultReadOptions(), []byte(os.Args[1]))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data.Data()))
}
