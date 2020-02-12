package main

import (
	"flag"
	"log"
	"os"

	"github.com/cloudflare/cloudflare-go"
	externalip "github.com/glendc/go-external-ip"
)

func main() {
	zoneName := flag.String("zone", "", "zone to update")
	recordName := flag.String("record", "", "record in zone to update")
	flag.Parse()

	log.Printf("syncing record for %s with current public IP address...\n", *recordName)

	cf, err := cloudflare.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	if err != nil {
		log.Fatalln(err)
	}

	zoneID, err := cf.ZoneIDByName(*zoneName)
	if err != nil {
		log.Fatalln(err)
	}

	recordFilter := cloudflare.DNSRecord{Name: *recordName}

	records, err := cf.DNSRecords(zoneID, recordFilter)
	if err != nil {
		log.Fatalln(err)
	}

	if len(records) == 0 {
		log.Fatalf("no records for zone %s found with name %s, aborting!\n", zoneID, *recordName)
	}

	record := records[0]

	log.Printf("current record IP: %s\n", record.Content)

	consensus := externalip.DefaultConsensus(nil, nil)

	log.Println("getting current public IP, just a sec...")

	extIP, err := consensus.ExternalIP()
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("current public IP: %s\n", extIP)

	if record.Content != extIP.String() {
		log.Printf("updating: %s -> %s\n", record.Content, extIP.String())

		update := cloudflare.DNSRecord{Content: extIP.String()}

		err := cf.UpdateDNSRecord(zoneID, record.ID, update)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("done!")
	} else {
		log.Println("Cloudflare is up-to-date, nothing to do...")
	}
}
