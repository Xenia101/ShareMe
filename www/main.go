package www

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RyuaNerin/ShareMe/share"
	"github.com/gin-gonic/gin"
)

func Main() {
	// init HTTP
	g := gin.New()
	initGin(g)

	server := http.Server{
		Handler: g,
	}

	// Listen
	var l net.Listener
	var err error

	if _, err := net.ResolveTCPAddr("tcp", share.Config.Listen); err == nil {
		l, err = net.Listen("tcp", share.Config.Listen)
	} else {
		if _, err := os.Stat(share.Config.Listen); !os.IsNotExist(err) {
			err := os.Remove(share.Config.Listen)
			if err != nil {
				panic(err)
			}
		}

		l, err = net.Listen("unix", share.Config.Listen)
		if err != nil {
			panic(err)
		}
		err = os.Chmod(share.Config.Listen, 0777)
	}
	if err != nil {
		panic(err)
	}

	// Serve
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	log.Println("Serve")
	go server.Serve(l)

	<-sig

	log.Println("Exit")
}
