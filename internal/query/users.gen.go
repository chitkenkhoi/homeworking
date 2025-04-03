// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"lqkhoi-go-http-api/internal/models"
)

func newUser(db *gorm.DB, opts ...gen.DOOption) user {
	_user := user{}

	_user.userDo.UseDB(db, opts...)
	_user.userDo.UseModel(&models.User{})

	tableName := _user.userDo.TableName()
	_user.ALL = field.NewAsterisk(tableName)
	_user.ID = field.NewInt(tableName, "id")
	_user.CreatedAt = field.NewTime(tableName, "created_at")
	_user.UpdatedAt = field.NewTime(tableName, "updated_at")
	_user.DeletedAt = field.NewField(tableName, "deleted_at")
	_user.Email = field.NewString(tableName, "email")
	_user.Password = field.NewString(tableName, "password")
	_user.Role = field.NewString(tableName, "role")
	_user.FirstName = field.NewString(tableName, "first_name")
	_user.LastName = field.NewString(tableName, "last_name")
	_user.CurrentProjectID = field.NewInt(tableName, "current_project_id")
	_user.ManagedProjects = userHasManyManagedProjects{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("ManagedProjects", "models.Project"),
		Manager: struct {
			field.RelationField
			CurrentProject struct {
				field.RelationField
			}
			ManagedProjects struct {
				field.RelationField
			}
			AssignedTasks struct {
				field.RelationField
				Assignee struct {
					field.RelationField
				}
				Project struct {
					field.RelationField
				}
				Sprint struct {
					field.RelationField
					Project struct {
						field.RelationField
					}
					Tasks struct {
						field.RelationField
					}
				}
			}
		}{
			RelationField: field.NewRelation("ManagedProjects.Manager", "models.User"),
			CurrentProject: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("ManagedProjects.Manager.CurrentProject", "models.Project"),
			},
			ManagedProjects: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("ManagedProjects.Manager.ManagedProjects", "models.Project"),
			},
			AssignedTasks: struct {
				field.RelationField
				Assignee struct {
					field.RelationField
				}
				Project struct {
					field.RelationField
				}
				Sprint struct {
					field.RelationField
					Project struct {
						field.RelationField
					}
					Tasks struct {
						field.RelationField
					}
				}
			}{
				RelationField: field.NewRelation("ManagedProjects.Manager.AssignedTasks", "models.Task"),
				Assignee: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("ManagedProjects.Manager.AssignedTasks.Assignee", "models.User"),
				},
				Project: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("ManagedProjects.Manager.AssignedTasks.Project", "models.Project"),
				},
				Sprint: struct {
					field.RelationField
					Project struct {
						field.RelationField
					}
					Tasks struct {
						field.RelationField
					}
				}{
					RelationField: field.NewRelation("ManagedProjects.Manager.AssignedTasks.Sprint", "models.Sprint"),
					Project: struct {
						field.RelationField
					}{
						RelationField: field.NewRelation("ManagedProjects.Manager.AssignedTasks.Sprint.Project", "models.Project"),
					},
					Tasks: struct {
						field.RelationField
					}{
						RelationField: field.NewRelation("ManagedProjects.Manager.AssignedTasks.Sprint.Tasks", "models.Task"),
					},
				},
			},
		},
		Tasks: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("ManagedProjects.Tasks", "models.Task"),
		},
		Sprints: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("ManagedProjects.Sprints", "models.Sprint"),
		},
		TeamMembers: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("ManagedProjects.TeamMembers", "models.User"),
		},
	}

	_user.AssignedTasks = userHasManyAssignedTasks{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("AssignedTasks", "models.Task"),
	}

	_user.CurrentProject = userBelongsToCurrentProject{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("CurrentProject", "models.Project"),
	}

	_user.fillFieldMap()

	return _user
}

type user struct {
	userDo userDo

	ALL              field.Asterisk
	ID               field.Int
	CreatedAt        field.Time
	UpdatedAt        field.Time
	DeletedAt        field.Field
	Email            field.String
	Password         field.String
	Role             field.String
	FirstName        field.String
	LastName         field.String
	CurrentProjectID field.Int
	ManagedProjects  userHasManyManagedProjects

	AssignedTasks userHasManyAssignedTasks

	CurrentProject userBelongsToCurrentProject

	fieldMap map[string]field.Expr
}

func (u user) Table(newTableName string) *user {
	u.userDo.UseTable(newTableName)
	return u.updateTableName(newTableName)
}

func (u user) As(alias string) *user {
	u.userDo.DO = *(u.userDo.As(alias).(*gen.DO))
	return u.updateTableName(alias)
}

func (u *user) updateTableName(table string) *user {
	u.ALL = field.NewAsterisk(table)
	u.ID = field.NewInt(table, "id")
	u.CreatedAt = field.NewTime(table, "created_at")
	u.UpdatedAt = field.NewTime(table, "updated_at")
	u.DeletedAt = field.NewField(table, "deleted_at")
	u.Email = field.NewString(table, "email")
	u.Password = field.NewString(table, "password")
	u.Role = field.NewString(table, "role")
	u.FirstName = field.NewString(table, "first_name")
	u.LastName = field.NewString(table, "last_name")
	u.CurrentProjectID = field.NewInt(table, "current_project_id")

	u.fillFieldMap()

	return u
}

func (u *user) WithContext(ctx context.Context) IUserDo { return u.userDo.WithContext(ctx) }

func (u user) TableName() string { return u.userDo.TableName() }

func (u user) Alias() string { return u.userDo.Alias() }

func (u user) Columns(cols ...field.Expr) gen.Columns { return u.userDo.Columns(cols...) }

func (u *user) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := u.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (u *user) fillFieldMap() {
	u.fieldMap = make(map[string]field.Expr, 13)
	u.fieldMap["id"] = u.ID
	u.fieldMap["created_at"] = u.CreatedAt
	u.fieldMap["updated_at"] = u.UpdatedAt
	u.fieldMap["deleted_at"] = u.DeletedAt
	u.fieldMap["email"] = u.Email
	u.fieldMap["password"] = u.Password
	u.fieldMap["role"] = u.Role
	u.fieldMap["first_name"] = u.FirstName
	u.fieldMap["last_name"] = u.LastName
	u.fieldMap["current_project_id"] = u.CurrentProjectID

}

func (u user) clone(db *gorm.DB) user {
	u.userDo.ReplaceConnPool(db.Statement.ConnPool)
	return u
}

func (u user) replaceDB(db *gorm.DB) user {
	u.userDo.ReplaceDB(db)
	return u
}

type userHasManyManagedProjects struct {
	db *gorm.DB

	field.RelationField

	Manager struct {
		field.RelationField
		CurrentProject struct {
			field.RelationField
		}
		ManagedProjects struct {
			field.RelationField
		}
		AssignedTasks struct {
			field.RelationField
			Assignee struct {
				field.RelationField
			}
			Project struct {
				field.RelationField
			}
			Sprint struct {
				field.RelationField
				Project struct {
					field.RelationField
				}
				Tasks struct {
					field.RelationField
				}
			}
		}
	}
	Tasks struct {
		field.RelationField
	}
	Sprints struct {
		field.RelationField
	}
	TeamMembers struct {
		field.RelationField
	}
}

func (a userHasManyManagedProjects) Where(conds ...field.Expr) *userHasManyManagedProjects {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a userHasManyManagedProjects) WithContext(ctx context.Context) *userHasManyManagedProjects {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a userHasManyManagedProjects) Session(session *gorm.Session) *userHasManyManagedProjects {
	a.db = a.db.Session(session)
	return &a
}

func (a userHasManyManagedProjects) Model(m *models.User) *userHasManyManagedProjectsTx {
	return &userHasManyManagedProjectsTx{a.db.Model(m).Association(a.Name())}
}

type userHasManyManagedProjectsTx struct{ tx *gorm.Association }

func (a userHasManyManagedProjectsTx) Find() (result []*models.Project, err error) {
	return result, a.tx.Find(&result)
}

func (a userHasManyManagedProjectsTx) Append(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a userHasManyManagedProjectsTx) Replace(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a userHasManyManagedProjectsTx) Delete(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a userHasManyManagedProjectsTx) Clear() error {
	return a.tx.Clear()
}

func (a userHasManyManagedProjectsTx) Count() int64 {
	return a.tx.Count()
}

type userHasManyAssignedTasks struct {
	db *gorm.DB

	field.RelationField
}

func (a userHasManyAssignedTasks) Where(conds ...field.Expr) *userHasManyAssignedTasks {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a userHasManyAssignedTasks) WithContext(ctx context.Context) *userHasManyAssignedTasks {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a userHasManyAssignedTasks) Session(session *gorm.Session) *userHasManyAssignedTasks {
	a.db = a.db.Session(session)
	return &a
}

func (a userHasManyAssignedTasks) Model(m *models.User) *userHasManyAssignedTasksTx {
	return &userHasManyAssignedTasksTx{a.db.Model(m).Association(a.Name())}
}

type userHasManyAssignedTasksTx struct{ tx *gorm.Association }

func (a userHasManyAssignedTasksTx) Find() (result []*models.Task, err error) {
	return result, a.tx.Find(&result)
}

func (a userHasManyAssignedTasksTx) Append(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a userHasManyAssignedTasksTx) Replace(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a userHasManyAssignedTasksTx) Delete(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a userHasManyAssignedTasksTx) Clear() error {
	return a.tx.Clear()
}

func (a userHasManyAssignedTasksTx) Count() int64 {
	return a.tx.Count()
}

type userBelongsToCurrentProject struct {
	db *gorm.DB

	field.RelationField
}

func (a userBelongsToCurrentProject) Where(conds ...field.Expr) *userBelongsToCurrentProject {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a userBelongsToCurrentProject) WithContext(ctx context.Context) *userBelongsToCurrentProject {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a userBelongsToCurrentProject) Session(session *gorm.Session) *userBelongsToCurrentProject {
	a.db = a.db.Session(session)
	return &a
}

func (a userBelongsToCurrentProject) Model(m *models.User) *userBelongsToCurrentProjectTx {
	return &userBelongsToCurrentProjectTx{a.db.Model(m).Association(a.Name())}
}

type userBelongsToCurrentProjectTx struct{ tx *gorm.Association }

func (a userBelongsToCurrentProjectTx) Find() (result *models.Project, err error) {
	return result, a.tx.Find(&result)
}

func (a userBelongsToCurrentProjectTx) Append(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a userBelongsToCurrentProjectTx) Replace(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a userBelongsToCurrentProjectTx) Delete(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a userBelongsToCurrentProjectTx) Clear() error {
	return a.tx.Clear()
}

func (a userBelongsToCurrentProjectTx) Count() int64 {
	return a.tx.Count()
}

type userDo struct{ gen.DO }

type IUserDo interface {
	gen.SubQuery
	Debug() IUserDo
	WithContext(ctx context.Context) IUserDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IUserDo
	WriteDB() IUserDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IUserDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IUserDo
	Not(conds ...gen.Condition) IUserDo
	Or(conds ...gen.Condition) IUserDo
	Select(conds ...field.Expr) IUserDo
	Where(conds ...gen.Condition) IUserDo
	Order(conds ...field.Expr) IUserDo
	Distinct(cols ...field.Expr) IUserDo
	Omit(cols ...field.Expr) IUserDo
	Join(table schema.Tabler, on ...field.Expr) IUserDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IUserDo
	RightJoin(table schema.Tabler, on ...field.Expr) IUserDo
	Group(cols ...field.Expr) IUserDo
	Having(conds ...gen.Condition) IUserDo
	Limit(limit int) IUserDo
	Offset(offset int) IUserDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IUserDo
	Unscoped() IUserDo
	Create(values ...*models.User) error
	CreateInBatches(values []*models.User, batchSize int) error
	Save(values ...*models.User) error
	First() (*models.User, error)
	Take() (*models.User, error)
	Last() (*models.User, error)
	Find() ([]*models.User, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.User, err error)
	FindInBatches(result *[]*models.User, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*models.User) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IUserDo
	Assign(attrs ...field.AssignExpr) IUserDo
	Joins(fields ...field.RelationField) IUserDo
	Preload(fields ...field.RelationField) IUserDo
	FirstOrInit() (*models.User, error)
	FirstOrCreate() (*models.User, error)
	FindByPage(offset int, limit int) (result []*models.User, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IUserDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (u userDo) Debug() IUserDo {
	return u.withDO(u.DO.Debug())
}

func (u userDo) WithContext(ctx context.Context) IUserDo {
	return u.withDO(u.DO.WithContext(ctx))
}

func (u userDo) ReadDB() IUserDo {
	return u.Clauses(dbresolver.Read)
}

func (u userDo) WriteDB() IUserDo {
	return u.Clauses(dbresolver.Write)
}

func (u userDo) Session(config *gorm.Session) IUserDo {
	return u.withDO(u.DO.Session(config))
}

func (u userDo) Clauses(conds ...clause.Expression) IUserDo {
	return u.withDO(u.DO.Clauses(conds...))
}

func (u userDo) Returning(value interface{}, columns ...string) IUserDo {
	return u.withDO(u.DO.Returning(value, columns...))
}

func (u userDo) Not(conds ...gen.Condition) IUserDo {
	return u.withDO(u.DO.Not(conds...))
}

func (u userDo) Or(conds ...gen.Condition) IUserDo {
	return u.withDO(u.DO.Or(conds...))
}

func (u userDo) Select(conds ...field.Expr) IUserDo {
	return u.withDO(u.DO.Select(conds...))
}

func (u userDo) Where(conds ...gen.Condition) IUserDo {
	return u.withDO(u.DO.Where(conds...))
}

func (u userDo) Order(conds ...field.Expr) IUserDo {
	return u.withDO(u.DO.Order(conds...))
}

func (u userDo) Distinct(cols ...field.Expr) IUserDo {
	return u.withDO(u.DO.Distinct(cols...))
}

func (u userDo) Omit(cols ...field.Expr) IUserDo {
	return u.withDO(u.DO.Omit(cols...))
}

func (u userDo) Join(table schema.Tabler, on ...field.Expr) IUserDo {
	return u.withDO(u.DO.Join(table, on...))
}

func (u userDo) LeftJoin(table schema.Tabler, on ...field.Expr) IUserDo {
	return u.withDO(u.DO.LeftJoin(table, on...))
}

func (u userDo) RightJoin(table schema.Tabler, on ...field.Expr) IUserDo {
	return u.withDO(u.DO.RightJoin(table, on...))
}

func (u userDo) Group(cols ...field.Expr) IUserDo {
	return u.withDO(u.DO.Group(cols...))
}

func (u userDo) Having(conds ...gen.Condition) IUserDo {
	return u.withDO(u.DO.Having(conds...))
}

func (u userDo) Limit(limit int) IUserDo {
	return u.withDO(u.DO.Limit(limit))
}

func (u userDo) Offset(offset int) IUserDo {
	return u.withDO(u.DO.Offset(offset))
}

func (u userDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IUserDo {
	return u.withDO(u.DO.Scopes(funcs...))
}

func (u userDo) Unscoped() IUserDo {
	return u.withDO(u.DO.Unscoped())
}

func (u userDo) Create(values ...*models.User) error {
	if len(values) == 0 {
		return nil
	}
	return u.DO.Create(values)
}

func (u userDo) CreateInBatches(values []*models.User, batchSize int) error {
	return u.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (u userDo) Save(values ...*models.User) error {
	if len(values) == 0 {
		return nil
	}
	return u.DO.Save(values)
}

func (u userDo) First() (*models.User, error) {
	if result, err := u.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.User), nil
	}
}

func (u userDo) Take() (*models.User, error) {
	if result, err := u.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.User), nil
	}
}

func (u userDo) Last() (*models.User, error) {
	if result, err := u.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.User), nil
	}
}

func (u userDo) Find() ([]*models.User, error) {
	result, err := u.DO.Find()
	return result.([]*models.User), err
}

func (u userDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.User, err error) {
	buf := make([]*models.User, 0, batchSize)
	err = u.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (u userDo) FindInBatches(result *[]*models.User, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return u.DO.FindInBatches(result, batchSize, fc)
}

func (u userDo) Attrs(attrs ...field.AssignExpr) IUserDo {
	return u.withDO(u.DO.Attrs(attrs...))
}

func (u userDo) Assign(attrs ...field.AssignExpr) IUserDo {
	return u.withDO(u.DO.Assign(attrs...))
}

func (u userDo) Joins(fields ...field.RelationField) IUserDo {
	for _, _f := range fields {
		u = *u.withDO(u.DO.Joins(_f))
	}
	return &u
}

func (u userDo) Preload(fields ...field.RelationField) IUserDo {
	for _, _f := range fields {
		u = *u.withDO(u.DO.Preload(_f))
	}
	return &u
}

func (u userDo) FirstOrInit() (*models.User, error) {
	if result, err := u.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.User), nil
	}
}

func (u userDo) FirstOrCreate() (*models.User, error) {
	if result, err := u.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.User), nil
	}
}

func (u userDo) FindByPage(offset int, limit int) (result []*models.User, count int64, err error) {
	result, err = u.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = u.Offset(-1).Limit(-1).Count()
	return
}

func (u userDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = u.Count()
	if err != nil {
		return
	}

	err = u.Offset(offset).Limit(limit).Scan(result)
	return
}

func (u userDo) Scan(result interface{}) (err error) {
	return u.DO.Scan(result)
}

func (u userDo) Delete(models ...*models.User) (result gen.ResultInfo, err error) {
	return u.DO.Delete(models)
}

func (u *userDo) withDO(do gen.Dao) *userDo {
	u.DO = *do.(*gen.DO)
	return u
}
