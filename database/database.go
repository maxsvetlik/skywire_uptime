package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // use sqllite3 as the database
)

var defaultStartTime, _ = time.Parse("15:04", "09:00")
var defaultEndTime, _ = time.Parse("15:04", "22:00")
var defaultBlockTime = 30

// ErrUserNotFound is thrown whenever a user is not found
var ErrUserNotFound = errors.New("User not found")

// DbConn stores the currently active database connection
type DbConn struct {
	db *sql.DB
}

type Node struct {
	PublicKey string
	FirstSeen time.Time
	LastSeen  time.Time
	TimesSeen int64
}

// ConnectDB connects to the databse and return a connection object
func ConnectDB(dbName string) *DbConn {
	db, err := sql.Open("sqlite3", dbName)
	checkError(err, "Failed to load database")
	err = db.Ping()
	checkError(err, "Failed to ping database")
	return &DbConn{db: db}
}

// InsertNode will create a new user by inserting them into the users and tasks tables
func (dbc *DbConn) InsertNode(publicKey string, first_seen time.Time, last_seen time.Time, time_seen int64) *Node {

	tx, err := dbc.db.Begin()
	checkError(err, "Failed to create transaction")
	defer tx.Rollback() // in case the tx couldn't get committed

	nodeAdd, err := tx.Prepare("INSERT INTO nodes (publicKey, first_seen, last_seen, time_seen) VALUES (?,?,?,?,?)")
	checkError(err, "Failed to prepare user statement")
	defer nodeAdd.Close()

	_, err = nodeAdd.Exec(publicKey, first_seen, last_seen, time_seen)
	checkError(err, "Failed to execute user statement")

	return &Node{publicKey, first_seen, last_seen, time_seen}
}

// GetNodeByKey will get a node's data using its private key
func (dbc *DbConn) GetNodeByKey(publicKey string) (*Node, error) {
	row := dbc.db.QueryRow("SELECT * FROM nodes WHERE publickey = ?", publicKey)

	n := &Node{}
	err := row.Scan(&n.PublicKey, &n.FirstSeen, &n.LastSeen, &n.TimesSeen)
	if err == sql.ErrNoRows {
		return n, ErrUserNotFound
	} else if err != nil {
		log.Printf("Failed to scan user get query for email: %s\n", publicKey)
		log.Println(err)
		return n, err
	}

	return n, nil
}

// SetupDb will create the users and tasks tables (only used by the tools package)
func (dbc *DbConn) SetupDb() {
	createNodesCmd := `CREATE TABLE 'nodes' (
    'publickey' CHARACTER(66) PRIMARY KEY NOT NULL,
    'first_seen' TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	'last_seen' TIMESTAMP,
	'times_seen' BIGINT
);`
	dbc.createTable(createNodesCmd)
}

func (dbc *DbConn) createTable(cmd string) {
	_, err := dbc.db.Exec(cmd)
	if err != nil {
		fmt.Printf("Failed to create table:\n %s\n", cmd)
		log.Fatal(err)
	}
}

// Close a dababase connection
func (dbc *DbConn) Close() {
	dbc.db.Close()
}

func checkError(err error, str string) {
	if err != nil {
		log.Println(str)
		log.Fatal(err)
	}
}
