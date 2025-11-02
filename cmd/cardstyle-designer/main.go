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
		help = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("TCG Cardstyle Designer - Web-based template editor")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Printf("  %s                    # Start on localhost:8080\n", os.Args[0])
		fmt.Printf("  %s -port 3000        # Start on localhost:3000\n", os.Args[0])
		fmt.Printf("  %s -host 0.0.0.0     # Listen on all interfaces\n", os.Args[0])
		return
	}

	addr := fmt.Sprintf("%s:%s", *host, *port)

	// Create the web server
	server, err := designer.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("🎨 TCG Cardstyle Designer starting on http://%s\n", addr)
	fmt.Println("   📝 Create custom cardstyles with visual editor")
	fmt.Println("   🎯 No external dependencies, secure Go-hosted")
	fmt.Println("   🔒 Zero npm packages, controlled supply chain")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	// Start the server
	log.Fatal(http.ListenAndServe(addr, server))
}
