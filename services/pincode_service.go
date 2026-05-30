package services

import (
	"errors"
	"strconv"
	"sync"

	"github.com/19481A1281/go-pincode-service/models"
	"github.com/19481A1281/go-pincode-service/repositories"
	"github.com/19481A1281/go-pincode-service/excel"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type PincodeService interface {
	Create(*models.Pincode) (*models.Pincode, error)
	//BulkCreate([]models.Pincode)error
	Update(pincode uint32, updates map[string]interface{}) (*models.Pincode, error)
	Delete(pincode uint32) error
	GetAll(page int, limit int) (*[]models.Pincode, error)
	GetByID(id uint16) (*models.Pincode, error)
	GetByPincode(pincode uint32) (*models.Pincode, error)
}

type pincodeService struct {
	repo    repositories.PincodeRepository
	cache   sync.Map
	sfGroup singleflight.Group
}

func NewPincodeService(repo repositories.PincodeRepository) PincodeService {
	return &pincodeService{
		repo: repo,
	}
}

func (s *pincodeService) Create(pincode *models.Pincode) (*models.Pincode, error) {
	data, err := s.repo.Create(pincode)
	if err != nil {
		return nil, err
	}
	s.cache.Store(data.Pincode, data)
	s.cache.Store(data.ID, data)
	return data, nil
}

func (s *pincodeService) Update(pincode uint32, updates map[string]interface{}) (*models.Pincode, error) {
	data, err := s.repo.Update(pincode, updates)
	if err != nil {
		return nil, err
	}
	s.cache.Delete(pincode)
	s.cache.Delete(data.ID)
	s.cache.Store(pincode, data)
	s.cache.Store(data.ID, data)
	return data, nil
}

func (s *pincodeService) Delete(pincode uint32) error {
	if val, ok := s.cache.Load(pincode); ok {
		if cached, ok := val.(*models.Pincode); ok {
			s.cache.Delete(cached.ID)
		}
	}
	s.cache.Delete(pincode)
	return s.repo.Delete(pincode)
}

func (s *pincodeService) GetAll(page int, limit int) (*[]models.Pincode, error) {
	return s.repo.GetAll(page, limit)
}

func (s *pincodeService) GetByID(id uint16) (*models.Pincode, error) {
	if val, ok := s.cache.Load(id); ok {
		if cached, ok := val.(*models.Pincode); ok {
			return cached, nil
		}
	}

	key := "id:" + strconv.FormatUint(uint64(id), 10)
	v, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		if val, ok := s.cache.Load(id); ok {
			if cached, ok := val.(*models.Pincode); ok {
				return cached, nil
			}
		}

		data, err := s.repo.GetByID(id)
		if err != nil {
			return nil, err
		}

		s.cache.Store(id, data)
		s.cache.Store(data.Pincode, data)
		return data, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*models.Pincode), nil
}

func (s *pincodeService) GetByPincode(pincode uint32) (*models.Pincode, error) {
	if val, ok := s.cache.Load(pincode); ok {
		if cached, ok := val.(*models.Pincode); ok {
			return cached, nil
		}
	}

	key := strconv.FormatUint(uint64(pincode), 10)
	v, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		if val, ok := s.cache.Load(pincode); ok {
			if cached, ok := val.(*models.Pincode); ok {
				return cached, nil
			}
		}

		data, err := s.repo.GetByPincode(pincode)
		if err == nil {
			s.cache.Store(pincode, data)
			s.cache.Store(data.ID, data)
			return data, nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		excelData, err := excel.SearchPincodeInExcel(pincode)
		if err != nil {
			return nil, err
		}

		newPincode, err := s.repo.Create(excelData)
		if err != nil {
			return nil, err
		}

		s.cache.Store(pincode, newPincode)
		s.cache.Store(newPincode.ID, newPincode)
		return newPincode, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*models.Pincode), nil
}


