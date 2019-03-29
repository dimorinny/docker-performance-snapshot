package main

import (
	"bufio"
	"fmt"
	"io"
)

type CsvReporter struct {
	directoryPath string
}

func NewCsvReporter(directoryPath string) Reporter {
	return &CsvReporter{
		directoryPath: directoryPath,
	}
}

func (r *CsvReporter) Report(name string, description string, metrics []*Metric) error {
	csvFile, err := createFile(
		r.directoryPath,
		fmt.Sprintf("%s.csv", name),
	)
	if err != nil {
		return err
	}
	defer func() { _ = csvFile.Close() }()

	scvMetricsWriter := bufio.NewWriter(csvFile)

	for _, metric := range metrics {
		err := r.writeCsvMetricsEntry(
			scvMetricsWriter,
			metric.Timestamp,
			metric.Value,
		)
		if err != nil {
			return err
		}
	}

	err = scvMetricsWriter.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (r *CsvReporter) writeCsvMetricsEntry(writer io.Writer, timestamp int64, value float64) error {
	_, err := fmt.Fprintln(
		writer,
		fmt.Sprintf(
			"%d;%f",
			timestamp,
			value,
		),
	)

	return err
}
