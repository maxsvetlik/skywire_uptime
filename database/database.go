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
	UID       int64
	Email     string
	PublicKey string
	FirstSeen time.Time
	LastSeen  time.Time
	TimesSeen int64
}

// TasksSchema represents the schema in the database associated with tasks
type TasksSchema struct {
	id     int64
	userid int64
	tasks  string
}

// ConnectDB connects to the databse and return a connection object
func ConnectDB(dbName string) *DbConn {
	db, err := sql.Open("sqlite3", dbName)
	checkError(err, "Failed to load database")
	err = db.Ping()
	checkError(err, "Failed to ping database")
	return &DbConn{db: db}
}

// InsertUser will create a new user by inserting them into the users and tasks tables
func (dbc *DbConn) InsertUser(email string, publickey int64, first_seen time.Time, last_seen time.Time, time_seen int64) int64 {

	tx, err := dbc.db.Begin()
	checkError(err, "Failed to create transaction")
	defer tx.Rollback() // in case the tx couldn't get committed

	nodeAdd, err := tx.Prepare("INSERT INTO nodes (email, publickey, first_seen, last_seen, time_seen) VALUES (?,?,?,?,?)")
	checkError(err, "Failed to prepare user statement")
	defer nodeAdd.Close()

	_, err = nodeAdd.Exec(email, publickey, first_seen, last_seen, time_seen)
	checkError(err, "Failed to execute user statement")

	return time_seen
}

// UpdateUserToken will update a user's token data
/*
func (dbc *DbConn) UpdateUserToken(userID int64, tokenAccess, tokenRefresh string, tokenExpiry time.Time, tokenType string) error {
	//create transaction
	tx, err := dbc.db.Begin()
	checkError(err, "Failed to create transaction")
	defer tx.Rollback()

	// prepare query
	prefsUpdate, err := tx.Prepare("UPDATE users SET tokenAccess=?, tokenRefresh=?, tokenExpiry=?, tokenType=? WHERE uid=?")
	checkError(err, "Failed to prepare token update statement")
	defer prefsUpdate.Close()

	// execute query
	_, err = prefsUpdate.Exec(tokenAccess, tokenRefresh, tokenExpiry, tokenType, userID)
	checkError(err, "Failed to execute token update query")

	// commit transaction
	err = tx.Commit()
	checkError(err, "Failed to commit token update transaction")

	return nil
}
*/
// GetUserTokenData will get a user's token using their UID
func (dbc *DbConn) GetUserTokenData(uid int64) (string, string, time.Time, string, error) {
	row := dbc.db.QueryRow("SELECT tokenAccess,tokenRefresh,tokenExpiry,tokenType FROM users WHERE uid= ?", uid)

	var access, refresh, ttype string
	var expiry time.Time

	err := row.Scan(&access, &refresh, &expiry, &ttype)
	if err == sql.ErrNoRows {
		return access, refresh, expiry, ttype, ErrUserNotFound
	} else if err != nil {
		log.Printf("Failed to scan user get query for uid: %s\n", uid)
		log.Println(err)
		return access, refresh, expiry, ttype, err
	}

	return access, refresh, expiry, ttype, nil
}

// SetupDb will create the users and tasks tables (only used by the tools package)
func (dbc *DbConn) SetupDb() {
	createNodesCmd := `CREATE TABLE 'nodes' (
    'uid' INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    'email' NVARCHAR(90) NOT NULL,
    'key' CHARACTER(66) NOT NULL,
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
