package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

var operation string
var startDate, endDate string
var projectID, dataset, table string
var gcsPrefix string

var startTime, endTime time.Time

func init() {
	flag.StringVar(&operation, "o", "", "operation to perform, either import or export")
	flag.StringVar(&startDate, "s", "", "start date to import/export, YYYY-MM-DD format")
	flag.StringVar(&endDate, "e", "", "end date to import/export, YYYY-MM-DD format")
	flag.StringVar(&projectID, "p", "", "target project")
	flag.StringVar(&dataset, "d", "", "target dataset")
	flag.StringVar(&table, "t", "", "target table name")
	flag.StringVar(&gcsPrefix, "g", "", "prefix of GCS source/destination")
}

func main() {
	flag.Parse()

	if err := checkFlags(operation, startDate, endDate, dataset, table, gcsPrefix); err != nil {
		fmt.Printf(err.Error() + "\n")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	bqc, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		fmt.Printf("could not initialize BigQuery client, %s\n", err.Error())
		os.Exit(1)
	}

	if err := run(ctx, bqc, startTime, endTime, dataset, table, gcsPrefix); err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, bqc *bigquery.Client, startTime, endTime time.Time, dataset, table, gcsPrefix string) error {
	if err := checkDataset(ctx, bqc, dataset); err != nil {
		return err
	}
	if err := checkTable(ctx, bqc, dataset, table); err != nil {
		return err
	}

	if operation == "import" {
		return runImport(ctx, bqc, startTime, endTime, dataset, table, gcsPrefix)
	}
	return runExport(ctx, bqc, startTime, endTime, dataset, table, gcsPrefix)
}

func checkDataset(ctx context.Context, bqc *bigquery.Client, dataset string) error {
	it := bqc.Datasets(ctx)
	found := false
	for {
		cur, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		if cur.DatasetID == dataset {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("dataset %s not found\n", dataset)
	}
	return nil
}

func checkTable(ctx context.Context, bqc *bigquery.Client, dataset, table string) error {
	it := bqc.Dataset(dataset).Tables(ctx)
	found := false
	for {
		cur, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		if cur.TableID == table {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("table %s not found\n", table)
	}
	return nil
}

func checkFlags(operation, startDate, endDate, dataset, table, gcsPrefix string) error {
	if operation != "import" && operation != "export" {
		return fmt.Errorf("invalid operation")
	}

	dateLayout := "2006-01-02"
	var err error
	if startTime, err = time.Parse(dateLayout, startDate); err != nil {
		return fmt.Errorf("invalid start date")
	}
	if endTime, err = time.Parse(dateLayout, endDate); err != nil {
		return fmt.Errorf("invalid end date")
	}
	if startTime.After(endTime) {
		return fmt.Errorf("the start date you entered is after the end date")
	}

	if projectID == "" {
		return fmt.Errorf("invalid project")
	}

	if dataset == "" || table == "" {
		return fmt.Errorf("invalid BigQuery dataset of table name")
	}

	gcsScheme := "gs://"
	if len(gcsPrefix) < len(gcsScheme) || gcsPrefix[:len(gcsScheme)] != gcsScheme {
		return fmt.Errorf("gcs prefix must start with gs://")
	}

	// trim if '/' is at the last of gcsPrefix
	gcsPrefix = strings.TrimRight(gcsPrefix, "/")
	return nil
}
