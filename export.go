package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

func runExport(ctx context.Context, bqc *bigquery.Client, startTime, endTime time.Time, dataset, table, gcsPrefix string) error {
	ptDateLayout := "20060102"

	for targetTime := startTime; targetTime.Equal(endTime) || targetTime.Before(endTime); targetTime = targetTime.AddDate(0, 0, 1) {
		ptStr := targetTime.Format(ptDateLayout)
		gcsRef := bigquery.NewGCSReference(fmt.Sprintf("%s/%s/export-*.json", gcsPrefix, ptStr))
		gcsRef.DestinationFormat = bigquery.JSON

		exporter := bqc.Dataset(dataset).Table(fmt.Sprintf("%s$%s", table, ptStr)).ExtractorTo(gcsRef)

		job, err := exporter.Run(ctx)
		if err != nil {
			return fmt.Errorf("could not run a job to export %s.%s$%s, %s\n", dataset, table, ptStr, err.Error())
		}

		status, err := job.Wait(ctx)
		if err != nil {
			return fmt.Errorf("failed when waiting a job to export %s.%s$%s, %s\n", dataset, table, ptStr, err.Error())
		}
		if err := status.Err(); err != nil {
			return fmt.Errorf("failed when running a job to export %s.%s$%s, %s\n", dataset, table, ptStr, err.Error())
		}

		fmt.Printf("exporting %s.%s$%s succeeded\n", dataset, table, ptStr)
	}
	return nil
}
