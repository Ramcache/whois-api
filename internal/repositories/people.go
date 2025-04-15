package repositories

import (
	"context"
	"errors"
	"whois-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PeopleRepositoryInterface interface {
	CreatePeople(ctx context.Context, p models.People) (int, error)
	GetPeople(ctx context.Context, name string, age int, limit, offset int) ([]models.People, error)
	GetPeopleByID(ctx context.Context, id int) (models.People, error)
	UpdatePeople(ctx context.Context, id int, p models.People) error
	DeletePeople(ctx context.Context, id int) error
}

type PeopleRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPeopleRepository(db *pgxpool.Pool, logger *zap.Logger) *PeopleRepository {
	return &PeopleRepository{db: db, logger: logger}
}

func (r *PeopleRepository) CreatePeople(ctx context.Context, people models.People) (int, error) {
	r.logger.Debug("Creating people", zap.Any("payload", people))

	query := `
		INSERT INTO people (name, surname, patronymic, age, gender, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := r.db.QueryRow(
		ctx,
		query,
		people.Name,
		people.Surname,
		people.Patronymic,
		people.Age,
		people.Gender,
		people.Nationality,
	).Scan(&id)

	if err != nil {
		r.logger.Error("Failed to create people", zap.Error(err))
		return 0, err
	}

	r.logger.Info("People created", zap.Int("id", id))
	return id, nil
}

func (r *PeopleRepository) GetPeople(ctx context.Context, name string, age int, limit, offset int) ([]models.People, error) {
	r.logger.Debug("Getting people", zap.String("name", name), zap.Int("age", age), zap.Int("limit", limit), zap.Int("offset", offset))

	query := `
		SELECT id, name, surname, patronymic, age, gender, nationality
		FROM people
		WHERE ($1 = '' OR name ILIKE $1)
		AND ($2 = 0 OR age = $2)
		LIMIT $3 OFFSET $4
	`

	nameFilter := name + "%"

	rows, err := r.db.Query(ctx, query, nameFilter, age, limit, offset)
	if err != nil {
		r.logger.Error("Query failed in GetPeople", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var people []models.People
	for rows.Next() {
		var p models.People
		err := rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic, &p.Age, &p.Gender, &p.Nationality)
		if err != nil {
			r.logger.Error("Row scan failed in GetPeople", zap.Error(err))
			return nil, err
		}
		people = append(people, p)
	}

	r.logger.Info("People fetched", zap.Int("count", len(people)))
	return people, nil
}

func (r *PeopleRepository) GetPeopleByID(ctx context.Context, id int) (models.People, error) {
	r.logger.Debug("Fetching people by ID", zap.Int("id", id))

	query := `
		SELECT id, name, surname, patronymic, age, gender, nationality
		FROM people
		WHERE id = $1
	`

	var p models.People
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Surname, &p.Patronymic, &p.Age, &p.Gender, &p.Nationality,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("People not found", zap.Int("id", id))
		} else {
			r.logger.Error("Failed to fetch people by ID", zap.Int("id", id), zap.Error(err))
		}
		return p, err
	}

	r.logger.Info("People fetched", zap.Int("id", p.ID))
	return p, nil
}

func (r *PeopleRepository) UpdatePeople(ctx context.Context, id int, p models.People) error {
	r.logger.Debug("Updating people", zap.Int("id", id), zap.Any("payload", p))

	query := `
		UPDATE people
		SET name = $1,
			surname = $2,
			patronymic = $3,
			age = $4,
			gender = $5,
			nationality = $6
		WHERE id = $7
	`

	cmdTag, err := r.db.Exec(ctx, query,
		p.Name,
		p.Surname,
		p.Patronymic,
		p.Age,
		p.Gender,
		p.Nationality,
		id,
	)

	if err != nil {
		r.logger.Error("Failed to update people", zap.Int("id", id), zap.Error(err))
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		r.logger.Warn("People not found for update", zap.Int("id", id))
		return pgx.ErrNoRows
	}

	r.logger.Info("People updated", zap.Int("id", id))
	return nil
}

func (r *PeopleRepository) DeletePeople(ctx context.Context, id int) error {
	r.logger.Debug("Deleting people", zap.Int("id", id))

	query := `DELETE FROM people WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)

	if err != nil {
		r.logger.Error("Failed to delete people", zap.Int("id", id), zap.Error(err))
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		r.logger.Warn("People not found for deletion", zap.Int("id", id))
		return pgx.ErrNoRows
	}

	r.logger.Info("People deleted", zap.Int("id", id))
	return nil
}
