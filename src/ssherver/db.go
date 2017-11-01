package main

import (
    "database/sql"
    
    _ "github.com/go-sql-driver/mysql"
)

/*

events
======
pk: id : int
timestamp : datetime
request_type : int
client_id : fk->client

CREATE TABLE events
(
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    client_id INT NOT NULL,
    FOREIGN KEY (client_id)
        REFERENCES client_versions(id)
        ON DELETE CASCADE
);

keys
====
pk : id : int
key : byte[]

CREATE TABLE publickeys
(
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    hash varchar(32) UNIQUE,
    publickey TEXT NOT NULL
);


key_events
==========
event_id -> fk->events : pk
key_id -> fk->keys : pk 

CREATE TABLE key_events (
    event_id INT,
    key_id INT,
    FOREIGN KEY (event_id) REFERENCES events(id),
    FOREIGN KEY (key_id) REFERENCES publickeys(id),
    PRIMARY KEY (event_id, key_id)
);



client
======
pk : id : int
version : string

CREATE TABLE client_versions
(
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    version varchar(100) CHARACTER SET latin1 COLLATE latin1_general_cs UNIQUE
);



DELIMITER //
CREATE PROCEDURE addEvent
(
    IN param_version varchar(100)
)
BEGIN
    DECLARE clientID INT;
    DECLARE eventID INT;

    INSERT INTO `client_versions` (version) VALUES (param_version) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id), `version`=param_version;
    SELECT LAST_INSERT_ID() INTO clientID;

    INSERT INTO `events` (client_id) VALUES (clientID);
    SELECT LAST_INSERT_ID() INTO eventID;

    SELECT eventID;

END //
DELIMITER ;


DELIMITER //
CREATE PROCEDURE addKeyToEvent
(
    IN param_eventID int,
    IN param_publickey TEXT
)
BEGIN
    DECLARE keyID INT;
    DECLARE md5 varchar(32);

    SELECT MD5(param_publickey) into md5;

    INSERT INTO `publickeys` (hash, publickey) VALUES (md5, param_publickey) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id), `hash`=md5;
    SELECT LAST_INSERT_ID() INTO keyID;

    INSERT INTO `key_events` (event_id, key_id) VALUES (param_eventID, keyID);

END //
DELIMITER ;



call addEvent('client1');
call addKeyToEvent(2, '0,12,14');

*/

func addToDatabase(dbConnectionString string, logEntry* logEntry) {
    db, err := sql.Open("mysql", dbConnectionString)
    if err != nil {
        panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
    }
    defer db.Close()

    // Open doesn't open a connection. Validate DSN data:
    err = db.Ping()
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }

    // Prepare statement for reading data
	addEventStatement, err := db.Prepare("call addEvent(?)")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
    defer addEventStatement.Close()

    // execute addEvent
    var eventID int // we "scan" the result in here
	err = addEventStatement.QueryRow(logEntry.ClientVersion).Scan(&eventID) // WHERE number = 13
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
    }
    
	// Prepare statement for inserting data
	addKeyStatement, err := db.Prepare("call addKeyToEvent(?, ?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer addKeyStatement.Close() // Close the statement when we leave main() / the program terminates

    for key := range logEntry.KeysOffered {
		_, err = addKeyStatement.Exec(eventID, key) // Insert tuples (i, i^2)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
	}
}
