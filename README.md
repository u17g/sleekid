# sleekid

The unique ID for the modern software. Stripe inspired.

![id format](./id_format.png)

```
Prefix := user defined string
Timestamp := base62(unix time - 2024-01-01), T: default is 5 digits.
Random := base62(N length of random digits), N: default is 12
Checksum := base62(C length of checksum digits), C: default is 2
Total length = Prefix + T + N + C
```

## Points

- Human-readable prefix
- More compact than UUID while maintaining collision resistance
  - when random digits length is 16, total length is near 26 bytes
  - shorter than UUID, the same collision resistance with UUID.
- Monotonically increasing
- B-tree optimized with built-in timestamps
- Configurable length
  - timestamp length: 4 ~ 6
  - checksum length: 1 ~
  - random digits length: 1 ~
- Built-in checksum verification
  - Detects errors with a probability of 99.97%
- Cryptographically random
- URL friendly with base62 encoding

## Usage

```go
import "github.com/RyosukeCla/sleekid"

// Setup generator, before using sleekid.
sleekid.Setup(sleekid.GeneratorInit{
  // Set this token on your production environment.
  // and keep it secret.
  ChecksumToken:      729823908348,

  // Optional
  RandomDigitsLength: 12,
  Delimiter:          '_',
  ChecksumLength:     2,
  TimestampLength:    5,
})

// Generate id.
id, err := sleekid.New("usr")

// Generate id with custom random digits length
id, err := sleekid.New("usr", sleekid.WithRandomDigits(27))

// Validate id.
sleekid.Validate(id)
sleekid.ValidateWithPrefix("usr", id)

// Get Prefix
prefix := sleekid.Prefix(id)

// Get Timestamp
timestamp := sleekid.Timestamp(id)

// Custom Generator
gen := sleekid.NewGenerator(sleekid.GeneratorInit{
  // ...
})
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

- Tamper detection probability is 99.97% when checksum length is 2 digits
  - where 1 - 1/62^2 ≈ 0.9997
- This helps to prevent brute-force attacks / DDoS / etc.

### Length Recommendations

- For small scale systems (<100K IDs/day): N=10
- For medium scale systems (<10M IDs/day): N=12
- For large scale systems (>10M IDs/day): N=14
- For extreme scale systems (>1B IDs/day): N=16

## Benchmark

```
$ go test -bench . -benchmem

goos: darwin
goarch: arm64
pkg: github.com/RyosukeCla/sleekid
cpu: Apple M1 Max
BenchmarkNew-10                          2757162               434.6 ns/op           120 B/op          6 allocs/op
BenchmarkPrefix-10                      50348416                23.00 ns/op           16 B/op          2 allocs/op
BenchmarkTimestamp-10                   24204141                48.88 ns/op            0 B/op          0 allocs/op
BenchmarkValidate-10                    32242606                36.41 ns/op            2 B/op          1 allocs/op
BenchmarkValidateWithPrefix-10          29806628                40.36 ns/op            2 B/op          1 allocs/op
BenchmarkNewUUID-10                      4148452               281.1 ns/op            16 B/op          1 allocs/op
BenchmarkNewXid-10                      24769122                47.97 ns/op            0 B/op          0 allocs/op
```
