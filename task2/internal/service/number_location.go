package service

import (
	xerrors "github.com/pkg/errors"
	"sync"
	"task/internal/database"
	models "task/internal/models/number"

	"github.com/jinzhu/gorm"
)

// NumberLocation service.
var NumberLocation = &numberLocationService{
	baseService{
		mutex: &sync.Mutex{},
	},
}

type numberLocationService struct {
	baseService
}

//GetLoctaionByNumber 根据号码获取信息
func (s *numberLocationService) GetLoctaionByNumber(number int) (models.Location, error) {
	var m models.Location
	if err := s.GetDB().Where("`number` = ?", number).Find(&m).Error; nil != err {
		return m, xerrors.Wrap(err, "service find data error")
	}
	return m, nil
}

// 获取数据库
func (s *numberLocationService) GetDB() *gorm.DB {
	return database.DB
}
