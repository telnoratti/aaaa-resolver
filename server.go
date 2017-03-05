/*
   DNS server that will return AAAA records all ipv6 records.
   Copyright (C) 2017 Calvin Winkowski <calvin@winkowski.me>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/
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
	"time"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpuprofile to file")
	debug      = flag.Bool("debug", false, "debugging output")
	zone       = flag.String("zone", "ipv6-literal", "zone to generate subdomains for")
	port       = flag.Int("port", 8053, "port to run on")
	ns         = flag.String("ns", "ipv6-literal.", "NS")
	mbox       = flag.String("mbox", "", "mbox string for SOA")
)

func handleLiteral(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	if r.Question[0].Qtype == dns.TypeSOA {
		log.Println("SOA reply" + *zone)
		serial, err := strconv.Atoi(time.Now().Format("20060102"))
		if err != nil {
			log.Fatalln(err.Error())
		}
		rr := &dns.SOA{
			Hdr: dns.RR_Header{
				Name:   *zone,
				Rrtype: dns.TypeSOA,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Ns:      *ns,
			Mbox:    *mbox,
			Serial:  uint32(serial),
			Refresh: 43200,
			Retry:   180,
			Expire:  2419200,
			Minttl:  10800,
		}
		m.Answer = append(m.Answer, rr)
		w.WriteMsg(m)
		return
	}
	if r.Question[0].Qtype == dns.TypeNS && r.Question[0].Name == *zone {
		log.Println("NS reply" + *ns)
		rr := &dns.NS{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeNS,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Ns: *ns,
		}
		m.Answer = append(m.Answer, rr)
		w.WriteMsg(m)
		return
	}

	if r.Question[0].Qtype != dns.TypeAAAA {
		log.Println("Wrong query type")
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}

	// This is a AAAA request now
	query := r.Question[0].Name
	if len(query)-len(*zone) <= 0 {
		// This is not something we'll respond to, probably the tld itself
		log.Println("Length is wrong")
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}
	// query is simply the v6 address
	query = query[:len(query)-len(*zone)-1]
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
	// Use only an FQDN
	z := dns.Fqdn(*zone)
	zone = &z
	n := dns.Fqdn(*ns)
	ns = &n
	var e string
	// Pick a good looking email if none provided
	if *mbox == "" {
		e = "hostmaster." + *zone
	} else {
		e = dns.Fqdn(*mbox)
	}
	mbox = &e

	dns.HandleFunc(*zone, handleLiteral)
	go serve("tcp", *port)
	go serve("udp", *port)
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	fmt.Printf("Signal (%s) received, stopping\n", s)
}
