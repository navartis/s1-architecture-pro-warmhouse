package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Device — запись реестра устройств.
type Device struct {
	ID           string `json:"id"`
	HouseID      string `json:"houseId"`
	TypeID       string `json:"typeId"`
	SerialNumber string `json:"serialNumber"`
	Status       string `json:"status"`
}

// DeviceCreate — данные для регистрации устройства.
type DeviceCreate struct {
	HouseID      string `json:"houseId" binding:"required,uuid"`
	TypeID       string `json:"typeId" binding:"required,uuid"`
	SerialNumber string `json:"serialNumber" binding:"required"`
}

// DeviceUpdate — частичное обновление устройства (пока только состояние).
type DeviceUpdate struct {
	Status string `json:"status" binding:"required,oneof=on off"`
}

// ErrNotFound — устройство не найдено.
var ErrNotFound = errors.New("device not found")

// Store — доступ к базе устройств.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore открывает пул соединений с Device DB.
func NewStore(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Ping проверяет доступность базы.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// List возвращает все устройства.
func (s *Store) List(ctx context.Context) ([]Device, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, house_id::text, type_id::text, serial_number, status
		FROM devices
		ORDER BY serial_number
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]Device, 0)
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.HouseID, &d.TypeID, &d.SerialNumber, &d.Status); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}

	return devices, rows.Err()
}

// Get возвращает устройство по идентификатору.
func (s *Store) Get(ctx context.Context, id string) (Device, error) {
	var d Device
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, house_id::text, type_id::text, serial_number, status
		FROM devices
		WHERE id = $1
	`, id).Scan(&d.ID, &d.HouseID, &d.TypeID, &d.SerialNumber, &d.Status)

	if errors.Is(err, pgx.ErrNoRows) {
		return Device{}, ErrNotFound
	}
	if err != nil {
		return Device{}, err
	}

	return d, nil
}

// Create регистрирует устройство и возвращает созданную запись.
func (s *Store) Create(ctx context.Context, input DeviceCreate) (Device, error) {
	var d Device
	err := s.pool.QueryRow(ctx, `
		INSERT INTO devices (house_id, type_id, serial_number)
		VALUES ($1, $2, $3)
		RETURNING id::text, house_id::text, type_id::text, serial_number, status
	`, input.HouseID, input.TypeID, input.SerialNumber).
		Scan(&d.ID, &d.HouseID, &d.TypeID, &d.SerialNumber, &d.Status)
	if err != nil {
		return Device{}, err
	}

	return d, nil
}

// UpdateStatus меняет состояние устройства.
func (s *Store) UpdateStatus(ctx context.Context, id, status string) (Device, error) {
	var d Device
	err := s.pool.QueryRow(ctx, `
		UPDATE devices
		SET status = $2
		WHERE id = $1
		RETURNING id::text, house_id::text, type_id::text, serial_number, status
	`, id, status).Scan(&d.ID, &d.HouseID, &d.TypeID, &d.SerialNumber, &d.Status)

	if errors.Is(err, pgx.ErrNoRows) {
		return Device{}, ErrNotFound
	}
	if err != nil {
		return Device{}, err
	}

	return d, nil
}
