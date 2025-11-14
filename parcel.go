package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add добавляет новую посылку со статусом "registered"
func (s ParcelStore) Add(p Parcel) (int, error) {
	const query = `
		INSERT INTO parcel (client, status, address, created_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, p.Client, ParcelStatusRegistered, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("ошибка при добавлении посылки: %w", err)
	}

	var id int
	err = s.db.QueryRow(`SELECT last_insert_rowid()`).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID: %w", err)
	}

	return id, nil
}

// Get возвращает посылку по номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	const query = `
		SELECT number, client, status, address, created_at
		FROM parcel
		WHERE number = ?
	`
	var p Parcel
	err := s.db.QueryRow(query, number).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, fmt.Errorf("посылка %d не найдена", number)
		}
		return p, fmt.Errorf("ошибка БД: %w", err)
	}

	return p, nil
}

// GetByClient возвращает все посылки клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	const query = `
		SELECT number, client, status, address, created_at
		FROM parcel
		WHERE client = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе по клиенту: %w", err)
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %w", err)
		}
		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации: %w", err)
	}

	return res, nil
}

// SetStatus обновляет статус посылки
func (s ParcelStore) SetStatus(number int, status string) error {
	const query = `
		UPDATE parcel
		SET status = ?
		WHERE number = ?
	`
	res, err := s.db.Exec(query, status, number)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("посылка %d не найдена", number)
	}

	return nil
}

// SetAddress обновляет адрес, только если статус "registered"
func (s ParcelStore) SetAddress(number int, address string) error {
	const query = `
		UPDATE parcel
		SET address = ?
		WHERE number = ? AND status = ?
	`
	res, err := s.db.Exec(query, address, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("ошибка обновления адреса: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("адрес не может быть изменён: посылка %d либо не найдена, либо не в статусе registered", number)
	}

	return nil
}

// Delete удаляет посылку, только если статус "registered"
func (s ParcelStore) Delete(number int) error {
	const query = `
		DELETE FROM parcel
		WHERE number = ? AND status = ?
	`
	res, err := s.db.Exec(query, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("ошибка удаления посылки: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("посылку %d нельзя удалить: либо не найдена, либо не в статусе registered", number)
	}

	return nil
}
