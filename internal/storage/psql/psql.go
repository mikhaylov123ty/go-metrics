package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"metrics/internal/models"
	"metrics/pkg"
)

const (
	migrateFilesPath = "file://./internal/storage/psql/migrations"
)

// DataBase - структура инстанса хранилища
type DataBase struct {
	Instance *sql.DB
}

// NewPSQLDataBase - конструктор PostgreSQL
func NewPSQLDataBase(connectionString string) (*DataBase, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %w", err)
	}

	if err = pkg.AnyFunc(db.Ping).WithRetry(); err != nil {
		return nil, fmt.Errorf("pinging database: %w, connection string %s", err, connectionString)
	}

	return &DataBase{Instance: db}, nil
}

// BootStrap подготовка БД
func (db *DataBase) BootStrap(connectionString string) error {
	// Создание новой миграции
	migration, err := migrate.New(migrateFilesPath, connectionString)
	if err != nil {
		return fmt.Errorf("creation migration database: %w", err)
	}

	// Запуск миграции
	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrating up database: %w", err)
	}

	return nil
}

// Read получает метрику из хранилища по названию
func (db *DataBase) Read(name string) (*models.Data, error) {
	res := models.Data{}

	// Формирование строки запроса и аргументов
	query, args, err := sq.Select("type, name, value, delta").
		From("metrics").
		Where(sq.Eq{"name": name}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building query data from metrics: %w", err)
	}

	// Запрос в базу
	row := db.Instance.QueryRow(query, args...)
	if err = row.Scan(&res.Type, &res.Name, &res.Value, &res.Delta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning data: %w", err)
	}

	return &res, nil
}

// ReadAll получает все метрики из хранилища
func (db *DataBase) ReadAll() ([]*models.Data, error) {
	res := make([]*models.Data, 0)

	// Формирование строки запроса и аргументов
	query, args, err := sq.Select("type, name, value, delta").
		From("metrics").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building query metrics: %w", err)
	}

	// Выполнение запроса
	rows, err := db.Instance.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying all metrics: %w", err)
	}

	defer func() {
		if err = rows.Close(); err != nil {
			log.Printf("Read All Closing Rows Error: %v", err)
		}
	}()

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	// Сканирование строк
	for rows.Next() {
		row := models.Data{}
		if err = rows.Scan(&row.Type, &row.Name, &row.Value, &row.Delta); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		res = append(res, &row)
	}

	return res, nil
}

// Update создает новую или обновляет существующую запись метрики в хранилище
func (db *DataBase) Update(query *models.Data) error {
	// Начало транзакции
	tx, err := db.Instance.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	defer func() {
		if err = tx.Rollback(); err != nil {
			log.Printf("Update Rollback Error: %v", err)
		}
	}()

	// Выполнение запроса
	if _, err = tx.Exec(`
		INSERT INTO metrics (name, type, value, delta)
		VALUES($1,$2,$3,$4)
		ON CONFLICT (name) DO UPDATE 
		SET
			value = excluded.value,
			delta = metrics.delta + excluded.delta;`,
		query.Name,
		query.Type,
		query.Value,
		query.Delta); err != nil {
		return fmt.Errorf("updating metrics: %w", err)
	}

	// Коммит транзакции
	return tx.Commit()
}

// UpdateBatch создает новые или обновляет существующие записи метрики в хранилище
func (db *DataBase) UpdateBatch(queries []*models.Data) error {
	// Начало транзакции
	tx, err := db.Instance.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	defer func() {
		if err = tx.Rollback(); err != nil {
			log.Printf("Update Rollback Error: %v", err)
		}
	}()

	// Парсинг запроса в контексте транзакции
	statement, err := tx.Prepare(`
		INSERT INTO metrics (name, type, value, delta)
		VALUES($1,$2,$3,$4)
		ON CONFLICT (name) DO UPDATE 
		SET 
			value = excluded.value,
			delta = metrics.delta + excluded.delta;`)
	if err != nil {
		return fmt.Errorf("preparing transaction: %w", err)
	}

	defer func() {
		if err = statement.Close(); err != nil {
			log.Printf("Update Close Statement Error: %v", err)
		}
	}()

	// Проход по метрикам и запись в базу
	for _, query := range queries {
		if _, err = statement.Exec(
			query.Name,
			query.Type,
			query.Value,
			query.Delta); err != nil {
			return fmt.Errorf("updating metric: %w", err)
		}
	}

	// Коммит транзакции
	return tx.Commit()
}

// Ping проверяет доступность БД
func (db *DataBase) Ping() error {
	if err := pkg.AnyFunc(db.Instance.Ping).WithRetry(); err != nil {
		return err
	}

	return nil
}
