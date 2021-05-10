package main

import (
	"fmt"
	xerrors "github.com/pkg/errors"
	"task/internal/service"
	"github.com/jinzhu/gorm"
)

func main() {
	_, err := service.NumberLocation.GetLoctaionByNumber(1001)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("ErrRecordNotFound")
		} else {
			fmt.Println("other error")
		}
	}
}