# Bcrypt Cost Benchmark Utility

## About

This utility benchmarks the performance of the bcrypt password hashing algorithm at various cost levels. It helps you determine the optimal bcrypt cost setting for your environment by measuring the time required to hash passwords at different cost factors. The tool provides statistical analysis (mean, standard deviation, percentiles) for each cost level, making it easier to balance security and performance for your application.

## Install

```
go install github.com/eldad/bcryptbenchmark@latest
```

## How to Compile from Source

You need Go installed on your system.

```
go build -o bcrypt-benchmark ./cmd/benchmark
```

This will produce a `bcrypt-benchmark` executable in your current directory.

## How to Use

Run the benchmark tool from your terminal:

```
./bcrypt-benchmark [options]
```

Example:

```
./bcrypt-benchmark -start 10 -end 14 -iterations 5 -generate 16
```

This runs the benchmark for bcrypt costs 10 through 14, using a randomly generated 16-character password, with 5 iterations per cost level.

## Command Options

- `-start <int>`
  - Starting bcrypt cost value (default: 10, minimum: 4)
- `-end <int>`
  - Ending bcrypt cost value (default: 16, maximum: 31)
- `-password <string>`
  - Password to hash (default: "correct-horse-battery-staple")
- `-generate <int>`
  - Generate a random password of the given length (overrides `-password` if set)
- `-iterations <int>`
  - Number of iterations per cost level (default: 3, minimum: 1)

## Output

The tool prints a table of results for each cost level, including:
- Mean hashing time
- Standard deviation
- 25th, 75th, 95th, and 99th percentiles

It also provides a recommendation for each cost level based on the measured mean time.
