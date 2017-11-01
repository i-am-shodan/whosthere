package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	"gopkg.in/yaml.v2"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
)

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	HostKeyLocation string `yaml:"HostKeyLocation"`

	MySQL string `yaml:"MySQL"`

	Listen string `yaml:"Listen"`
	Debug  string `yaml:"Debug"`
}

func main() {
	configText, err := ioutil.ReadFile("config.yml")
	fatalIfErr(err)
	var C Config
	fatalIfErr(yaml.Unmarshal(configText, &C))

	go func() {
		log.Println(http.ListenAndServe(C.Debug, nil))
	}()

	db, err := sql.Open("mysql", C.MySQL)
	fatalIfErr(err)
	fatalIfErr(db.Ping())
	_, err = db.Exec("SET NAMES UTF8")
	fatalIfErr(err)
	
	_, err = db.Exec("select id from events limit 1")
	fatalIfErr(err)

	server := &Server{
		dbConnectionString: C.MySQL,
		sessionInfo: make(map[string]sessionInfo),
	}
	server.sshConfig = &ssh.ServerConfig{
		KeyboardInteractiveCallback: server.KeyboardInteractiveCallback,
		PublicKeyCallback:           server.PublicKeyCallback,
	}

	privateBytes, err := ioutil.ReadFile(C.HostKeyLocation)
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	fatalIfErr(err)
	server.sshConfig.AddHostKey(private)

	listener, err := net.Listen("tcp", C.Listen)
	fatalIfErr(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept failed:", err)
			continue
		}

		go server.Handle(conn)
	}
}
