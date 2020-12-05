package storage

import (
	"fmt"

	"github.com/DataManager-Go/DataManagerServer/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//ConnectToDatabase connects to database
func ConnectToDatabase(config *models.Config) (*gorm.DB, error) {
	sslMode := ""
	if len(config.Server.Database.SSLMode) > 0 {
		sslMode = "sslmode='" + config.Server.Database.SSLMode + "'"
	}

	var dialector gorm.Dialector
	if config.Server.Database.Type == "postgres" {
		dialector = postgres.Open(fmt.Sprintf("host='%s' port='%d' user='%s' dbname='%s' password='%s' %s", config.Server.Database.Host, config.Server.Database.DatabasePort, config.Server.Database.Username, config.Server.Database.Database, config.Server.Database.Pass, sslMode))
	} else {
		dbFile := config.Server.Database.Database
		if len(dbFile) == 0 {
			dbFile = "data.db"
		}

		dialector = sqlite.Open(dbFile)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
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
func CheckConnection(db *gorm.DB, config *models.Config) (bool, error) {
	if config.Server.Database.Type == "sqlite" {
		return true, nil
	}

	err := db.Exec("SELECT version();").Error
	return err == nil, err
}
