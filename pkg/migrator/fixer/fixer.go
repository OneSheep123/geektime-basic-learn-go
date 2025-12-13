package fixer

import (
	"context"
	"ddd_demo/pkg/migrator"
	"ddd_demo/pkg/migrator/events"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	columns []string
}

func NewOverrideFixerV1[T migrator.Entity](base *gorm.DB, target *gorm.DB,
	columns []string) *OverrideFixer[T] {
	return &OverrideFixer[T]{base: base, target: target, columns: columns}
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	rows, err := base.Model(new(T)).Order("id").Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	return &OverrideFixer[T]{base: base, target: target, columns: columns}, err
}

func (f *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	// 最最粗暴的
	var t T
	err := f.base.WithContext(ctx).Where("id=?", id).First(&t).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return f.target.WithContext(ctx).Model(&t).Delete("id = ?", id).Error
	case nil:
		// upsert: 这里使用upsert而不判断update/insert
		// 是因为后续开启双写之后，也要校验和修复，那么显然在那个时候源表和目标表都会被修改。
		// 此时会出现在双写阶段，你校验的时候 target 没有这条数据，但是紧接着你修复的时候，就有这个数据了
		// 此时如果进行insert的话，若没有唯一索引，就会出现重复数据问题/报错(并发问题)。因此这里直接upsert
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	default:
		return err
	}
}

func (f *OverrideFixer[T]) FixV1(evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNEQ, events.InconsistentEventTypeTargetMissing:
		var t T
		err := f.base.Where("id=?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			return f.target.Model(&t).Delete("id = ?", evt.ID).Error
		case nil:
			// upsert
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.Model(new(T)).Delete("id = ?", evt.ID).Error
	}
	return nil
}
