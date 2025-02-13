package psql

import (
	"database/sql"
	"errors"
	"fmt"

	"metrics/internal/storage"
	"metrics/pkg"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	migrateFilesPath = "file://./internal/storage/psql/migrations"
)

// Структура хранилища
type DataBase struct {
	Instance *sql.DB
}

// Конструктор PostgreSQL
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

// Метод подготовки БД
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

// Метод получения записи из хранилища по id
func (db *DataBase) Read(name string) (*storage.Data, error) {
	res := storage.Data{}

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

// Метод получения записей из хранилища
func (db *DataBase) ReadAll() ([]*storage.Data, error) {
	res := make([]*storage.Data, 0)

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
	defer rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	// Сканирование строк
	for rows.Next() {
		row := storage.Data{}
		if err = rows.Scan(&row.Type, &row.Name, &row.Value, &row.Delta); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		res = append(res, &row)
	}

	return res, nil
}

// Метод создания или обновления существующей записи в хранилище
func (db *DataBase) Update(query *storage.Data) error {
	// Начало транзакции
	tx, err := db.Instance.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()
	
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

// Метод создания или обновление существующих записей в хранилище
func (db *DataBase) UpdateBatch(queries []*storage.Data) error {
	// Начало транзакции
	tx, err := db.Instance.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	// Парсинг запроса в контексте транзакции
	statement, err := tx.Prepare(`
		INSERT INTO metrics (name, type, value, delta)
		VALUES($1,$2,$3,$4)
		ON CONFLICT (name) DO UPDATE 
		SET 
			value = excluded.value,
			delta = metrics.delta + excluded.delta;`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("preparing transaction: %w", err)
	}
	defer statement.Close()

	// Проход по метрикам и запись в базу
	for _, query := range queries {
		if _, err = statement.Exec(
			query.Name,
			query.Type,
			query.Value,
			query.Delta); err != nil {
			tx.Rollback()
			return fmt.Errorf("updating metric: %w", err)
		}
	}

	// Коммит транзакции
	return tx.Commit()
}

// Метод проверки доступности БД
func (db *DataBase) Ping() error {
	if err := pkg.AnyFunc(db.Instance.Ping).WithRetry(); err != nil {
		return err
	}

	return nil
}
