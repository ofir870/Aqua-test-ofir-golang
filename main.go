package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Users struct {
	db *sql.DB
}

type Host struct {
	ID         int    `json:"id"`
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	IP_Address string `json:"ip_address"`
}
type Container struct {
	ID         int    `json:"id"`
	Host_ID    int    `json:"host_id"`
	Name       string `json:"name"`
	Image_Name int    `json:"image_name"`
}

func main() {

	os.Remove("aqua.db")

	log.Println("Creating aqua.db...")

	file, err := os.Create("aqua.db")

	if err != nil {

		log.Fatal(err.Error())

	}
	file.Close()
	log.Println("aqua.db created")

	sqliteDatabase, _ := sql.Open("sqlite3", "./aqua.db")
	defer sqliteDatabase.Close()
	createTables(sqliteDatabase)

	// INSERT RECORDS
	insertHost(sqliteDatabase, 1, "4e9edc48-2869-4172-903d-65008fd2895e", "AWS Host", "1.2.3.4")
	insertHost(sqliteDatabase, 2, "f89cda2e-628a-4f6e-b1d8-1ecf389e2454", "Azure Host", "4.5.6.7")
	insertHost(sqliteDatabase, 3, "863d9084-935e-4c71-990d-3a7dad113097", "GCP Host", "7.8.9.0")
	insertHost(sqliteDatabase, 4, "86d33421-8945-4a50-bcf6-fd3750e51942", "IBM Host", "4.5.6.7")

	handleRequests()
}

//
// INNER FUNCTIONS
//

func handleRequests() {

	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage)

	db, _ := sql.Open("sqlite3", "./aqua.db") // Open the created SQLite File
	defer db.Close()                          // Defer Closing the database

	// -------------------------------ALL API CALLS -------------------------------
	// ---------------------------!!!!!!!!! -------------------------------
	// returnAllHosts :>> /host
	myRouter.HandleFunc("/host", func(w http.ResponseWriter, r *http.Request) {
		returnAllHosts(w, r, db)
	})
	// returnAllContainers :>> /container
	myRouter.HandleFunc("/container", func(w http.ResponseWriter, r *http.Request) {
		returnAllContainers(w, r, db)
	})
	// createNewContainer :>> /container/create
	myRouter.HandleFunc("/container/create", func(w http.ResponseWriter, r *http.Request) {
		createNewContainer(w, r, db)
	}).Methods(("POST"))
	// returnSingleHost :>> /host/{id}
	myRouter.HandleFunc("/host/{id}", func(w http.ResponseWriter, r *http.Request) {
		returnSingleHost(w, r, db)
	})
	// returnSingleContainer :>> /container/{id}
	myRouter.HandleFunc("/container/{id}", func(w http.ResponseWriter, r *http.Request) {
		returnSingleContainer(w, r, db)
	})
	// returnAllContainerByName :>> /container/sort/{host_id}
	myRouter.HandleFunc("/container/sort/{host_id}", func(w http.ResponseWriter, r *http.Request) {
		returnAllContainerByHostID(w, r, db)
	})

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func createTables(db *sql.DB) {
	createHostTableSQL := `CREATE TABLE hosts (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"uuid" TEXT,
		"name" TEXT,
		"ip_address" TEXT		
	  );`

	log.Println("Create Host table...")
	statementHost, err := db.Prepare(createHostTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statementHost.Exec() // Execute SQL Statements

	createContainerTableSQL := `CREATE TABLE containers (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT,		
		"host_id" integer,
		"image_name" integer		
		
	  );`

	log.Println("Create Host table...")

	statementCon, err := db.Prepare(createContainerTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statementCon.Exec()

	log.Println("containers table created")
}

func insertHost(db *sql.DB, id int, uuid string, name string, ip_address string) {
	log.Println("Inserting Host record ...")
	insertHostSQL := `INSERT INTO hosts(id, uuid, name, ip_address) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertHostSQL)

	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, uuid, name, ip_address)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func insertContainer(db *sql.DB, id int, name string, host_id int, image_name int) {
	log.Println("Inserting containers record ...")

	insertContainerSQL := `INSERT INTO containers(id, name, host_id, image_name) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertContainerSQL)

	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, name, host_id, image_name)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

//
//
//
//
//  API FUNCTIONS
//
//
//

func createNewContainer(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	reqBody, _ := ioutil.ReadAll(r.Body)
	var container Container
	json.Unmarshal(reqBody, &container)

	aquaDB, _ := sql.Open("sqlite3", "./aqua.db")
	defer aquaDB.Close()

	insertContainer(aquaDB, container.ID, container.Name, container.Host_ID, container.Image_Name)

	json.NewEncoder(w).Encode(container)
}

func returnAllHosts(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	row, err := db.Query("SELECT * FROM hosts ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var arr []Host
	for row.Next() {

		var id int
		var uuid string
		var name string
		var ip_address string

		row.Scan(&id, &uuid, &name, &ip_address)
		var host Host

		host.ID = id
		host.UUID = uuid
		host.Name = name
		host.IP_Address = ip_address

		arr = append(arr, host)

	}

	json.NewEncoder(w).Encode(arr)
}

func returnAllContainers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	row, err := db.Query("SELECT * FROM containers ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var arr []Container
	for row.Next() {

		var id int
		var name string
		var host_id int
		var image_name int

		row.Scan(&id, &name, &host_id, &image_name)
		var container Container

		container.ID = id
		container.Name = name
		container.Host_ID = host_id
		container.Image_Name = image_name

		arr = append(arr, container)

	}

	json.NewEncoder(w).Encode(arr)
}

func returnSingleHost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	key := vars["id"]

	var sb strings.Builder
	sb.WriteString("SELECT * FROM hosts WHERE id = ")
	sb.WriteString(key)

	row, err := db.Query(sb.String())
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var host Host
	for row.Next() {
		var id int
		var uuid string
		var name string
		var ip_address string

		row.Scan(&id, &uuid, &name, &ip_address)

		host.ID = id
		host.UUID = uuid
		host.Name = name
		host.IP_Address = ip_address
	}
	json.NewEncoder(w).Encode(host)
}

func returnSingleContainer(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	vars := mux.Vars(r)
	key := vars["id"]

	var sb strings.Builder
	sb.WriteString("SELECT * FROM containers WHERE id = ")
	sb.WriteString(key)
	fmt.Println(sb.String())
	row, err := db.Query(sb.String())
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var container Container
	for row.Next() {

		var id int
		var name string
		var host_id int
		var image_name int

		row.Scan(&id, &name, &host_id, &image_name)

		container.ID = id
		container.Name = name
		container.Host_ID = host_id
		container.Image_Name = image_name

	}

	json.NewEncoder(w).Encode(container)
}

func returnAllContainerByHostID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	key := vars["host_id"]

	var sb strings.Builder
	sb.WriteString("SELECT * FROM containers WHERE host_id = ")
	sb.WriteString(key)

	row, err := db.Query(sb.String())
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var container Container
	var arr []Container

	for row.Next() {

		var id int
		var name string
		var host_id int
		var image_name int

		row.Scan(&id, &name, &host_id, &image_name)

		container.ID = id
		container.Name = name
		container.Host_ID = host_id
		container.Image_Name = image_name
		arr = append(arr, container)
	}

	json.NewEncoder(w).Encode(arr)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the OfirsTest!")
	fmt.Println("Endpoint Hit: homePage")
}
