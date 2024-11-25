package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DataBase struct {
	DB *sql.DB
}

func NewPSQLDataBase(connectionString string) (*DataBase, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w, connection string %s", err, connectionString)
	}

	return &DataBase{DB: db}, nil
}

func (d *DataBase) Read(id string) (*Data, error) {

	fmt.Println("Read psql database")

	return nil, nil
}

func (d *DataBase) ReadAll() ([]*Data, error) {
	fmt.Println("ReadAll psql database")

	return nil, nil
}

func (d *DataBase) Update(id string, query *Data) error {
	fmt.Println("Update psql database")

	return nil
}

func (d *DataBase) Delete(id string) error {
	fmt.Println("Delete psql database")

	return nil
}

func (d *DataBase) Ping() error {
	if err := d.DB.Ping(); err != nil {
		return err
	}

	return nil
}
