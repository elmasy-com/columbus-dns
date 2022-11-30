package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	sdk "github.com/elmasy-com/columbus-sdk"
	"github.com/miekg/dns"
)

var (
	ReplyChan     chan *dns.Msg
	WarnThreshold int
	Version       string
	Commit        string
	resolvers     []string
	resolversNum  int32
)

func getRandomResolver() string {

	if resolversNum == 1 {
		return resolvers[0]
	}

	return resolvers[rand.Int31n(resolversNum)]
}

// isValidResponse checks the type and the content of m.
// If m indicates a valid reply, returns true.
// This function is needed to not rely on RCODE only.
func isValidResponse(m dns.RR) bool {

	switch t := m.(type) {
	case *dns.SOA:
		// SOA returned if no record found
		return false
	case *dns.A:
		return true
	case *dns.AAAA:
		return true
	case *dns.CNAME:
		return true
	case *dns.MX:
		return true
	case *dns.TXT:
		return true
	case *dns.NS:
		return true
	case *dns.CERT:
		return true
	case *dns.SRV:
		return true
	case *dns.PTR:
		// PTR records are out of context
		return false
	default:
		fmt.Printf("Unknown reply type: %T\n", t)
		return false
	}
}

// insertWorker is a goroutine.
// NumWorkers controls the number of workers.
func insertWorker(wg *sync.WaitGroup) {

	defer wg.Done()

	for r := range ReplyChan {

		switch {
		case r.Answer == nil:
			continue
		case r.Question == nil:
			fmt.Fprintf(os.Stderr, "Error: question section is nil\n")
			continue
		case len(r.Question) == 0:
			fmt.Fprintf(os.Stderr, "Error: question section is empty\n")
			continue
		case len(r.Question) > 1:
			fmt.Fprintf(os.Stderr, "Error: multiple question\n")
			continue
		case len(r.Answer) == 0:
			fmt.Fprintf(os.Stderr, "Error: message answer is empty\n")
			continue
		case !isValidResponse(r.Answer[0]):
			// Further check requires, see: https://community.cloudflare.com/t/noerror-response-for-not-exist-domain-breaks-nslookup/173897
			continue
		}

		err := sdk.Insert(r.Question[0].Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to insert %s: %s\n", r.Question[0].Name, err)
		}

		if len(ReplyChan) > WarnThreshold {
			fmt.Fprintf(os.Stderr, "Number of reply messages in queue exceeds the threshold: %d\n", len(ReplyChan))
		}
	}
}

func handleFunc(w dns.ResponseWriter, q *dns.Msg) {

	start := time.Now()

	r, err := dns.Exchange(q, getRandomResolver())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to exchange message: %s\n", err)
		w.Close()
		return
	}
	if r == nil {
		fmt.Fprintf(os.Stderr, "Error: reply is nil\n")
		w.Close()
		return
	}

	if r.Rcode == 0 {
		ReplyChan <- r
	}

	err = w.WriteMsg(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write reply: %s\n", err)
	}

	fmt.Printf("%s -> %s %s %s %s %s\n",
		w.RemoteAddr().String(),
		q.Question[0].Name,
		dns.ClassToString[q.Question[0].Qclass],
		dns.TypeToString[q.Question[0].Qtype],
		dns.RcodeToString[r.Rcode],
		time.Since(start))
}

func main() {

	configPath := flag.String("config", "", "Path to the config file")
	printVersion := flag.Bool("version", false, "Print version")
	flag.Parse()

	// Print version and  exit
	if *printVersion {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", Commit)
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "-config is empty!\n")
		fmt.Printf("Use -help for help\n")
		os.Exit(1)
	}

	conf, err := parseConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config file: %s\n", err)
		os.Exit(1)
	}

	// Set global resolvers to get random ones
	resolvers = conf.Resolvers
	resolversNum = int32(len(resolvers))

	// Create buff channel
	ReplyChan = make(chan *dns.Msg, conf.BuffSize)
	WarnThreshold = conf.BuffSize / 10 * 9

	// Set ColumbusServer
	sdk.SetURI(conf.ColumbusServer)

	// Get Columbus user
	err = sdk.GetDefaultUser(conf.ApiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get Columbus user: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting %d workers...\n", conf.NumWorkers)
	// Start workers
	wg := sync.WaitGroup{}
	for i := 0; i < conf.NumWorkers; i++ {
		wg.Add(1)
		go insertWorker(&wg)
	}

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	udpServer := UDPStart(conf.ListenAddress, stopSignal)

	tcpServer := TCPStart(conf.ListenAddress, stopSignal)

	// Wait for the SIGTERM
	<-stopSignal
	fmt.Printf("Caught a SIGTERM, closing...\n")
	udpServer.Shutdown()
	tcpServer.Shutdown()
	close(ReplyChan)
	wg.Wait()
}
