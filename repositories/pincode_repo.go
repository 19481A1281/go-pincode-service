package repositories

import (
	"github.com/19481A1281/go-pincode-service/models"
	"gorm.io/gorm"
)

type PincodeRepository interface {
	Create(*models.Pincode) (*models.Pincode, error)
	//BulkCreate([]models.Pincode)error
	Update(id uint32, updates map[string]interface{}) (*models.Pincode, error)
	Delete(id uint16) error
	GetAll(page int, limit int) (*[]models.Pincode, error)
	GetByID(id uint16) (*models.Pincode, error)
	GetByPincode(pin uint32)(*models.Pincode, error)
}

type pincodeRepository struct {
	db *gorm.DB
}

func NewPincodeRepository(db *gorm.DB) PincodeRepository {
	return &pincodeRepository{
		db: db,
	}
}

func (r *pincodeRepository) Create(pincode *models.Pincode) (*models.Pincode, error) {
	err := r.db.Create(pincode).Error

	if err != nil {
		return nil, err
	}

	return pincode, nil
}

func (r *pincodeRepository) GetAll(page int, limit int) (*[]models.Pincode, error) {
	var pincodes []models.Pincode

	offset := (page - 1) * limit

	err := r.db.Limit(limit).Offset(offset).Find(&pincodes).Error

	if err != nil {
		return nil, err
	}

	return &pincodes, nil
}

func (r *pincodeRepository) GetByID(id uint16) (*models.Pincode, error) {
	var pincode models.Pincode

	err := r.db.First(&pincode, id).Error
	if err != nil {
		return nil, err
	}

	return &pincode, nil
}

func (r *pincodeRepository) Update(pin uint32, updates map[string]interface{}) (*models.Pincode, error) {
	var pincode models.Pincode

	result := r.db.Model(&pincode).Where("pincode = ?", pin).Updates(updates)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	err := r.db.Where("pincode = ?",pin).First(&pincode).Error
	if err != nil {
		return nil, err
	}

	return &pincode, nil
}

func (r *pincodeRepository) Delete(id uint16) error {
	result := r.db.Delete(&models.Pincode{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *pincodeRepository) GetByPincode(pincode uint32)(*models.Pincode,error){
	var data models.Pincode

	err := r.db.Where("pincode = ?",pincode).First(&data).Error

	if err!=nil{
		return nil, err
	}

	return &data,nil
}
