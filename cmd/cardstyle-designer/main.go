package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Merith-TK/tcg-cardgen/pkg/designer"
)

func main() {
	var (
		port = flag.String("port", "3000", "Port to serve on")
		host = flag.String("host", "localhost", "Host to bind to")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintf(os.Stderr, "  %s                 # localhost:3000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -port 8080      # custom port\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -host 0.0.0.0   # all interfaces\n", os.Args[0])
	}
	flag.Parse()

	addr := fmt.Sprintf("%s:%s", *host, *port)

	srv := designer.New()

	fmt.Printf("TCG Cardstyle Designer  →  http://%s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")

	log.Fatal(http.ListenAndServe(addr, srv))
}
