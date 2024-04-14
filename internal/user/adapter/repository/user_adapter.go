package repository

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/core-go/search/query"
	s "github.com/core-go/sql"

	"go-service/internal/user/domain"
)

func NewUserAdapter(db *sql.DB, buildQuery func(*domain.UserFilter) (string, []interface{})) (*UserAdapter, error) {
	userType := reflect.TypeOf(domain.User{})
	if buildQuery == nil {
		userQueryBuilder := query.NewBuilder(db, "users", userType)
		buildQuery = func(filter *domain.UserFilter) (s string, i []interface{}) {
			return userQueryBuilder.BuildQuery(filter)
		}
	}
	fieldsIndex, err := s.GetColumnIndexes(userType)
	if err != nil {
		return nil, err
	}
	return &UserAdapter{DB: db, Map: fieldsIndex, BuildQuery: buildQuery}, nil
}

type UserAdapter struct {
	DB         *sql.DB
	Map        map[string]int
	BuildQuery func(*domain.UserFilter) (string, []interface{})
}

func (r *UserAdapter) Load(ctx context.Context, id string) (*domain.User, error) {
	query := `
		select
			id, 
			username,
			email,
			phone,
			date_of_birth
		from users where id = $1`
	rows, err := r.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user domain.User
		err = rows.Scan(
			&user.Id,
			&user.Username,
			&user.Email,
			&user.Phone,
			&user.DateOfBirth)
		return &user, nil
	}
	return nil, nil
}
func (r *UserAdapter) Create(ctx context.Context, user *domain.User) (int64, error) {
	query := `
		insert into users (
			id,
			username,
			email,
			phone,
			date_of_birth)
		values (
			$1,
			$2,
			$3, 
			$4,
			$5)`
	tx := GetTx(ctx)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return -1, err
	}
	res, err := stmt.ExecContext(ctx,
		user.Id,
		user.Username,
		user.Email,
		user.Phone,
		user.DateOfBirth)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
func (r *UserAdapter) Update(ctx context.Context, user *domain.User) (int64, error) {
	query := `
		update users 
		set
			username = $1,
			email = $2,
			phone = $3,
			date_of_birth = $4
		where id = $5`
	tx := GetTx(ctx)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return -1, err
	}
	res, err := stmt.ExecContext(ctx,
		user.Username,
		user.Email,
		user.Phone,
		user.DateOfBirth,
		user.Id)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
func (r *UserAdapter) Patch(ctx context.Context, user map[string]interface{}) (int64, error) {
	updateClause := "update users set"
	whereClause := fmt.Sprintf("where id='%s'", user["id"])

	setClause := make([]string, 0)
	if user["username"] != nil {
		msg := fmt.Sprintf("username='%s'", fmt.Sprint(user["username"]))
		setClause = append(setClause, msg)
	}
	if user["email"] != nil {
		msg := fmt.Sprintf("email='%s'", fmt.Sprint(user["email"]))
		setClause = append(setClause, msg)
	}
	if user["phone"] != nil {
		msg := fmt.Sprintf("phone='%s'", fmt.Sprint(user["phone"]))
		setClause = append(setClause, msg)
	}
	setClauseRes := strings.Join(setClause, ",")
	querySlice := []string{updateClause, setClauseRes, whereClause}
	query := strings.Join(querySlice, " ")

	tx := GetTx(ctx)
	res, err := tx.ExecContext(ctx, query)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
func (r *UserAdapter) Delete(ctx context.Context, id string) (int64, error) {
	query := "delete from users where id = $1"
	tx := GetTx(ctx)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return -1, err
	}
	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
func (r *UserAdapter) Search(ctx context.Context, filter *domain.UserFilter) ([]domain.User, int64, error) {
	var users []domain.User
	if filter.Limit <= 0 {
		return users, 0, nil
	}
	query, params := r.BuildQuery(filter)
	offset := s.GetOffset(filter.Limit, filter.Page)
	pagingQuery := s.BuildPagingQuery(query, filter.Limit, offset)
	countQuery := s.BuildCountQuery(query)

	row := r.DB.QueryRowContext(ctx, countQuery, params...)
	if row.Err() != nil {
		return users, 0, row.Err()
	}
	var total int64
	err := row.Scan(&total)
	if err != nil || total == 0 {
		return users, total, err
	}

	err = s.Query(ctx, r.DB, r.Map, &users, pagingQuery, params...)
	return users, total, err
}

func GetTx(ctx context.Context) *sql.Tx {
	t := ctx.Value("tx")
	if t != nil {
		tx, ok := t.(*sql.Tx)
		if ok {
			return tx
		}
	}
	return nil
}

func BuildQuery(filter *domain.UserFilter) (string, []interface{}) {
	query := "select * from users"
	where, params := BuildFilter(filter)
	if len(where) > 0 {
		query = query + " where " + where
	}
	return query, params
}
func BuildFilter(filter *domain.UserFilter) (string, []interface{}) {
	buildParam := s.BuildDollarParam
	var where []string
	var params []interface{}
	i := 1
	if len(filter.Id) > 0 {
		params = append(params, filter.Id)
		where = append(where, fmt.Sprintf(`id = %s`, buildParam(i)))
		i++
	}
	if filter.DateOfBirth != nil {
		if filter.DateOfBirth.Min != nil {
			params = append(params, filter.DateOfBirth.Min)
			where = append(where, fmt.Sprintf(`date_of_birth >= %s`, buildParam(i)))
			i++
		}
		if filter.DateOfBirth.Max != nil {
			params = append(params, filter.DateOfBirth.Max)
			where = append(where, fmt.Sprintf(`date_of_birth <= %s`, buildParam(i)))
			i++
		}
	}
	if len(filter.Username) > 0 {
		q := filter.Username + "%"
		params = append(params, q)
		where = append(where, fmt.Sprintf(`username like %s`, buildParam(i)))
		i++
	}
	if len(filter.Email) > 0 {
		q := filter.Email + "%"
		params = append(params, q)
		where = append(where, fmt.Sprintf(`email like %s`, buildParam(i)))
		i++
	}
	if len(filter.Phone) > 0 {
		q := "%" + filter.Phone + "%"
		params = append(params, q)
		where = append(where, fmt.Sprintf(`phone like %s`, buildParam(i)))
		i++
	}
	if len(where) > 0 {
		return strings.Join(where, " and "), params
	}
	return "", params
}
