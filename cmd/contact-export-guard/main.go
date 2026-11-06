package main

import (
	"encoding/json"
	"flag"
	"fmt"
	guard "github.com/novasia-estate/property-contact-export-guard"
	"os"
)

func main() {
	input := flag.String("input", "", "UTF-8 CSV input")
	output := flag.String("out", "", "new vCard output path")
	audit := flag.String("audit", "", "new value-free JSON audit path")
	keyPath := flag.String("uid-key-file", "", "private file containing at least 32 bytes for stable HMAC UIDs")
	flag.Parse()
	if *input == "" || *output == "" || *audit == "" || *keyPath == "" {
		flag.Usage()
		os.Exit(2)
	}
	fail := func(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(2) }
	key, e := os.ReadFile(*keyPath)
	if e != nil {
		fail(e)
	}
	in, e := os.Open(*input)
	if e != nil {
		fail(e)
	}
	defer in.Close()
	result, e := guard.Convert(in, key)
	if e != nil {
		fail(e)
	}
	// Exclusive creation prevents accidentally replacing an existing export or source.
	out, e := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		fail(e)
	}
	report, e := os.OpenFile(*audit, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		out.Close()
		fail(e)
	}
	if _, e = out.WriteString(guard.VCard(result.Contacts)); e != nil {
		out.Close()
		report.Close()
		fail(e)
	}
	if e = out.Close(); e != nil {
		report.Close()
		fail(e)
	}
	enc := json.NewEncoder(report)
	enc.SetIndent("", "  ")
	if e = enc.Encode(result.Audit); e != nil {
		report.Close()
		fail(e)
	}
	if e = report.Close(); e != nil {
		fail(e)
	}
	fmt.Printf("Exported %d of %d rows; %d rows require review.\n", result.Audit.Exported, result.Audit.Rows, len(result.Audit.Issues))
}
