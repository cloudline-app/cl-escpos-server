package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/seer-robotics/escpos"
	"github.com/spf13/viper"
)

func NewPrinterService(address string, port int) (*printerService, error) {

	d := net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Second * 15,
	}

	fmt.Printf("Connecting to printer at: %s:%d \n", address, port)

	daddress := address + ":" + strconv.Itoa(port)
	psock, err := d.Dial("tcp", daddress)
	if err != nil {
		fmt.Printf("Failed to connect to printer at: %s:%d \n", address, port)
		return nil, err
	}

	err = psock.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		fmt.Printf("Failed to set keepalive connection to printer at: %s:%d \n", address, port)
		return nil, err
	}

	//create a writer for us to add to on the socket
	pw := bufio.NewWriter(psock)

	//create a printer to write to
	pr := escpos.New(pw)

	ps := printerService{
		pr:        pr,
		pw:        pw,
		orderChan: make(chan Order, 10000),
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
		t := fmt.Sprintf("%02d:%02d", order.SubmittedTime.Local().Hour(), order.SubmittedTime.Local().Minute())
		if viper.GetBool("server.debug") == true {
			p.pr.Verbose = true
		}
		p.pr.Init()

		p.pr.SetAlign("center")
		p.pr.SetFontSize(1, 1)
		p.pr.SetEmphasize(1)
		p.pr.Write("Tulleys Drive In Movies")
		p.pr.SetEmphasize(0)
		p.pr.Linefeed()
		p.pr.Linefeed()

		shortOrderID := ""
		if len(order.ID) > 4 {
			shortOrderID = order.ID[:4]
		} else {
			shortOrderID = order.ID
		}

		writeLargeItem(p.pr, "Order ID", shortOrderID)
		writeLargeItem(p.pr, "Order Time", t)

		for _, o := range order.OrderInformation {
			writeOrderInformation(p.pr, o)
		}

		writeOrderItems(p.pr, order.OrderedItems)

		p.pr.Cut()
		p.pw.Flush()
	}
}

func writeOrderInformation(p *escpos.Escpos, o OrderInformation) {
	if o.AnswerString != "" {
		writeLargeItem(p, o.Question, o.AnswerString)
	}
	if o.AnswerNumber != 0 {
		writeLargeItem(p, o.Question, strconv.Itoa(o.AnswerNumber))
	}
}

func writeOrderItems(p *escpos.Escpos, mis []MenuItem) {

	dict := make(map[string]int)
	for _, mi := range mis {
		dict[mi.Name] = dict[mi.Name] + 1
	}

	for name, number := range dict {
		// price := float32()
		p.SetAlign("left")
		p.SetEmphasize(1)
		p.Write(fmt.Sprintf("%dx ", number))
		p.Write(name + " ")
		p.SetEmphasize(0)
		// p.WriteWEU(fmt.Sprintf("£%.2f", (price / 100)))
		p.Linefeed()
	}
}

func writeLargeItem(p *escpos.Escpos, header string, value string) {
	p.SetFontSize(1, 1)
	p.SetEmphasize(0)
	p.Write(header)
	p.Linefeed()
	p.SetFontSize(2, 2)
	p.SetEmphasize(1)
	p.Write(value)
	p.SetEmphasize(0)
	p.SetFontSize(1, 1)
	p.Linefeed()
	p.Linefeed()
}
