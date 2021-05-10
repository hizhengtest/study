package database

import (
	"fmt"
	"log"
	"task/config"
	numberModel "task/internal/models/number"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //mysql
)

//DB mysql db Instance
var DB *gorm.DB

func init() {
	connectDB()
}

func connectDB() {
	var (
		err                                  error
		dbType, dbName, user, password, host string
	)

	dbType = config.Default.Database.Type
	dbName = config.Default.Database.Name
	user = config.Default.Database.User
	password = config.Default.Database.Password
	host = config.Default.Database.Host

	DB, err = gorm.Open(dbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName))

	if err != nil {
		log.Println(err)
	} else {
		log.Print("Mysql连接成功")
	}

	DB.DB().SetMaxIdleConns(config.Default.Database.PoolIdleNum)
	DB.DB().SetMaxOpenConns(config.Default.Database.PoolOpenNum)

	// init table

	if DB.HasTable(&numberModel.Location{}) {
		DB.AutoMigrate(&numberModel.Location{})
	} else {
		fmt.Println("create table:number_locations")
		DB.CreateTable(&numberModel.Location{}) //AddUniqueIndex("idx_pack_id", "pack_id").AddIndex("idx_user_id", "user_id")
	}

	if config.Basic.App.Debug {
		DB.LogMode(true)
	}
}

// CloseDB close db
func CloseDB() {
	if err := DB.Close(); nil != err {
		fmt.Println("Disconnect from database failed: " + err.Error())
	}
}
