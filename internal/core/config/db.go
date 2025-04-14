package config

import (
	"log"

	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

var (
	db *gorm.DB
)

const (
	ReadDBKey  = "read"
	WriteDBKey = "write"
)

func GetDB() *gorm.DB {
	if db == nil {
		db = connectDatabase()
	}
	return db
}

func connectDatabase() *gorm.DB {
	envHasLog := GetEnvVariable("DB_HAS_LOG", "false")
	shouldLog, _ := strconv.ParseBool(envHasLog)
	logLevel := logger.Silent
	if shouldLog {
		logLevel = logger.Info
	}
	connString := GetConnectionString(WriteDBKey)
	var err error
	db, err = gorm.Open(mysql.Open(connString), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		logrus.Errorf("Error in connecting to db: %s", err.Error())
		panic("Failed to connect to database!")
	}

	connROString := GetConnectionString(ReadDBKey)
	db.Use(dbresolver.Register(dbresolver.Config{
		Replicas:          []gorm.Dialector{mysql.Open(connROString)},
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}))

	rawDB, err := db.DB()
	if err != nil {
		panic("Failed to get SQL DB")
	}

	envMaxIdleConn := GetEnvVariable("DB_MAX_IDLE_CONN", "10")
	maxIdleConn, _ := strconv.Atoi(envMaxIdleConn)

	envMaxOpenConn := GetEnvVariable("DB_MAX_OPEN_CONN", "50")
	maxOpenConn, _ := strconv.Atoi(envMaxOpenConn)

	envMaxLifeConn := GetEnvVariable("DB_MAX_LIFE_CONN", "1h")
	maxLifeConn, _ := time.ParseDuration(envMaxLifeConn)

	rawDB.SetMaxIdleConns(maxIdleConn)
	rawDB.SetMaxOpenConns(maxOpenConn)
	rawDB.SetConnMaxLifetime(maxLifeConn)

	log.Println("Database connection established")
	return db
}

func GetConnectionString(key string) string {
	user, pass, host, port, schema := getEnvVarFromDbKey(key)
	var connString strings.Builder
	connString.WriteString(user)
	connString.WriteString(":")
	connString.WriteString(pass)
	connString.WriteString("@tcp(")
	connString.WriteString(host)
	connString.WriteString(":")
	connString.WriteString(port)
	connString.WriteString(")/")
	connString.WriteString(schema)
	connString.WriteString("?charset=utf8")
	connString.WriteString("&parseTime=True")
	connString.WriteString("&loc=Local")
	return connString.String()
}

func getEnvVarFromDbKey(key string) (user, pass, host, port, schema string) {
	if key == WriteDBKey {
		return GetEnvVariable("DB_USER", "root"), GetEnvVariable("DB_PASSWORD", "password"), GetEnvVariable("DB_HOST", "localhost"), GetEnvVariable("DB_PORT", "3306"), GetEnvVariable("DB_NAME", "coopdb")
	} else {
		return GetEnvVariable("DB_RO_USER", "root"), GetEnvVariable("DB_RO_PASSWORD", "password"), GetEnvVariable("DB_RO_HOST", "host.docker.internal"), GetEnvVariable("DB_RO_PORT", "3306"), GetEnvVariable("DB_RO_NAME", "coopdb")
	}
}

func CloseDB() {
	if db != nil {
		rawDB, err := db.DB()
		if err != nil {
			panic("error when getting sql DB")
		}

		log.Println("Closing database connection")
		defer rawDB.Close()
	}
}
