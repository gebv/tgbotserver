package checkerbot

import pg "gopkg.in/pg.v4"

func CreateSchema(db *pg.DB) error {
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

	        updated timestamp with time zone DEFAULT now() NOT NULL
            )`,
		`CREATE TABLE IF NOT EXISTS visions (
			id text NOT NULL PRIMARY KEY, 
			owner_id int NOT NULL,
			tone text NOT NULL,
			description text NOT NULL,
			tags text[],
			status text NOT NULL,
			status_reason text,
			enabled boolean DEFAULT false
		)`,
		`CREATE TABLE IF NOT EXISTS vision_photos (
			id text NOT NULL PRIMARY KEY,
			ext_id text NOT NULL,
			vision_id text NOT NULL,
			w int8,
			h int8,
			size int8
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
