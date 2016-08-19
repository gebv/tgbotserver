package main

import (
	"gopkg.in/pg.v4"
)

func setupDatabase(dbuser, dbpass, dbaddr, dbname string) error {
	db = pg.Connect(&pg.Options{
		User:     dbuser,
		Password: dbpass,
		Addr:     dbaddr,
		Database: dbname,
	})

	err := appCreateSchema(db)
	if err != nil {
		return err
	}

	return nil
}
