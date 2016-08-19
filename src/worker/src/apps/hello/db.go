package hello

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
	}

	for _, q := range queries {
		_, err := db.Exec(q)

		if err != nil {
			return err
		}
	}

	return nil
}
