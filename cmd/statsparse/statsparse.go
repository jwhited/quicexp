package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	flagName   = flag.String("name", "", "test name")
	flagHeader = flag.Bool("header", false, "output header record")
)

func main() {
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	writer := csv.NewWriter(os.Stdout)
	header := []string{
		"label",
		"Throughput(kbps)",
		"EcnCapable",
		"RTT(us)",
		"SendTotalPackets",
		"SendSuspectedLostPackets",
		"SendSpuriousLostPackets",
		"SendCongestionCount",
		"SendEcnCongestionCount",
		"RecvTotalPackets",
		"RecvReorderedPackets",
		"RecvDroppedPackets",
		"RecvDuplicatePackets",
		"RecvDecryptionFailures",
		"RecvMaxCoalescedCount",
	}
	if *flagHeader {
		err := writer.Write(header)
		if err != nil {
			log.Fatalf("error writing: %v", err)
		}
	}
	records := make([]string, 0, len(header))
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.HasPrefix(txt, "[conn]") {
			if len(records) > 0 {
				err := writer.Write(records)
				if err != nil {
					log.Fatalf("error writing: %v", err)
				}
				for i := range records {
					records[i] = ""
				}
				records = records[:0]
			}
			records = records[:cap(records)]
			words := strings.Split(txt, " ")
			records[0] = *flagName
			j := 2
			for _, word := range words {
				split := strings.Split(word, "=")
				if len(split) == 2 {
					if j == len(header) {
						log.Fatalf("unexpected [conn] line: %v", txt)
					}
					records[j] = split[1]
					j++
				}
			}
		} else if strings.HasPrefix(txt, "Result:") {
			words := strings.Split(txt, " ")
			if len(words) < 8 || words[5] != "kbps" {
				log.Fatalf("unexpected Result line: %v", txt)
			}
			records[0] = *flagName
			records[1] = words[4]
			err := writer.Write(records)
			if err != nil {
				log.Fatalf("error writing: %v", err)
			}
			for i := range records {
				records[i] = ""
			}
			records = records[:0]
		}
	}
	if len(records) > 0 {
		err := writer.Write(records)
		if err != nil {
			log.Fatalf("error writing: %v", err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	writer.Flush()
}
