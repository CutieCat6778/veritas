package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormModel struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

// MarshalJSON converts Article to JSON string
func (a *Article) MarshalJSONToString() (string, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalJSONFromString converts JSON string to Article
func UnmarshalJSONFromString(article *Article, jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), article)
}

func (m *GormModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString() // Or your preferred ID generation logic
	}
	return
}
