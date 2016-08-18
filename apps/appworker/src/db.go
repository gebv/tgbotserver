package main

import (
	"gopkg.in/pg.v4"
)

var db *pg.DB

func setupDatabase(network, addr, name, user, pass string) error {
	db = pg.Connect(&pg.Options{
		User:     user,
		Password: pass,
		Addr:     addr,
		Database: name,
		Network:  network,
	})

	err := createSchema(db)
	if err != nil {
		return err
	}

	return nil
}

func createSchema(db *pg.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id bigint NOT NULL PRIMARY KEY, 
            fname text NOT NULL, 
            lname text,
            uname text,
			phone text,

            page text NOT NULL,
            section text NOT NULL,
            state jsonb NOT NULL DEFAULT '{}',
            
            is_open_dialog boolean DEFAULT false,
			dialog_id bigint,

            created timestamp with time zone NOT NULL,
	        updated timestamp with time zone DEFAULT now() NOT NULL
            )`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)

		if err != nil {
			return err
		}
	}

	return nil
}
