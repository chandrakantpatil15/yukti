package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	_ "github.com/lib/pq"
)

type ResourceIdentifier struct {
	InstanceID        string            `json:"instance_id"`
	PrivateIP         *string           `json:"private_ip"`
	PublicIP          *string           `json:"public_ip"`
	PrivateDNS        *string           `json:"private_dns"`
	PublicDNS         *string           `json:"public_dns"`
	Hostname          *string           `json:"hostname"`
	ApplicationName   *string           `json:"application_name"`
	ServiceName       *string           `json:"service_name"`
	CustomIdentifier  *string           `json:"custom_identifier"`
	Tags              map[string]string `json:"tags"`
}

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Failed to load AWS config:", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Sync resource identifiers
	if err := syncResourceIdentifiers(db, ec2Client); err != nil {
		log.Fatal("Failed to sync resource identifiers:", err)
	}

	fmt.Println("Resource identifiers synced successfully")
}

func syncResourceIdentifiers(db *sql.DB, ec2Client *ec2.Client) error {
	// Get all instances
	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return fmt.Errorf("failed to describe instances: %w", err)
	}

	var identifiers []ResourceIdentifier

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			identifier := ResourceIdentifier{
				InstanceID: *instance.InstanceId,
				Tags:       make(map[string]string),
			}

			// Network information
			if instance.PrivateIpAddress != nil {
				identifier.PrivateIP = instance.PrivateIpAddress
			}
			if instance.PublicIpAddress != nil {
				identifier.PublicIP = instance.PublicIpAddress
			}
			if instance.PrivateDnsName != nil {
				identifier.PrivateDNS = instance.PrivateDnsName
				identifier.Hostname = instance.PrivateDnsName // Use private DNS as hostname
			}
			if instance.PublicDnsName != nil {
				identifier.PublicDNS = instance.PublicDnsName
			}

			// Extract tags
			for _, tag := range instance.Tags {
				if tag.Key != nil && tag.Value != nil {
					identifier.Tags[*tag.Key] = *tag.Value
				}
			}

			// Application and service names from tags
			if appName, exists := identifier.Tags["Application"]; exists {
				identifier.ApplicationName = &appName
			}
			if serviceName, exists := identifier.Tags["Service"]; exists {
				identifier.ServiceName = &serviceName
			}

			// Custom identifier (prefer Name tag, fallback to instance ID)
			if name, exists := identifier.Tags["Name"]; exists {
				identifier.CustomIdentifier = &name
			} else {
				identifier.CustomIdentifier = &identifier.InstanceID
			}

			identifiers = append(identifiers, identifier)
		}
	}

	// Update database
	return updateResourceIdentifiersInDB(db, identifiers)
}

func updateResourceIdentifiersInDB(db *sql.DB, identifiers []ResourceIdentifier) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First update yt_aws_resources with network info
	updateResourceStmt := `
		UPDATE yt_aws_resources 
		SET private_ip = $2, public_ip = $3, private_dns = $4, public_dns = $5
		WHERE instance_id = $1`

	// Then upsert yt_resource_identifiers
	upsertIdentifierStmt := `
		INSERT INTO yt_resource_identifiers 
		(instance_id, private_ip, public_ip, hostname, application_name, service_name, custom_identifier)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (instance_id) DO UPDATE SET
			private_ip = EXCLUDED.private_ip,
			public_ip = EXCLUDED.public_ip,
			hostname = EXCLUDED.hostname,
			application_name = EXCLUDED.application_name,
			service_name = EXCLUDED.service_name,
			custom_identifier = EXCLUDED.custom_identifier,
			updated_at = NOW()`

	for _, id := range identifiers {
		// Update main resources table
		_, err := tx.Exec(updateResourceStmt, 
			id.InstanceID, id.PrivateIP, id.PublicIP, id.PrivateDNS, id.PublicDNS)
		if err != nil {
			return fmt.Errorf("failed to update resource %s: %w", id.InstanceID, err)
		}

		// Upsert identifiers table
		_, err = tx.Exec(upsertIdentifierStmt,
			id.InstanceID, id.PrivateIP, id.PublicIP, id.Hostname,
			id.ApplicationName, id.ServiceName, id.CustomIdentifier)
		if err != nil {
			return fmt.Errorf("failed to upsert identifier %s: %w", id.InstanceID, err)
		}

		fmt.Printf("Updated identifiers for %s (IP: %v, Name: %v)\n", 
			id.InstanceID, 
			getStringValue(id.PrivateIP), 
			getStringValue(id.CustomIdentifier))
	}

	return tx.Commit()
}

func getStringValue(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}