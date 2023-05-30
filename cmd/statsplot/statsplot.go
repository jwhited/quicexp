package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var (
	flagInputFile   = flag.String("input", "", "input filename, defaults to stdin if unspecified")
	flagOutputFile  = flag.String("output", "", "output filename")
	flagStat        = flag.String("stat", "Throughput(kbps)", "stat to plot")
	flagLabelFilter = flag.String("filter", "", "regular expression to filter labels by")
)

func main() {
	flag.Parse()
	p := plot.New()
	p.Title.Text = fmt.Sprintf("msquic %s via secnetperf", *flagStat)
	p.X.Label.Text = "Run #"
	p.Y.Label.Text = *flagStat
	p.X.Padding = 2 * vg.Inch
	p.Legend.YOffs = -1.5 * vg.Inch

	var ioReader = os.Stdin
	if len(*flagInputFile) > 0 {
		var err error
		ioReader, err = os.Open(*flagInputFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	reader := csv.NewReader(ioReader)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("error reading: %v", err)
	}

	if len(rows) < 1 {
		log.Fatal("no rows")
	}
	statIndex := -1
	for i, label := range rows[0] {
		if label == *flagStat {
			statIndex = i
		}
	}
	if statIndex == -1 {
		log.Fatalf("stat (%s) not found", *flagStat)
	}

	var re *regexp.Regexp
	if len(*flagLabelFilter) > 0 {
		re, err = regexp.Compile(*flagLabelFilter)
		if err != nil {
			log.Fatal(err)
		}
	}
	throughput := make(map[string][]int)
	for _, row := range rows[1:] {
		if re != nil && !re.MatchString(row[0]) {
			continue
		}
		val, err := strconv.Atoi(row[statIndex])
		if err != nil {
			log.Fatalf("error converting: %v", err)
		}
		vals, ok := throughput[row[0]]
		if !ok {
			vals = make([]int, 0)
		}
		vals = append(vals, val)
		throughput[row[0]] = vals
	}
	keysSorted := make([]string, 0)
	for k := range throughput {
		keysSorted = append(keysSorted, k)
	}
	sort.Strings(keysSorted)
	points := make([]interface{}, 0)
	for _, k := range keysSorted {
		vals, _ := throughput[k]
		pts := make(plotter.XYs, 0)
		for i, val := range vals {
			pts = append(pts, plotter.XY{X: float64(i + 1), Y: float64(val)})
		}
		points = append(points, k, pts)
	}
	err = plotutil.AddLinePoints(p, points...)
	if err != nil {
		log.Fatalf("error adding line points to plot: %v", err)
	}
	if len(*flagOutputFile) > 0 {
		err = p.Save(8*vg.Inch, 8*vg.Inch, *flagOutputFile)
		if err != nil {
			log.Fatalf("error saving png: %v", err)
		}
	} else {
		log.Println("no --output file specified, nothing to do")
	}
}
