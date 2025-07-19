package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/lavatee/astraltest/internal/model"
)

type DocumentsPostgres struct {
	db *sqlx.DB
}

func NewDocumentsPostgres(db *sqlx.DB) *DocumentsPostgres {
	return &DocumentsPostgres{db: db}
}

func (r *DocumentsPostgres) Create(ctx context.Context, token string, doc *model.Document, jsonData string, fileData []byte) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `INSERT INTO documents (id, name, mime, is_file, is_public, created_at, owner_id)
	VALUES ($1, $2, $3, $4, $5, $6, 
	(SELECT user_id FROM sessions WHERE token = $7))`
	_, err = tx.ExecContext(ctx, query,
		doc.ID, doc.Name, doc.Mime, doc.File, doc.Public, doc.Created, token)
	if err != nil {
		return err
	}
	query = `INSERT INTO document_files (document_id, data) VALUES ($1, $2)`
	if doc.File {
		_, err = tx.ExecContext(ctx, query, doc.ID, fileData)
	} else {
		query = `INSERT INTO document_data (document_id, data) VALUES ($1, $2)`
		_, err = tx.ExecContext(ctx, query, doc.ID, jsonData)
	}
	if err != nil {
		return err
	}
	if len(doc.Grant) > 0 {
		for _, login := range doc.Grant {
			query = `INSERT INTO document_grants (document_id, user_id)
			VALUES ($1, (SELECT user_id FROM users WHERE login = $2))`
			_, err = tx.ExecContext(ctx, query, doc.ID, login)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *DocumentsPostgres) GetAll(ctx context.Context, token, login, key, value string, limit int) ([]*model.Document, error) {
	query := `SELECT d.id, d.name, d.mime, d.is_file, d.is_public, d.created_at
	FROM documents d
	JOIN sessions s ON s.token = $1
	WHERE (d.owner_id = s.user_id OR d.is_public = true OR EXISTS (
	SELECT 1 FROM document_grants g 
	WHERE g.document_id = d.id AND g.user_id = s.user_id
	))`
	params := []interface{}{token}
	if login != "" {
		query += ` AND d.owner_id = (SELECT user_id FROM users WHERE login = $2)`
		params = append(params, login)
	}
	if key != "" && value != "" {
		query += ` AND d.` + key + ` = $` + strconv.Itoa(len(params)+1)
		params = append(params, value)
	}
	query += ` ORDER BY d.name, d.created_at`
	if limit > 0 {
		query += ` LIMIT $` + strconv.Itoa(len(params)+1)
		params = append(params, limit)
	}
	var docs []*model.Document
	err := r.db.SelectContext(ctx, &docs, query, params...)
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		var grants []string
		err := r.db.SelectContext(ctx, &grants, `
			SELECT u.login 
			FROM document_grants g
			JOIN users u ON u.user_id = g.user_id
			WHERE g.document_id = $1`, doc.ID)
		if err != nil {
			return nil, err
		}
		doc.Grant = grants
	}
	return docs, nil
}

func (r *DocumentsPostgres) GetByID(ctx context.Context, token, id string) (*model.Document, []byte, error) {
	var hasAccess bool
	query := `SELECT EXISTS(
	SELECT 1 FROM documents d
	JOIN sessions s ON s.token = $1
	WHERE d.id = $2 AND (d.owner_id = s.user_id OR d.is_public = true OR EXISTS (
	SELECT 1 FROM document_grants g 
	WHERE g.document_id = d.id AND g.user_id = s.user_id
	))
	)`
	err := r.db.GetContext(ctx, &hasAccess, query, token, id)
	if err != nil || !hasAccess {
		return nil, nil, errors.New("User doesn't have access to this file")
	}
	var doc model.Document
	query = `SELECT id, name, mime, is_file, is_public, created_at
	FROM documents WHERE id = $1`
	err = r.db.GetContext(ctx, &doc, query, id)
	if err != nil {
		return nil, nil, errors.New("Document has not been found")
	}
	var grants []string
	query = `SELECT u.login 
	FROM document_grants g
	JOIN users u ON u.user_id = g.user_id
	WHERE g.document_id = $1`
	err = r.db.SelectContext(ctx, &grants, query, id)
	if err != nil {
		return nil, nil, errors.New("Grants have not been found")
	}
	doc.Grant = grants
	var data []byte
	if doc.File {
		query = `SELECT data FROM document_files WHERE document_id = $1`
		err = r.db.GetContext(ctx, &data, query, id)
	} else {
		query = `SELECT data FROM document_data WHERE document_id = $1`
		err = r.db.GetContext(ctx, &data, query, id)
	}
	if err != nil {
		return nil, nil, err
	}
	return &doc, data, nil
}

func (r *DocumentsPostgres) GetFileData(ctx context.Context, token, id string) ([]byte, error) {
	var doc model.Document
	query := `SELECT id, name, mime, is_file, is_public, created_at
	FROM documents WHERE id = $1`
	err := r.db.GetContext(ctx, &doc, query, id)
	if err != nil {
		return nil, errors.New("Document has not been found")
	}
	var data []byte
	if doc.File {
		query = `SELECT data FROM document_files WHERE document_id = $1`
		err = r.db.GetContext(ctx, &data, query, id)
	} else {
		query = `SELECT data FROM document_data WHERE document_id = $1`
		err = r.db.GetContext(ctx, &data, query, id)
	}
	if err != nil {
		err = errors.New("File data has not been found")
	}
	return data, err
}

func (r *DocumentsPostgres) Delete(ctx context.Context, token, id string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var isOwner bool
	query := `SELECT EXISTS(
	SELECT 1 FROM documents d
	JOIN sessions s ON s.token = $1
	WHERE d.id = $2 AND d.owner_id = s.user_id
	)`
	err = tx.GetContext(ctx, &isOwner, query, token, id)
	if err != nil || !isOwner {
		return sql.ErrNoRows
	}
	query = `DELETE FROM document_grants WHERE document_id = $1`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	query = `DELETE FROM document_files WHERE document_id = $1`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	query = `DELETE FROM document_data WHERE document_id = $1`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	query = `DELETE FROM documents WHERE id = $1`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
