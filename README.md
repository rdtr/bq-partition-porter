# BigQuery partitioned table exporter / importer

## Description

bq-partition-porter is a command line tool that helps exporting a table on BigQuery by each partition.  
Data on each partition will be stored into Google Cloud Storage, while separating directory per partition.

Also, you can load data from Google Cloud Storage if you keep directory structure like
```
gs://your_bucket/prefix/20170905/*.json <- each directory name corresponds to partition date
gs://your_bucket/prefix/20170906/*.json
gs://your_bucket/prefix/20170907/*.json
:
:
```

## Usage
```
Usage of bq-partition-porter:
  -d string
    	target dataset
  -e string
    	end date to import/export
  -g string
    	prefix of GCS source/destination
  -o string
    	operation to perform, either import or export
  -p string
    	target project
  -s string
    	start date to import/export
  -t string
    	target table name
```

`start date` and `end date` work in an inclusive manner.

Example:
```
./bq-partition-porter -p=my-gcp-project -d=my-dataset -t=my-table -s=2017-08-30 -e=2017-09-27 -g=gs://my-bucket/rescue -o=export
exporting my-dataset.my-table$20170830 succeeded
exporting my-dataset.my-table$20170831 succeeded
:
:
```

## Warning

BigQuery export has following limits: 1,000 exports per day, up to 10TB
So you can't export beyond this quota by using this tool.

Also, currently import function using following hard-coded desposition:
```
		importer.CreateDisposition = bigquery.CreateIfNeeded
		importer.WriteDisposition = bigquery.WriteTruncate
```

So the whole table is replaced. I recommend first you import to a temp table then
if the data looks OK, copy the temp table to the actual destination.