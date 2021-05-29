package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var (
	appAddress string
	ccAddress  string
	ddAddress  string
)

func init() {
  flag.StringVar(&appAddress, "a", "127.0.0.1:4444", "connect local app server.")
	flag.StringVar(&ccAddress, "c", "x.x.x.x:3331", "connect control server.")
	flag.StringVar(&ddAddress, "d", "x.x.x.x:3332", "connect data server.")
}

func start() {
	for {
		//reconnection the server, while the connection is close.
		cc, err := net.Dial("tcp", ccAddress)
		if err != nil {
			log.Println(err.Error())
			if cc != nil {
				cc.Close()
			}
			time.Sleep(time.Second * 10)
			continue
		}

		for {
			data := make([]byte, 8)
			n, err := cc.Read(data)
			if err != nil {
				if err == io.EOF {
					log.Println("server is close.")
					cc.Close()
					break
				}
				continue
			}

			isok := string(data[0:n])
			log.Println(isok)
			if isok == "OK" || isok == "OKOK" {
				go connect()
			}
		}
	}

}

func connect() {
	log.Println("connect local app...")
	uc, err := net.Dial("tcp", appAddress)
	if err != nil {
		log.Println(err.Error())
		if uc != nil {
			uc.Close()
		}
		return
	}

	log.Println("connect ", appAddress, "success.")

	tc, err := net.Dial("tcp", ddAddress)
	if err != nil {
		log.Println(err.Error())
		if tc != nil {
			tc.Close()
		}
		return
	}
	log.Println("connect ", ddAddress, "success.")

	go send(&tc, &uc)
	go revice(&tc, &uc)

}

func send(tc *net.Conn, uc *net.Conn) {
	log.Println("come in send...")
	n, err := io.Copy(*uc, *tc)
	log.Println("send:", n, err)
	if (*uc) != nil {
		(*uc).Close()
		log.Println("send uc close")
	}
}

func revice(tc *net.Conn, uc *net.Conn) {
	log.Println("come in revice...")
	n, err := io.Copy(*tc, *uc)
	log.Println("revice:", n, err)
	if (*tc) != nil {
		(*tc).Close()
		log.Println("revice tc close")
	}
}

func main() {
	flag.Parse()
	flag.Usage()
	go start()
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, os.Kill)
	<-interrupt
	fmt.Println("'Ctrl + C' received, exiting...")
}
