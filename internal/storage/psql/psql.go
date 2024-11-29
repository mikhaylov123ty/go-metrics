package psql

import (
	"database/sql"
	"errors"
	"fmt"

	"metrics/internal/storage"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DataBase struct {
	Instance *sql.DB
}

func NewPSQLDataBase(connectionString string) (*DataBase, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w, connection string %s", err, connectionString)
	}

	return &DataBase{Instance: db}, nil
}

func (db *DataBase) BootStrap(connectionString string) error {
	migration, err := migrate.New("file://./internal/storage/psql/migrations", connectionString)
	if err != nil {
		return fmt.Errorf("creation migration database: %w", err)
	}

	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrating up database: %w", err)
	}

	return nil
}

func (db *DataBase) Read(id string) (*storage.Data, error) {
	res := storage.Data{}

	query, args, err := sq.Select("type, name, value, delta").
		From("metrics").
		Where(sq.Eq{"unique_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building query data from metrics: %w", err)
	}

	row := db.Instance.QueryRow(query, args...)
	if err = row.Scan(&res.Type, &res.Name, &res.Value, &res.Delta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning data: %w", err)
	}

	return &res, nil
}

func (db *DataBase) ReadAll() ([]*storage.Data, error) {
	res := make([]*storage.Data, 0)

	query, args, err := sq.Select("type, name, value, delta").
		From("metrics").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building query metrics: %w", err)
	}

	rows, err := db.Instance.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying all metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		row := storage.Data{}
		if err = rows.Scan(&row.Type, &row.Name, &row.Value, &row.Delta); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		res = append(res, &row)
	}

	return res, nil
}

func (db *DataBase) Update(id string, query *storage.Data) error {
	_, err := db.Instance.Exec(`INSERT INTO metrics (unique_id, type, name, value, delta)
										VALUES($1,$2,$3,$4,$5)
ON CONFLICT (unique_id) DO UPDATE 
SET type = $2,
    name = $3,
    value = $4,
delta = $5
    ;`,
		sql.Named("id", id),
		sql.Named("type", query.Type),
		sql.Named("name", query.Name),
		sql.Named("value", query.Value),
		sql.Named("delta", query.Delta),
	)
	if err != nil {
		return fmt.Errorf("updating metrics: %w", err)
	}

	return nil
}

func (db *DataBase) Delete(id string) error {
	fmt.Println("Delete psql database")

	return nil
}

func (db *DataBase) Ping() error {
	if err := db.Instance.Ping(); err != nil {
		return err
	}

	return nil
}
