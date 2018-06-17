package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // use sqllite3 as the database
)

var ErrNodeNotFound = errors.New("Node not found")

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

type Search struct {
	SearchID       int
	NumNodesOnline int
	Timestamp      time.Time
}

// ConnectDB connects to the databse and return a connection object
func ConnectDB(dbName string) *DbConn {
	db, err := sql.Open("sqlite3", dbName+"?parseTime=true")
	checkError(err, "Failed to load database")
	err = db.Ping()
	checkError(err, "Failed to ping database")
	return &DbConn{db: db}
}

// InsertNode will create a new user by inserting them into the users and tasks tables
func (dbc *DbConn) InsertNode(public_key string, first_seen time.Time, last_seen time.Time, times_seen int64) (*Node, error) {
	var return_error error
	tx, err := dbc.db.Begin()
	checkError(err, "Failed to create transaction")
	if err != nil {
		return_error = err
	}

	defer tx.Rollback() // in case the tx couldn't get committed

	nodeAdd, err := tx.Prepare("INSERT INTO nodes(public_key, first_seen, last_seen, times_seen) VALUES (?,?,?,?)")
	checkError(err, "Failed to prepare user statement")
	if err != nil {
		return_error = err
	}
	defer nodeAdd.Close()

	_, err = nodeAdd.Exec(public_key, first_seen, last_seen, times_seen)
	checkError(err, "Failed to execute user statement")
	if err != nil {
		return_error = err
	}

	// commit transaction
	err = tx.Commit()
	checkError(err, "Failed to commit tasks update transaction")
	if err != nil {
		return_error = err
	}

	return &Node{public_key, first_seen, last_seen, times_seen}, return_error
}

// Inserts a new row for network statistics
func (dbc *DbConn) InsertSearch(num_nodes int, timestamp time.Time) (*Search, error) {
	var return_error error
	tx, err := dbc.db.Begin()
	checkError(err, "Failed to create transaction")
	if err != nil {
		return_error = err
	}
	defer tx.Rollback() // in case the tx couldn't get committed

	nodeAdd, err := tx.Prepare("INSERT INTO search(num_online_nodes, timestamp) VALUES (?,?)")
	checkError(err, "Failed to prepare user statement")
	if err != nil {
		return_error = err
	}
	defer nodeAdd.Close()

	_, err = nodeAdd.Exec(num_nodes, timestamp)
	checkError(err, "Failed to execute user statement")
	if err != nil {
		return_error = err
	}

	// commit transaction
	err = tx.Commit()
	checkError(err, "Failed to commit tasks update transaction")
	if err != nil {
		return_error = err
	}

	return &Search{-1, num_nodes, timestamp}, return_error

}

// Gets the last search inserted for front page stats
func (dbc *DbConn) GetLastSearch() (*Search, error) {
	row := dbc.db.QueryRow("SELECT * FROM search ORDER BY search_id DESC LIMIT 1")
	s := &Search{}
	err := row.Scan(&s.SearchID, &s.NumNodesOnline, &s.Timestamp)
	if err == sql.ErrNoRows {
		return s, ErrNodeNotFound
	} else if err != nil {
		log.Printf("Failed to scan search get query for last entry.")
		log.Println(err)
		return s, err
	}

	return s, nil
}

// GetNodeByKey will get a node's data using its private key
func (dbc *DbConn) GetNodeByKey(public_key string) (*Node, error) {
	row := dbc.db.QueryRow("SELECT * FROM nodes WHERE public_key = ?", public_key)
	n := &Node{}
	err := row.Scan(&n.PublicKey, &n.FirstSeen, &n.LastSeen, &n.TimesSeen)
	if err == sql.ErrNoRows {
		return n, ErrNodeNotFound
	} else if err != nil {
		log.Printf("Failed to scan user get query for public_key: %s\n", public_key)
		log.Println(err)
		return n, err
	}

	return n, nil
}

// Updates times seen and lastTimeSeen for given nodeID
func (dbc *DbConn) UpdateNode(publicKey string, lastTimeSeen time.Time) error {

	//create transaction
	tx, err := dbc.db.Begin()
	checkError(err, "Failed to create transaction")
	defer tx.Rollback()

	// prepare query
	nodeUpdate, err := tx.Prepare("UPDATE nodes SET last_seen=?, times_seen=times_seen + 1 WHERE public_key=?")
	checkError(err, "Failed to prepare node update statement")
	defer nodeUpdate.Close()

	// execute query
	_, err = nodeUpdate.Exec(lastTimeSeen, publicKey)
	checkError(err, "Failed to execute node update query")

	// commit transaction
	err = tx.Commit()
	checkError(err, "Failed to commit node update transaction")

	return nil
}

// SetupDb will create the users and tasks tables (only used by the tools package)
func (dbc *DbConn) SetupDb() {
	createNodesCmd := `CREATE TABLE 'nodes' (
    'public_key' NVARCHAR(66) PRIMARY KEY NOT NULL,
    'first_seen' TIMESTAMP NULL,
	'last_seen' TIMESTAMP NULL,
	'times_seen' BIGINT NULL
);`

	createSearchCmd := `CREATE TABLE 'search' (
	'search_id' INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	'num_online_nodes' INTEGER, 
	'timestamp' TIMESTAMP
);`

	dbc.createTable(createNodesCmd)
	dbc.createTable(createSearchCmd)
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
