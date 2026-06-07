package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func findOne[T any](
	ctx context.Context,
	db *gorm.DB,
	condition string,
	args ...interface{}) (*T, error) {

	var result T
	err := db.WithContext(ctx).
		Where(condition, args...).
		First(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("findOne[%T]: %w", result, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("findOne[%T] %s: %w",
			result, condition, err)
	}
	return &result, nil
}

func findAll[T any](
	ctx context.Context,
	db *gorm.DB,
	condition string,
	args ...interface{}) ([]*T, error) {

	var results []*T
	query := db.WithContext(ctx)
	if condition != "" {
		query = query.Where(condition, args...)
	}
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("findAll[%T] %s: %w",
			*new(T), condition, err)
	}
	return results, nil
}
