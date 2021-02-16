package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/seer-robotics/escpos"
)

func NewPrinterService() (*printerService, error) {

	d := net.Dialer{
		Timeout: time.Second * 3,
	}
	psock, err := d.Dial("tcp", "localhost:9100")
	if err != nil {
		return nil, err
	}

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
		p.pr.Write(order.ID)
		p.pr.Linefeed()
		for _, mi := range order.OrderedItems {
			p.pr.SetAlign("left")
			p.pr.Write(mi.Name)
			p.pr.SetAlign("right")
			p.pr.Write(strconv.Itoa(mi.Price))
			p.pr.Linefeed()
		}
		p.pr.Write("Order Created:")
		p.pr.Linefeed()
		p.pr.Write(fmt.Sprintf("%02d:%02d:%02d", order.SubmittedTime.Local().Hour(), order.SubmittedTime.Local().Minute(), order.SubmittedTime.Local().Second()))
		p.pr.Linefeed()
		p.pr.Cut()
		p.pw.Flush()
		time.Sleep(3 * time.Second)
	}
}
