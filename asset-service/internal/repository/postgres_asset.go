package repository

import (
	"context"
	"fmt"
	"gym-membership/asset-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAssetRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresAssetRepository(pool *pgxpool.Pool) domain.AssetRepository {
	return &postgresAssetRepo{
		pool: pool,
	}
}

func (r *postgresAssetRepo) GetByID(id string) (*domain.Asset, error) {
	ctx := context.Background()
	query := `SELECT id, name, type, status, health_score, location, created_at, last_maintained_at FROM assets WHERE id = $1`
	
	var a domain.Asset
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.Name, &a.Type, &a.Status, &a.HealthScore, &a.Location, &a.CreatedAt, &a.LastMaintainedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("asset not found")
		}
		return nil, err
	}
	return &a, nil
}

func (r *postgresAssetRepo) ListByType(assetType string) ([]*domain.Asset, error) {
	ctx := context.Background()
	query := `SELECT id, name, type, status, health_score, location, created_at, last_maintained_at FROM assets`
	var args []interface{}
	
	if assetType != "" {
		query += " WHERE type = $1"
		args = append(args, assetType)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*domain.Asset
	for rows.Next() {
		var a domain.Asset
		err := rows.Scan(
			&a.ID, &a.Name, &a.Type, &a.Status, &a.HealthScore, &a.Location, &a.CreatedAt, &a.LastMaintainedAt,
		)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &a)
	}
	return assets, nil
}

func (r *postgresAssetRepo) UpdateStatus(id string, status string) (*domain.Asset, error) {
	ctx := context.Background()
	
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// SELECT FOR UPDATE to avoid race conditions (Phase 2 requirement)
	querySelect := `SELECT id, name, type, status, health_score, location, created_at, last_maintained_at 
                   FROM assets WHERE id = $1 FOR UPDATE`
	
	var a domain.Asset
	err = tx.QueryRow(ctx, querySelect, id).Scan(
		&a.ID, &a.Name, &a.Type, &a.Status, &a.HealthScore, &a.Location, &a.CreatedAt, &a.LastMaintainedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("asset not found")
		}
		return nil, err
	}

	// Update status
	queryUpdate := `UPDATE assets SET status = $1, last_maintained_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING status, last_maintained_at`
	err = tx.QueryRow(ctx, queryUpdate, status, id).Scan(&a.Status, &a.LastMaintainedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *postgresAssetRepo) CheckAvailability(id string, startTime, endTime string) (bool, error) {
	ctx := context.Background()
	query := `SELECT status FROM assets WHERE id = $1`
	
	var status string
	err := r.pool.QueryRow(ctx, query, id).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, fmt.Errorf("asset not found")
		}
		return false, err
	}
	
	return status == "available", nil
}
func (r *postgresAssetRepo) UpdateHealth(id string, healthDelta int) (*domain.Asset, error) {
	ctx := context.Background()
	
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	querySelect := `SELECT id, name, type, status, health_score, location, created_at, last_maintained_at 
                   FROM assets WHERE id = $1 FOR UPDATE`
	
	var a domain.Asset
	err = tx.QueryRow(ctx, querySelect, id).Scan(
		&a.ID, &a.Name, &a.Type, &a.Status, &a.HealthScore, &a.Location, &a.CreatedAt, &a.LastMaintainedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("asset not found")
		}
		return nil, err
	}

	newHealth := a.HealthScore + healthDelta
	if newHealth < 0 {
		newHealth = 0
	}
	if newHealth > 100 {
		newHealth = 100
	}

	queryUpdate := `UPDATE assets SET health_score = $1 WHERE id = $2 RETURNING health_score`
	err = tx.QueryRow(ctx, queryUpdate, newHealth, id).Scan(&a.HealthScore)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &a, nil
}
