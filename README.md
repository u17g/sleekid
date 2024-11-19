# sleekid

The unique ID for the modern software. Stripe inspired.

![id format](./id_format.png)

```
Prefix := user defined string
Timestamp := base62(unix time - 2024-01-01), 4 ~ 6 digits, mostly 5 digits during 30 years.
Random := base62(N length of random digits)
Checksum := base62(2 length of checksum digits)
Total length ~= prefix + N + 8
```

## Points

- Human-readable prefix
- More compact than UUID while maintaining collision resistance
  - when random digits length is 16, total length is near 26 bytes
  - shorter than UUID, the same collision resistance with UUID.
- Monotonically increasing
- B-tree optimized with built-in timestamps
- Configurable length
- Built-in checksum verification
  - Detects errors with a probability of 99.74%
- Cryptographically random
- URL friendly with base62 encoding

## Usage

```go
import "github.com/RyosukeCla/sleekid"

// Setup generator, before using sleekid.
sleekid.Setup(sleekid.GeneratorInit{Token: 12345, RandomDigitsLength: 12})

// Generate id.
id, err := sleekid.New("usr")

// Generate id with custom random digits length
id, err := sleekid.New("usr", sleekid.WithRandomDigits(27))

// Validate id.
sleekid.Validate("usr", id)

// Custom Generator
gen := sleekid.NewGenerator(sleekid.GeneratorInit{Token: 12345, RandomDigitsLength: 12})
id, err := gen.New("usr")
gen.Validate("usr", id)
```

## Theory

### Space Size

timestamp part with 4 ~ 6 digits.
- Each character: 62 possibilities (0-9, a-z, A-Z)
- 6 digits: 62^6 ≈ 56.8 quadrillion combinations


random digits part with N length

- Each character: 62 possibilities (0-9, a-z, A-Z)
- N characters: 62^N possible combinations
- For N=12: 62^12 ≈ 3.22 × 10^21 combinations

total space
- Combined unique possibilities: 62^6 * 62^N = 62^(6+N)
- For N=12: 62^18 ≈ 1.83 × 10^32 combinations

### Birthday Problem

For a 50% collision probability:
- √(π/2 * 62^N) attempts needed
- For N=12: √(π/2 * 62^12) ≈ 2.8 trillion attempts

### Tamper Detection Probability

- Tamper detection probability is 99.74% when checksum length is 2 digits
  - where 1 - 1/62^2 ≈ 0.9974
- This helps to prevent brute-force attacks / DDoS / etc.

### Length Recommendations

- For small scale systems (<100K IDs/day): N=10
- For medium scale systems (<10M IDs/day): N=12
- For large scale systems (>10M IDs/day): N=14
- For extreme scale systems (>1B IDs/day): N=16

## Benchmark

```
goos: darwin
goarch: arm64
pkg: github.com/RyosukeCla/sleekid
cpu: Apple M1 Max
BenchmarkNew-10                  2806604               438.4 ns/op            88 B/op          6 allocs/op
BenchmarkValidate-10            32127796               37.45 ns/op             2 B/op          1 allocs/op
BenchmarkNewUUID-10              4314640               279.5 ns/op            16 B/op          1 allocs/op
```
