package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)


const PORT = "8080"

func initSession() *scs.SessionManager{
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime= 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

func initRedis() *redis.Pool{
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func()(redis.Conn, error){
			return redis.Dial("tcp",os.Getenv("REDIS"))
		},
	}
	return redisPool
}


func main(){
	// connect to the database
	db := initDB()
	// create sessions
	session:= initSession()

	//create loggers
	infoLog := log.New(os.Stdout, "INFO\t",log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t",log.Ldate|log.Ltime|log.Lshortfile)

	// create channels

	// create wait group
	wg := sync.WaitGroup{}

	// set up the application config
	app := Config{
		Session: session,
		DB: db,
		Wait: &wg,
		InfoLog: infoLog,
		ErrorLog: errLog,
	}

	// set up mail

	// listen for web connections
	app.serve()

}

func(app *Config) serve(){
	// start http server
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s",PORT),
		Handler: app.routes(),
	}
	app.InfoLog.Println("Starting web server")
	err := srv.ListenAndServe()
	if err !=nil{
		log.Panic(err)
	}
}

func initDB()*sql.DB {
	conn := connectToDB()
	if conn == nil{
		log.Panic("Can't connect to the database")
	}
	return conn
}

func connectToDB() *sql.DB{
	counts := 0

	dsn := os.Getenv("DSN")
	for {
		connection, err := openDB(dsn)
		if err!=nil{
			log.Println("Postgres not ready.")
		}else{
			log.Println("Connected to DB")
			return connection
		}
		if counts > 10{
			return nil
		}
		counts++
		log.Panicln("Backing off for 1 sec")
		time.Sleep(1 * time.Second)
		continue
	}

}

func openDB(dsn string)(*sql.DB, error){
	db, err := sql.Open("pgx",dsn)
	if err !=nil{
		return nil, err
	}

	err = db.Ping()

	if err !=nil{
		return nil , err
	}

	return db, nil
}


