package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"slices"
	"text/tabwriter"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Config struct {
	StartCost      int
	EndCost        int
	Password       string
	GenerateLength int
	Iterations     int
}

type CostResult struct {
	Cost       int
	Durations  []time.Duration
	Mean       time.Duration
	StdDev     time.Duration
	P25        time.Duration
	P75        time.Duration
	P95        time.Duration
	P99        time.Duration
	Iterations int
}

func main() {
	cfg := parseFlags()

	password := resolvePassword(cfg)

	fmt.Println("Bcrypt Cost Benchmark")
	fmt.Println("=====================")
	fmt.Println()

	results := runBenchmark(cfg, password)

	printReport(cfg, password, results)
}

func parseFlags() Config {
	cfg := Config{}

	flag.IntVar(&cfg.StartCost, "start", 10, "Starting cost value")
	flag.IntVar(&cfg.EndCost, "end", 16, "Ending cost value")
	flag.StringVar(&cfg.Password, "password", "correct-horse-battery-staple", "Password to hash")
	flag.IntVar(&cfg.GenerateLength, "generate", 0, "Generate random password of given length (overrides -password)")
	flag.IntVar(&cfg.Iterations, "iterations", 3, "Number of iterations per cost level")

	flag.Parse()

	if cfg.StartCost < bcrypt.MinCost {
		log.Fatalf("Start cost must be at least %d", bcrypt.MinCost)
	}
	if cfg.EndCost > bcrypt.MaxCost {
		log.Fatalf("End cost must be at most %d", bcrypt.MaxCost)
	}
	if cfg.StartCost > cfg.EndCost {
		log.Fatal("Start cost must be less than or equal to end cost")
	}
	if cfg.Iterations < 1 {
		log.Fatal("Iterations must be at least 1")
	}

	return cfg
}

func resolvePassword(cfg Config) []byte {
	if cfg.GenerateLength > 0 {
		return generateRandomPassword(cfg.GenerateLength)
	}
	return []byte(cfg.Password)
}

func generateRandomPassword(length int) []byte {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatalf("Error generating random password: %v", err)
	}

	for i := range length {
		password[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	return password
}

func runBenchmark(cfg Config, password []byte) []CostResult {
	results := make([]CostResult, 0, cfg.EndCost-cfg.StartCost+1)
	spinnerIdx := 0

	for cost := cfg.StartCost; cost <= cfg.EndCost; cost++ {
		durations := make([]time.Duration, 0, cfg.Iterations)

		for iter := 1; iter <= cfg.Iterations; iter++ {
			spinnerIdx = (spinnerIdx + 1) % len(spinnerFrames)
			fmt.Printf("\r%s Running: cost=%d, iteration=%d/%d    ",
				spinnerFrames[spinnerIdx], cost, iter, cfg.Iterations)

			start := time.Now()
			_, err := bcrypt.GenerateFromPassword(password, cost)
			if err != nil {
				log.Fatalf("\nError generating hash: %v", err)
			}
			durations = append(durations, time.Since(start))
		}

		results = append(results, calculateStats(cost, durations))
	}

	fmt.Print("\r\033[K")

	return results
}

func calculateStats(cost int, durations []time.Duration) CostResult {
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	slices.Sort(sorted)

	mean := calculateMean(sorted)
	stdDev := calculateStdDev(sorted, mean)

	return CostResult{
		Cost:       cost,
		Durations:  durations,
		Mean:       mean,
		StdDev:     stdDev,
		P25:        calculatePercentile(sorted, 25),
		P75:        calculatePercentile(sorted, 75),
		P95:        calculatePercentile(sorted, 95),
		P99:        calculatePercentile(sorted, 99),
		Iterations: len(durations),
	}
}

func calculateMean(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func calculateStdDev(durations []time.Duration, mean time.Duration) time.Duration {
	if len(durations) < 2 {
		return 0
	}

	var sumSquares float64
	for _, d := range durations {
		diff := float64(d - mean)
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(durations)-1)
	return time.Duration(math.Sqrt(variance))
}

func calculatePercentile(sorted []time.Duration, percentile float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := (percentile / 100) * float64(len(sorted)-1)
	lower := int(rank)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	weight := rank - float64(lower)
	return time.Duration(float64(sorted[lower])*(1-weight) + float64(sorted[upper])*weight)
}

func printReport(cfg Config, password []byte, results []CostResult) {
	fmt.Println("Benchmark Configuration")
	fmt.Println("-----------------------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Cost Range:\t%d - %d\n", cfg.StartCost, cfg.EndCost)
	fmt.Fprintf(w, "Iterations:\t%d per cost level\n", cfg.Iterations)
	fmt.Fprintf(w, "Password Length:\t%d characters\n", len(password))
	if cfg.GenerateLength > 0 {
		fmt.Fprintf(w, "Password Source:\tGenerated (random)\n")
	} else {
		fmt.Fprintf(w, "Password Source:\tProvided\n")
	}
	w.Flush()

	fmt.Println()
	fmt.Println("Results")
	fmt.Println("-------")
	fmt.Println()

	w = tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Cost\tIterations\tMean\tStdDev\tP25\tP75\tP95\tP99\t")
	fmt.Fprintln(w, "----\t----------\t----\t------\t---\t---\t---\t---\t")

	for _, r := range results {
		fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
			r.Cost,
			r.Iterations,
			formatDuration(r.Mean),
			formatDuration(r.StdDev),
			formatDuration(r.P25),
			formatDuration(r.P75),
			formatDuration(r.P95),
			formatDuration(r.P99),
		)
	}
	w.Flush()

	fmt.Println()
	fmt.Println("Analysis")
	fmt.Println("--------")

	for _, r := range results {
		var recommendation string
		switch {
		case r.Mean < 100*time.Millisecond:
			recommendation = "Fast - consider higher cost for sensitive data"
		case r.Mean < 250*time.Millisecond:
			recommendation = "Good - balanced security and performance"
		case r.Mean < 500*time.Millisecond:
			recommendation = "Acceptable - may impact UX under load"
		case r.Mean < 1*time.Second:
			recommendation = "Slow - may cause timeouts under load"
		default:
			recommendation = "Too slow - not recommended for production"
		}
		fmt.Printf("  Cost %d: %s\n", r.Cost, recommendation)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
