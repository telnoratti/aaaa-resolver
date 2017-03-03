package main

import (
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpuprofile to file")
	debug      = flag.Bool("debug", false, "debugging output")
	zone       = flag.String("zone", "ipv6-literal", "zone to generate subdomains for")
	port       = flag.Int("port", 8053, "port to run on")
)

func handleLiteral(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	if r.Question[0].Qtype != dns.TypeAAAA {
		log.Println("Wrong query type")
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}

	// This is a AAAA request now
	query := r.Question[0].Name
	if len(query)-len(*zone)-1 <= 0 {
		// This is not something we'll respond to, probably the tld itself
		log.Println("Length is wrong")
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}
	// query is simply the v6 address
	query = query[:len(query)-len(*zone)-2]
	if strings.Count(query, ".") > 0 {
		// There is a subdomain query here, not something we care about
		// Someone also may have tried a v4 address, we don't do those
		log.Println("Too many .")
		log.Println(query)
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}

	// Now we know qstring should just be the mangled v6
	query = strings.Replace(query, "-", ":", -1)
	address := net.ParseIP(query)
	// For sanity's sake we'll ensure it's a v6 address
	if address.To16() == nil {
		// This is not an ipv6 address.
		log.Println("Not an ipv6")
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}
	log.Println("Sending reply")
	rr := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   r.Question[0].Name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		AAAA: address,
	}
	m.Answer = append(m.Answer, rr)
	w.WriteMsg(m)
}

func serve(net string, port int) {
	server := &dns.Server{Addr: ":" + strconv.FormatInt(int64(port), 10), Net: net, TsigSecret: nil}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln("Failed to setup the "+net+" server: $s\n", err.Error())
	}
}

func main() {
	flag.Parse()
	dns.HandleFunc(*zone, handleLiteral)
	go serve("tcp", *port)
	go serve("udp", *port)
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	fmt.Printf("Signal (%s) received, stopping\n", s)
}
