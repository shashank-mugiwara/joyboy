package database

import "gorm.io/gorm"

// TaskRepository defines the interface for all database operations
// currently performed via *gorm.DB. This abstraction allows for
// easier testing and future database implementation changes.
type TaskRepository interface {
	// Save inserts or updates a record in the database
	Save(value interface{}) error

	// Find retrieves records matching the given conditions
	Find(dest interface{}, conditions ...interface{}) error

	// Where returns a RepositoryQuery for building chainable queries
	Where(query interface{}, args ...interface{}) RepositoryQuery

	// First retrieves the first record matching the given conditions
	First(dest interface{}, conditions ...interface{}) error

	// Delete removes a record from the database
	Delete(value interface{}) error

	// AutoMigrate runs automatic migration for the given models
	AutoMigrate(models ...interface{}) error

	// Model specifies the model for the query, returns a chainable query
	Model(value interface{}) RepositoryQuery
}

// RepositoryQuery defines the interface for chainable query operations
// used in patterns like Where().Find() and Model().Where().Update()
type RepositoryQuery interface {
	// Find retrieves records matching the query
	Find(dest interface{}) error

	// Where adds a WHERE clause to the query
	Where(query interface{}, args ...interface{}) RepositoryQuery

	// Take retrieves a single record, similar to First but without ordering
	Take(dest interface{}) error

	// Update updates a column with a value
	Update(column string, value interface{}) error

	// RowsAffected returns the number of rows affected by the query
	RowsAffected() int64

	// Error returns any error that occurred during the query
	Error() error
}

// GORMRepository is a concrete implementation of TaskRepository
// that wraps a *gorm.DB instance and delegates all operations to it.
type GORMRepository struct {
	db *gorm.DB
}

// NewGORMRepository creates a new GORMRepository instance
func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

// Save inserts or updates a record in the database
func (r *GORMRepository) Save(value interface{}) error {
	result := r.db.Save(value)
	return result.Error
}

// Find retrieves records matching the given conditions
func (r *GORMRepository) Find(dest interface{}, conditions ...interface{}) error {
	result := r.db.Find(dest, conditions...)
	return result.Error
}

// Where returns a RepositoryQuery for building chainable queries
func (r *GORMRepository) Where(query interface{}, args ...interface{}) RepositoryQuery {
	return &GORMQuery{db: r.db.Where(query, args...)}
}

// First retrieves the first record matching the given conditions
func (r *GORMRepository) First(dest interface{}, conditions ...interface{}) error {
	result := r.db.First(dest, conditions...)
	return result.Error
}

// Delete removes a record from the database
func (r *GORMRepository) Delete(value interface{}) error {
	result := r.db.Delete(value)
	return result.Error
}

// AutoMigrate runs automatic migration for the given models
func (r *GORMRepository) AutoMigrate(models ...interface{}) error {
	return r.db.AutoMigrate(models...)
}

// Model specifies the model for the query, returns a chainable query
func (r *GORMRepository) Model(value interface{}) RepositoryQuery {
	return &GORMQuery{db: r.db.Model(value)}
}

// GORMQuery is a concrete implementation of RepositoryQuery
// that wraps a *gorm.DB instance for chainable operations.
type GORMQuery struct {
	db *gorm.DB
}

// Find retrieves records matching the query
func (q *GORMQuery) Find(dest interface{}) error {
	result := q.db.Find(dest)
	return result.Error
}

// Where adds a WHERE clause to the query
func (q *GORMQuery) Where(query interface{}, args ...interface{}) RepositoryQuery {
	return &GORMQuery{db: q.db.Where(query, args...)}
}

// Take retrieves a single record, similar to First but without ordering
func (q *GORMQuery) Take(dest interface{}) error {
	result := q.db.Take(dest)
	return result.Error
}

// Update updates a column with a value
func (q *GORMQuery) Update(column string, value interface{}) error {
	result := q.db.Update(column, value)
	return result.Error
}

// RowsAffected returns the number of rows affected by the query
func (q *GORMQuery) RowsAffected() int64 {
	return q.db.RowsAffected
}

// Error returns any error that occurred during the query
func (q *GORMQuery) Error() error {
	return q.db.Error
}
