package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

func runImport(ctx context.Context, bqc *bigquery.Client, startTime, endTime time.Time, dataset, table, gcsPrefix string) error {
	ptDateLayout := "20060102"

	for targetTime := startTime; targetTime.Equal(endTime) || targetTime.Before(endTime); targetTime = targetTime.AddDate(0, 0, 1) {
		ptStr := targetTime.Format(ptDateLayout)
		gcsRef := bigquery.NewGCSReference(fmt.Sprintf("%s/%s/*", gcsPrefix, ptStr))
		gcsRef.SourceFormat = bigquery.JSON

		importer := bqc.Dataset(dataset).Table(fmt.Sprintf("%s$%s", table, ptStr)).LoaderFrom(gcsRef)
		importer.CreateDisposition = bigquery.CreateIfNeeded
		importer.WriteDisposition = bigquery.WriteTruncate

		job, err := importer.Run(ctx)
		if err != nil {
			return fmt.Errorf("could not run a job to import to %s.%s$%s, %s\n", dataset, table, ptStr, err.Error())
		}

		status, err := job.Wait(ctx)
		if err != nil {
			return fmt.Errorf("failed when waiting a job to import to %s.%s$%s, %s\n", dataset, table, ptStr, err.Error())
		}
		if err := status.Err(); err != nil {
			return fmt.Errorf("failed when running a job to import to %s.%s$%s, %s\n", dataset, table, ptStr, err.Error())
		}

		fmt.Printf("importing %s/%s/* to %s.%s$%s succeeded\n", gcsPrefix, ptStr, dataset, table, ptStr)
	}
	return nil
}
