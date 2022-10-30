package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-service/internal/user/domain"
)

func NewUserAdapter(db *sql.DB) *UserAdapter {
	return &UserAdapter{DB: db}
}

type UserAdapter struct {
	DB *sql.DB
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
			&user.Phone,
			&user.Email,
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
	res, err := stmt.ExecContext(ctx, user.Username, user.Email, user.Phone, user.DateOfBirth, user.Id)
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
