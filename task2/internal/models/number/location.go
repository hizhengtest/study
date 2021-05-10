package models

// Location model.
type Location struct {
	ID         int    `gorm:"primary_key"`
	Type       uint   `gorm:"size:1;default:0;" json:"type"`          // 运营商
	Number     int    `gorm:"size:11;index;default:0;" json:"number"` // 号段
	Province   string `gorm:"size:50" json:"province"`                // 省份
	City       string `gorm:"size:50" json:"city"`                    // 城市
	AreaNumber string `gorm:"size:100;" json:"area_number"`           // 区号
	Zipcode    int    `gorm:"size:11;default:0;" json:"zipcode"`      // 邮编
	//DateCreate time.Time `gorm:"" json:"date_create"`                    // 新增时间
}

//TableName 获取表名
func (Location) TableName() string {
	return "number_locations"
}
