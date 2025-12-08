package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *SysAPIResource) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.AddClause(
		clause.OnConflict{
			Columns: []clause.Column{
				clause.Column{
					Name: "id",
				},
			},
			DoUpdates: []clause.Assignment{
				{
					Column: clause.Column{
						Name: "deleted_at",
					},
					Value: gorm.DeletedAt{Valid: false},
				},
				{
					Column: clause.Column{
						Name: "deleted_by",
					},
				},
			},
		},
	)
	return nil
}
