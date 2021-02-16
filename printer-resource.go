package main

import (
	"bufio"
	"net"
	"time"

	"github.com/seer-robotics/escpos"
)

func NewPrinterService() (*printerService, error) {

	d := net.Dialer{
		Timeout: time.Second * 3,
	}
	psock, err := d.Dial("tcp", "localhost:1026")
	if err != nil {
		return nil, err
	}
	// defer psock.Close()

	//create a writer for us to add to on the socket
	pw := bufio.NewWriter(psock)

	//create a printer to write to
	pr := escpos.New(pw)

	ps := printerService{
		pr:        pr,
		pw:        pw,
		orderChan: make(chan Order, 1000),
	}

	go ps.printer(ps.orderChan)

	return &ps, nil
}

type printerService struct {
	pr        *escpos.Escpos
	pw        *bufio.Writer
	orderChan chan Order
}

func (p *printerService) AddToPrintQueue(o Order) error {
	p.orderChan <- o
	return nil
}

func (p *printerService) printer(orderChan <-chan Order) {
	for order := range orderChan {
		p.pr.Verbose = true
		p.pr.Init()
		p.pr.Beep(8)
		p.pr.Formfeed()
		p.pr.Write(order.name)
		p.pr.Formfeed()
		p.pr.Cut()
		p.pw.Flush()
		time.Sleep(3 * time.Second)
	}
}
