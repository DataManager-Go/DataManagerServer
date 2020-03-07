package storage

import (
	"fmt"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/jinzhu/gorm"
)

//ConnectToDatabase connects to database
func ConnectToDatabase(config *models.Config) (*gorm.DB, error) {
	db, err := gorm.Open("postgres", fmt.Sprintf("host='%s' port='%d' user='%s' dbname='%s' password='%s'", config.Server.Database.Host, config.Server.Database.DatabasePort, config.Server.Database.Username, config.Server.Database.Database, config.Server.Database.Pass))
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(
		&models.Namespace{},
		&models.Tag{},
		&models.File{},
		&models.Group{},
	)

	//Create default namespace
	return db, createDefaultNamespace(db)
}

//CheckConnection return true if connected succesfully
func CheckConnection(db *gorm.DB) (bool, error) {
	err := db.Exec("SELECT version();").Error
	return err == nil, err
}

func createDefaultNamespace(db *gorm.DB) error {
	db.Where("name = ?", "default").Find(&models.DefaultNamespace)
	if models.DefaultNamespace.ID == 0 {
		models.DefaultNamespace = models.Namespace{
			Name: "default",
		}
		return db.Create(&models.DefaultNamespace).Error
	}
	return nil
}
