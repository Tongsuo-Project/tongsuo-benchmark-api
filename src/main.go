package main

import (
	"fmt"
	"flag"
	"strings"
	"net/http"
	"io/ioutil"
	_ "encoding/json"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	DB *sqlx.DB
	dbDefaultPort int = 3306
	webDefaultPort int = 9999
)

type SymmetricEncryptionRecord struct {
	Id				int64 `db:"id"`
	LastCommit		string `db:"last_commit"`
	MyCommit		string `db:"my_commit"`
	Algorithm		string `db:"algorithm"`
	Date			string `db:"date"`
	JobDate			string `db:"job_date"`
	Bytes16			int64 `db:"bytes16"`
	Bytes64			int64 `db:"bytes64"`
	Bytes256		int64 `db:"bytes256"`
	Bytes1024		int64 `db:"bytes1024"`
	Bytes8192		int64 `db:"bytes8192"`
}

type DigestRecord struct {
	Id				int64 `db:"id"`
	LastCommit		string `db:"last_commit"`
	MyCommit		string `db:"my_commit"`
	Algorithm		string `db:"algorithm"`
	Date			string `db:"date"`
	JobDate			string `db:"job_date"`
	Bytes16			int64 `db:"bytes16"`
	Bytes64			int64 `db:"bytes64"`
	Bytes256		int64 `db:"bytes256"`
	Bytes1024		int64 `db:"bytes1024"`
	Bytes8192		int64 `db:"bytes8192"`
	Bytes16384		int64 `db:"bytes16384"`
}

type KeyExchangeRecord struct {
	Id				int64 `db:"id"`
	LastCommit		string `db:"last_commit"`
	MyCommit		string `db:"my_commit"`
	Algorithm		string `db:"algorithm"`
	Date			string `db:"date"`
	JobDate			string `db:"job_date"`
	OpTime			float32 `db:"op_time"`
	OpQPS			float32 `db:"op_qps"`
}

type PheRecord struct {
	Id				int64 `db:"id"`
	LastCommit		string `db:"last_commit"`
	MyCommit		string `db:"my_commit"`
	Algorithm		string `db:"algorithm"`
	Date			string `db:"date"`
	JobDate			string `db:"job_date"`
	A				int32 `db:"a"`
	B				int32 `db:"b"`
	EncryptQPS		float32 `db:"encrypt_qps"`
	DecryptQPS		float32 `db:"decrypt_qps"`
	AddQPS			float32 `db:"add_qps"`
	SubQPS			float32 `db:"sub_qps"`
	ScalarMulQPS	float32 `db:"scalar_mul_qps"`
}

type SignatureRecord struct {
	Id				int64 `db:"id"`
	LastCommit		string `db:"last_commit"`
	MyCommit		string `db:"my_commit"`
	Algorithm		string `db:"algorithm"`
	Date			string `db:"date"`
	JobDate			string `db:"job_date"`
	SignTime		float32 `db:"sign_time"`
	VerifyTime		float32 `db:"verify_time"`
	SignQPS			float32 `db:"sign_qps"`
	VerifyQPS		float32 `db:"verify_qps"`
}

type Web struct {
	Port		int			`yaml:"port"`
}

type DbConfig struct {
	Username	string    `yaml:"username"`
	Password	string    `yaml:"password"`
	Addr		string    `yaml:"addr"`
	Port		int		  `yaml:"port"`
	Database	string    `yaml:"database"`
}

type Config struct {
	Db *DbConfig `yaml:"db"`
	Web *Web	 `yaml:"web"`
}

func initDB(dbConf *DbConfig) (db *sqlx.DB, err error) {
	var port = dbDefaultPort
	if dbConf.Port != 0 {
		port = dbConf.Port
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True", dbConf.Username, dbConf.Password, dbConf.Addr, port, dbConf.Database)
	db, err = sqlx.Connect("mysql", dsn)
    if err != nil {
        fmt.Printf("connect DB failed, err:%v\n", err)
        return nil, err
    }
    db.SetMaxOpenConns(20)
    db.SetMaxIdleConns(10)
    return db, nil
}

func Cors() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
        }

        if c.GetHeader("Origin") != "" {
            c.Header("Access-Control-Allow-Origin", "*")
            c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
            c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
            c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
            c.Header("Access-Control-Allow-Credentials", "true")
        }

        c.Next()
    }
}

func SearchTime() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method != "GET" {
            c.AbortWithStatus(http.StatusNoContent)
        }

		startTime := c.Query("start_time")
		endTime := c.Query("end_time")

		if startTime != "" || endTime != "" {
			var condition []string

			if startTime != "" {
				condition = append(condition, fmt.Sprintf("`job_date` >= '%s'", startTime))
			}

			if endTime != "" {
				condition = append(condition, fmt.Sprintf("`job_date` <= '%s'", endTime))
			}

			whereSql := fmt.Sprintf("where %s", strings.Join(condition, " and "))

			c.Set("where", whereSql)
		}

        c.Next()
    }
}

func OrderBy() gin.HandlerFunc {
    return func(c *gin.Context) {
		c.Set("order", "order by id asc")
        c.Next()
    }
}

func main() {
	configPath := flag.String("config", "./config.yaml", "input config file path")
	flag.Parse()

	yamlFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
        panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	var config *Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
        panic(fmt.Errorf("unmarshal conf failed, err:%s \n", err))
	}

	DB, err := initDB(config.Db)
	defer DB.Close()

    r := gin.Default()
	r.Use(Cors())
	r.Use(SearchTime())
	r.Use(OrderBy())

	v1 := r.Group("/v1")
	{
		v1.GET("/:algo", func(c *gin.Context) {
			var records interface{}
			algo := c.Param("algo")
			switch (algo) {
			case "symmetric_encryption":
				records = new([]SymmetricEncryptionRecord)
			case "digest":
				records = new([]DigestRecord)
			case "key_exchange":
				records = new([]KeyExchangeRecord)
			case "signature":
				records = new([]SignatureRecord)
			case "phe":
				records = new([]PheRecord)
			default:
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			sql := fmt.Sprintf("select * from %s %s %s", algo, c.GetString("where"), c.GetString("order"))
			err = DB.Select(records, sql)
			if err != nil {
				fmt.Println("exec failed, ", err)
				return
			}

			c.JSON(http.StatusOK, records)
		})
	}

	var port string
	if config.Web != nil && config.Web.Port != 0 {
		port = fmt.Sprintf(":%d", config.Web.Port)
	} else {
		port = fmt.Sprintf(":%d", webDefaultPort)
	}

	r.Run(port)
}
