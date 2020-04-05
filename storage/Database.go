package storage

import (
	"fmt"

	"github.com/DataManager-Go/DataManagerServer/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

//ConnectToDatabase connects to database
func ConnectToDatabase(config *models.Config) (*gorm.DB, error) {
	sslMode := ""
	if len(config.Server.Database.SSLMode) > 0 {
		sslMode = "sslmode='" + config.Server.Database.SSLMode + "'"
	}

	db, err := gorm.Open("postgres", fmt.Sprintf("host='%s' port='%d' user='%s' dbname='%s' password='%s' %s", config.Server.Database.Host, config.Server.Database.DatabasePort, config.Server.Database.Username, config.Server.Database.Database, config.Server.Database.Pass, sslMode))
	if err != nil {
		return nil, err
	}

	//Automigration
	err = db.AutoMigrate(
		&models.Role{},
		&models.Namespace{},
		&models.Tag{},
		&models.File{},
		&models.Group{},
		&models.User{},
		&models.LoginSession{},
	).Error

	//Return error if automigration fails
	if err != nil {
		return nil, err
	}

	createRoles(db, config)

	//Create default namespace
	return db, createDefaultNamespace(db)
}

func createRoles(db *gorm.DB, config *models.Config) {
	//Create in config specified roles
	for _, role := range config.Server.Roles.Roles {
		err := db.FirstOrCreate(&role).Error
		if err != nil {
			log.Fatalln(err)
		}
	}
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
