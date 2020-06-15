package main

import (
	"database/sql"
	"fmt"
	cache "github.com/Nuclear-Catapult/Youtube-Crawler/ID-Cache"
	b64 "github.com/Nuclear-Catapult/Youtube-Crawler/ytbase64"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

var thread_count int = 200

func main() {
	c := make(chan []interface{})

	load_db()
	go inserter(c)
	for i := 0; i < thread_count; i++ {
		go crawler(c)
		for cache.QueueCount() < 5 {
			time.Sleep(time.Millisecond * 100)
		}
	}

	var input string
	for {
		fmt.Println("[s]tatus, [q]uite, [a]dd thread: followed by enter")
		fmt.Scan(&input)
		switch input {
		case "a":
			go crawler(c)
			thread_count++
			fmt.Printf("Thread count: %d\n", thread_count)
		case "s":
			cache.Status()
			fmt.Printf("Threads: %d\n", thread_count)
		case "q":
			return
		default:
			fmt.Println("Invalid input")
		}
	}
}

func load_db() {
	db, err := sql.Open("sqlite3", "./yt-videos.db")
	table_stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS video (
	video_id INTEGER(64) PRIMARY KEY,
	title VARCHAR(100) NOT NULL,
	views INTEGER(64) NOT NULL,
	likes INTEGER(64) NOT NULL,
	dislikes INTEGER(64) NOT NULL,
	rec_1 INTEGER(64) NOT NULL,
	rec_2 INTEGER(64) NOT NULL,
	rec_3 INTEGER(64) NOT NULL,
	rec_4 INTEGER(64) NOT NULL,
	rec_5 INTEGER(64) NOT NULL,
	rec_6 INTEGER(64) NOT NULL,
	rec_7 INTEGER(64) NOT NULL,
	rec_8 INTEGER(64) NOT NULL,
	rec_9 INTEGER(64) NOT NULL,
	rec_10 INTEGER(64) NOT NULL,
	rec_11 INTEGER(64) NOT NULL,
	rec_12 INTEGER(64) NOT NULL,
	rec_13 INTEGER(64) NOT NULL,
	rec_14 INTEGER(64) NOT NULL,
	rec_15 INTEGER(64) NOT NULL,
	rec_16 INTEGER(64) NOT NULL,
	rec_17 INTEGER(64) NOT NULL,
	rec_18 INTEGER(64) NOT NULL);`)
	checkErr(err)
	table_stmt.Exec()

	var row_count int64
	rows, _ := db.Query("SELECT COUNT(*) FROM video")
	for rows.Next() {
		rows.Scan(&row_count)
	}

	if row_count == 0 {
		fmt.Println("yt-video.db not found. Loading seed ID hsWr_JWTZss")
		seed_ID := "hsWr_JWTZss"
		cache.Insert(b64.Decode64(seed_ID))
		return
	}
	fmt.Println("Loading yt-video.db...")

	var video_id int64
	rows, _ = db.Query("SELECT video_id FROM video")
	for rows.Next() {
		rows.Scan(&video_id)
		cache.Key_Insert(video_id)
	}

	if cache.QueueCount() != 0 {
		log.Fatal("Fatal: Queue count is not zero")
	}

	var rec [18]int64
	rows2, _ := db.Query(`SELECT rec_1, rec_2, rec_3, rec_4, rec_5, rec_6, rec_7, rec_8, rec_9,
	                    rec_10, rec_11, rec_12, rec_13, rec_14, rec_15, rec_16, rec_17, rec_18 FROM video`)
	for rows2.Next() {
		rows2.Scan(&rec[0], &rec[1], &rec[2], &rec[3], &rec[4], &rec[5], &rec[6], &rec[7], &rec[8],
			&rec[9], &rec[10], &rec[11], &rec[12], &rec[13], &rec[14], &rec[15], &rec[16], &rec[17])
		for i := 0; i < 18; i++ {
			cache.Insert(rec[i])
		}
	}
	cache.Status()
	db.Close()
}

func inserter(c chan []interface{}) {
	db, err := sql.Open("sqlite3", "./yt-videos.db?_sync=0")
	stmt, err := db.Prepare(`INSERT INTO video
	(video_id, title, views, likes, dislikes, rec_1, rec_2, rec_3, rec_4, rec_5, rec_6, rec_7, rec_8, rec_9,
	rec_10, rec_11, rec_12, rec_13, rec_14, rec_15, rec_16, rec_17, rec_18)
	values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	checkErr(err)

	for {
		_, err := stmt.Exec(<-c...)
		checkErr(err)
	}
	db.Close()
}

func crawler(c chan []interface{}) {
	for id := cache.Next(); id != 0; id = cache.Next() {
		doc, err := goquery.NewDocument("https://www.youtube.com/watch?v=" + b64.Encode64(id))
		checkErr(err)
		title := doc.Find("title").Text()
		if len(title) > 7 {
			ParseHTML(doc, id, title, c)
		} else {
			cache.TryAgainLater(id)
		}
	}
	fmt.Println("Stack empty. Thread leaving")
	thread_count--
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("Uh Oh..")
		log.Fatal(err)
	}
}
