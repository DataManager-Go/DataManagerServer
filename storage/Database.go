package storage

import (
	"fmt"

	"github.com/DataManager-Go/DataManagerServer/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//ConnectToDatabase connects to database
func ConnectToDatabase(config *models.Config) (*gorm.DB, error) {
	sslMode := ""
	if len(config.Server.Database.SSLMode) > 0 {
		sslMode = "sslmode='" + config.Server.Database.SSLMode + "'"
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host='%s' port='%d' user='%s' dbname='%s' password='%s' %s", config.Server.Database.Host, config.Server.Database.DatabasePort, config.Server.Database.Username, config.Server.Database.Database, config.Server.Database.Pass, sslMode)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	//	db = db.Debug()

	//Automigration
	db.AutoMigrate(
		&models.Role{},
		&models.Namespace{},
		&models.Tag{},
		&models.File{},
		&models.Group{},
		&models.User{},
		&models.LoginSession{},
	)

	//Return error if automigration fails
	if err != nil {
		return nil, err
	}

	createRoles(db, config)

	//Create default namespace
	return db, nil
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
