package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type linkSaver struct {
	gorm.Model
	LinkReal  string `json:"link_real" ,gorm:"link_real"`
	LinkShort string `json:"link_short" ,gorm:"link_short"`
	Name      string `json:"name" ,gorm:"name"`
}

type inputLink struct {
	Name     string `binding:"required" ,json:"name"`
	LinkReal string `json:"link_real"`
}

func main() {
	dbDriver, err := readEnvSupabase()
	if err != nil {
		panic(err.Error())
	}
	db, err := makeConnection(dbDriver)
	if err != nil {
		panic(err.Error())
	}
	if err := db.AutoMigrate(new(linkSaver)); err != nil {
		panic(err.Error())
	}

	domain := "sl.adityaariizkyy.my.id"

	router := gin.Default()

	router.POST("", func(ctx *gin.Context) {
		c := ctx.Request.Context()
		var input inputLink
		if err := ctx.BindJSON(&input); err != nil {
			ctx.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}

		shortLink := fmt.Sprintf("https://%s/%s", domain, input.Name)
		data := linkSaver{
			LinkReal:  input.LinkReal,
			LinkShort: shortLink,
			Name:      input.Name,
		}

		if err := db.WithContext(c).Create(&data).Error; err != nil {
			ctx.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"is_success": true,
			"data":       data,
		})
	})

	router.GET(":linkshort", func(ctx *gin.Context) {
		c := ctx.Request.Context()
		var data linkSaver
		linkShort := ctx.Param("linkshort")
		if err := db.WithContext(c).Where("name = ?", linkShort).Find(&data).Error; err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Redirect(http.StatusPermanentRedirect, data.LinkReal)
	})

	router.Run(":10001")

}

type driverSupabase struct {
	User     string
	Password string
	Host     string
	Port     string
	DbName   string
}

func readEnvSupabase() (driverSupabase, error) {
	envSupabase, err := godotenv.Read()
	if err != nil {
		return driverSupabase{}, err
	}
	return driverSupabase{
		User:     envSupabase["SUPABASE_USER"],
		Password: envSupabase["SUPABASE_PASSWORD"],
		Host:     envSupabase["SUPABASE_HOST"],
		Port:     envSupabase["SUPABASE_PORT"],
		DbName:   envSupabase["SUPABASE_DB_NAME"],
	}, nil
}

func makeConnection(data driverSupabase) (*gorm.DB, error) {
	dsn := fmt.Sprintf("user=%s "+
		"password=%s "+
		"host=%s "+
		"TimeZone=Asia/Singapore "+
		"port=%s "+
		"dbname=%s", data.User, data.Password, data.Host, data.Port, data.DbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(); err != nil {
		return nil, err
	}
	return db, nil
}
