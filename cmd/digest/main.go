// File: cmd/digest/main.go

// This program sends daily, weekly, monthly, and quarterly digests.
// It also does caching and cleanup jobs.

package main

import (
	"log"
	"os"

	"twothumbs/internal/config"
	"twothumbs/internal/cronjobs"
	"twothumbs/internal/digests"
	"twothumbs/internal/utils"
)

func main() {
	log.SetOutput(os.Stdout)
	cfg := config.LoadDigestConfig()

	// Connect to the database
	conn, err := utils.ConnectToDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Run daily cache job
	log.Println("Starting daily cache job...")
	if err := cronjobs.RunDailyCacheJob(conn, cfg); err != nil {
		log.Printf("Daily cache job failed: %v", err)
	} else {
		log.Println("Daily cache job completed.")
	}

	// *Run digest jobs*

	var digestErr error

	// Send daily digests
	log.Println("Sending daily digests...")
	if err := digests.SendDailyDigests(conn, cfg); err != nil {
		log.Printf("Failed to send daily digests: %v", err)
		digestErr = err
	} else {
		log.Println("Daily digests sent successfully.")
	}

	// Proceed with weekly, monthly, and quarterly digests
	if utils.IsMonday() {
		// Send weekly digests
		log.Println("Sending weekly digests...")
		if err := digests.SendWeeklyDigests(conn, cfg); err != nil {
			log.Printf("Failed to send weekly digests: %v", err)
			digestErr = err
		} else {
			log.Println("Weekly digests sent successfully.")
		}
	}

	if utils.IsFirstWeekdayOfMonth() {
		// Send monthly digests
		log.Println("Sending monthly digests...")
		if err := digests.SendMonthlyDigests(conn, cfg); err != nil {
			log.Printf("Failed to send monthly digests: %v", err)
			digestErr = err
		} else {
			log.Println("Monthly digests sent successfully.")
		}
	}

	if utils.IsSecondWeekdayOfQuarter() {
		// Send quarterly digests
		log.Println("Sending quarterly digests...")
		if err := digests.SendQuarterlyDigests(conn, cfg); err != nil {
			log.Printf("Failed to send quarterly digests: %v", err)
			digestErr = err
		} else {
			log.Println("Quarterly digests sent successfully.")
		}
	}

	// *Run maintenance jobs*

	// Run the cleanup job only after processing the digests
	if utils.IsFirstWeekdayOfMonth() && digestErr == nil {
		log.Println("Starting the cleanup job...")
		if err := cronjobs.RunCleanup(conn); err != nil {
			log.Printf("Cleanup job failed: %v", err)
		} else {
			log.Println("Cleanup job completed.")
		}
	}
}
