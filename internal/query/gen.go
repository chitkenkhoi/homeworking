// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"gorm.io/gen"

	"gorm.io/plugin/dbresolver"
)

var (
	Q       = new(Query)
	Project *project
	Sprint  *sprint
	Task    *task
	User    *user
)

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	*Q = *Use(db, opts...)
	Project = &Q.Project
	Sprint = &Q.Sprint
	Task = &Q.Task
	User = &Q.User
}

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:      db,
		Project: newProject(db, opts...),
		Sprint:  newSprint(db, opts...),
		Task:    newTask(db, opts...),
		User:    newUser(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	Project project
	Sprint  sprint
	Task    task
	User    user
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:      db,
		Project: q.Project.clone(db),
		Sprint:  q.Sprint.clone(db),
		Task:    q.Task.clone(db),
		User:    q.User.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:      db,
		Project: q.Project.replaceDB(db),
		Sprint:  q.Sprint.replaceDB(db),
		Task:    q.Task.replaceDB(db),
		User:    q.User.replaceDB(db),
	}
}

type queryCtx struct {
	Project IProjectDo
	Sprint  ISprintDo
	Task    ITaskDo
	User    IUserDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		Project: q.Project.WithContext(ctx),
		Sprint:  q.Sprint.WithContext(ctx),
		Task:    q.Task.WithContext(ctx),
		User:    q.User.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	tx := q.db.Begin(opts...)
	return &QueryTx{Query: q.clone(tx), Error: tx.Error}
}

type QueryTx struct {
	*Query
	Error error
}

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}
