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

	daddress := address + ":" + strconv.Itoa(port)

	ps := printerService{
		orderChan: make(chan Order, 10000),
		address:   daddress,
	}

	go ps.printer(ps.orderChan)

	return &ps, nil
}

type printerService struct {
	pr        *escpos.Escpos
	pw        *bufio.Writer
	orderChan chan Order
	address   string
}

func (p *printerService) AddToPrintQueue(o Order) error {
	p.orderChan <- o
	return nil
}

func (p *printerService) printer(orderChan <-chan Order) {
	for order := range orderChan {

		d := net.Dialer{
			Timeout:   time.Second * 5,
			KeepAlive: time.Second * 15,
		}

		fmt.Printf("Connecting to printer at: %s \n", p.address)

		psock, err := d.Dial("tcp", p.address)
		if err != nil {
			fmt.Printf("Failed to connect to printer at: %s \n", p.address)
			return
		}

		if err != nil {
			fmt.Printf("Failed to set keepalive connection to printer at: %s \n", p.address)
			return
		}

		//create a writer for us to add to on the socket
		pw := bufio.NewWriter(psock)

		//create a printer to write to
		pr := escpos.New(pw)

		t := fmt.Sprintf("%02d:%02d", order.SubmittedTime.Local().Hour(), order.SubmittedTime.Local().Minute())
		if viper.GetBool("server.debug") == true {
			pr.Verbose = true
		}
		pr.Init()

		pr.SetAlign("center")
		pr.SetFontSize(1, 1)
		pr.SetEmphasize(1)
		pr.Write("Tulleys Drive In Movies")
		pr.SetEmphasize(0)
		pr.Linefeed()
		pr.Linefeed()

		shortOrderID := ""
		if len(order.ID) > 4 {
			shortOrderID = order.ID[:4]
		} else {
			shortOrderID = order.ID
		}

		writeLargeItem(pr, "Order ID", shortOrderID)
		writeLargeItem(pr, "Order Time", t)

		for _, o := range order.OrderInformation {
			writeOrderInformation(pr, o)
		}

		writeOrderItems(pr, order.OrderedItems)

		pr.Cut()
		pw.Flush()
		psock.Close()
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
