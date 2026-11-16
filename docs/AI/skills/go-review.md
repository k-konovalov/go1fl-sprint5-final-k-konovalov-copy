# Go Code Review Guide

A code review checklist based on the official Go guidelines, Effective Go, and community best practices.

## Quick Review Checklist

### Must-check items
- [ ] Are errors handled properly (not ignored, with context)
- [ ] Do goroutines have an exit mechanism (to avoid leaks)
- [ ] Are contexts passed and canceled correctly
- [ ] Is the receiver type chosen sensibly (value/pointer)
- [ ] Is the code formatted with `gofmt`

### Frequent issues
- [ ] Loop variable capture issues (Go < 1.22)
- [ ] Are nil checks complete
- [ ] Is a map initialized before use
- [ ] Using defer inside loops
- [ ] Variable shadowing

---

## 1. Error Handling

> 📖 For general principles and cross-language examples see [Cross-language Error Handling Guide](cross-cutting/error-handling-principles.md)

### 1.1 Never Ignore Errors

```go
// ❌ Wrong: ignoring the error
result, _ := SomeFunction()

// ✅ Correct: handling the error
result, err := SomeFunction()
if err != nil {
    return fmt.Errorf("some function failed: %w", err)
}
```

### 1.2 Error Wrapping and Context

```go
// ❌ Wrong: losing context
if err != nil {
    return err
}

// ❌ Wrong: using %v loses the error chain
if err != nil {
    return fmt.Errorf("failed: %v", err)
}

// ✅ Correct: using %w preserves the error chain
if err != nil {
    return fmt.Errorf("failed to process user %d: %w", userID, err)
}
```

### 1.3 Using errors.Is and errors.As

```go
// ❌ Wrong: direct comparison (cannot handle wrapped errors)
if err == sql.ErrNoRows {
    // ...
}

// ✅ Correct: use errors.Is (supports the error chain)
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}

// ✅ Correct: use errors.As to extract a specific type
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    log.Printf("path error: %s", pathErr.Path)
}
```

### 1.4 Custom Error Types

```go
// ✅ Recommended: define sentinel errors
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
)

// ✅ Recommended: custom error with context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}
```

### 1.5 Handle an Error Only Once

```go
// ❌ Wrong: both logging and returning (double handling)
if err != nil {
    log.Printf("error: %v", err)
    return err
}

// ✅ Correct: only return and let the caller decide
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// ✅ Or: only log and handle (do not return)
if err != nil {
    log.Printf("non-critical error: %v", err)
    // continue with fallback logic
}
```

---

## 2. Concurrency and Goroutines

> 📖 For general concurrency patterns and cross-language examples see [Cross-language Async and Concurrency Guide](cross-cutting/async-concurrency-patterns.md)

### 2.1 Avoid Goroutine Leaks

```go
// ❌ Wrong: the goroutine can never exit
func bad() {
    ch := make(chan int)
    go func() {
        val := <-ch // blocks forever, nobody sends
        fmt.Println(val)
    }()
    // the function returns, the goroutine leaks
}

// ✅ Correct: use a context or a done channel
func good(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return // exit gracefully
        }
    }()
}
```

### 2.2 Channel Usage Rules

```go
// ❌ Wrong: sending to a nil channel (blocks forever)
var ch chan int
ch <- 1 // blocks forever

// ❌ Wrong: sending to a closed channel (panic)
close(ch)
ch <- 1 // panic!

// ✅ Correct: the sender closes the channel
func producer(ch chan<- int) {
    defer close(ch) // the sender is responsible for closing
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

// ✅ Correct: the receiver detects closure
for val := range ch {
    process(val)
}
// or
val, ok := <-ch
if !ok {
    // the channel is closed
}
```

### 2.3 Using sync.WaitGroup

```go
// ❌ Wrong: Add called inside the goroutine
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    go func() {
        wg.Add(1) // race condition!
        defer wg.Done()
        work()
    }()
}
wg.Wait()

// ✅ Correct: Add called before starting the goroutine
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        work()
    }()
}
wg.Wait()
```

### 2.4 Avoid Capturing Variables in Loops (Go < 1.22)

```go
// ❌ Wrong (Go < 1.22): capturing the loop variable
for _, item := range items {
    go func() {
        process(item) // all goroutines may use the same item
    }()
}

// ✅ Correct: pass the argument
for _, item := range items {
    go func(it Item) {
        process(it)
    }(item)
}

// ✅ Go 1.22+: the default behavior is fixed, a new variable is created per iteration
```

### 2.5 Worker Pool Pattern

```go
// ✅ Recommended: limit the concurrency level
func processWithWorkerPool(ctx context.Context, items []Item, workers int) error {
    jobs := make(chan Item, len(items))
    results := make(chan error, len(items))

    // start the workers
    for w := 0; w < workers; w++ {
        go func() {
            for item := range jobs {
                results <- process(item)
            }
        }()
    }

    // send the jobs
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    // collect the results
    for range items {
        if err := <-results; err != nil {
            return err
        }
    }
    return nil
}
```

---

## 3. Using Context

### 3.1 Context as the First Parameter

```go
// ❌ Wrong: context is not the first parameter
func Process(data []byte, ctx context.Context) error

// ❌ Wrong: context stored in a struct
type Service struct {
    ctx context.Context // do not do this!
}

// ✅ Correct: context as the first parameter, named ctx
func Process(ctx context.Context, data []byte) error
```

### 3.2 Propagate Instead of Creating a New Root Context

```go
// ❌ Wrong: creating a new root context in the call chain
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.Background() // the request context is lost!
        process(ctx)
        next.ServeHTTP(w, r)
    })
}

// ✅ Correct: take it from the request and propagate it
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        ctx = context.WithValue(ctx, key, value)
        process(ctx)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 3.3 Always Call the cancel Function

```go
// ❌ Wrong: cancel is not called
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
// the cancel() call is missing, resources may leak

// ✅ Correct: use defer to guarantee the call
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel() // must be called even if the timeout expires
```

### 3.4 Respond to Context Cancellation

```go
// ✅ Recommended: check the context during long-running operations
func LongRunningTask(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // returns context.Canceled or context.DeadlineExceeded
        default:
            // do a small chunk of work
            if err := doChunk(); err != nil {
                return err
            }
        }
    }
}
```

### 3.5 Distinguish Cancellation Reasons

```go
// ✅ Distinguish the cancellation reason via ctx.Err()
if err := ctx.Err(); err != nil {
    switch {
    case errors.Is(err, context.Canceled):
        log.Println("operation was canceled")
    case errors.Is(err, context.DeadlineExceeded):
        log.Println("operation timed out")
    }
    return err
}
```

---

## 4. Interface Design

### 4.1 Accept Interfaces, Return Structs

```go
// ❌ Not recommended: accepting a concrete type
func SaveUser(db *sql.DB, user User) error

// ✅ Recommended: accept an interface (decoupled, easy to test)
type UserStore interface {
    Save(ctx context.Context, user User) error
}

func SaveUser(store UserStore, user User) error

// ❌ Not recommended: returning an interface
func NewUserService() UserServiceInterface

// ✅ Recommended: return a concrete type
func NewUserService(store UserStore) *UserService
```

### 4.2 Define Interfaces at the Consumer Side

```go
// ❌ Not recommended: defining the interface in the implementation package
// package database
type Database interface {
    Query(ctx context.Context, query string) ([]Row, error)
    // ... 20 methods
}

// ✅ Recommended: define the minimal required interface in the consumer package
// package userservice
type UserQuerier interface {
    QueryUsers(ctx context.Context, filter Filter) ([]User, error)
}
```

### 4.3 Keep Interfaces Small and Focused

```go
// ❌ Not recommended: a large, all-in-one interface
type Repository interface {
    GetUser(id int) (*User, error)
    CreateUser(u *User) error
    UpdateUser(u *User) error
    DeleteUser(id int) error
    GetOrder(id int) (*Order, error)
    CreateOrder(o *Order) error
    // ... more methods
}

// ✅ Recommended: small, focused interfaces
type UserReader interface {
    GetUser(ctx context.Context, id int) (*User, error)
}

type UserWriter interface {
    CreateUser(ctx context.Context, u *User) error
    UpdateUser(ctx context.Context, u *User) error
}

// composed interface
type UserRepository interface {
    UserReader
    UserWriter
}
```

### 4.4 Avoid Overusing Empty Interfaces

```go
// ❌ Not recommended: excessive use of interface{}
func Process(data interface{}) interface{}

// ✅ Recommended: use generics (Go 1.18+)
func Process[T any](data T) T

// ✅ Recommended: define a concrete interface
type Processor interface {
    Process() Result
}
```

---

## 5. Choosing Receiver Types

### 5.1 When to Use Pointer Receivers

```go
// ✅ When the receiver needs to be modified
func (u *User) SetName(name string) {
    u.Name = name
}

// ✅ The receiver contains synchronization primitives such as sync.Mutex
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// ✅ The receiver is a large struct (avoid copy overhead)
type LargeStruct struct {
    Data [1024]byte
    // ...
}

func (l *LargeStruct) Process() { /* ... */ }
```

### 5.2 When to Use Value Receivers

```go
// ✅ The receiver is a small immutable struct
type Point struct {
    X, Y float64
}

func (p Point) Distance(other Point) float64 {
    return math.Sqrt(math.Pow(p.X-other.X, 2) + math.Pow(p.Y-other.Y, 2))
}

// ✅ The receiver is an alias of a basic type
type Counter int

func (c Counter) String() string {
    return fmt.Sprintf("%d", c)
}

// ✅ The receiver is a map, func, or chan (they are reference types already)
type StringSet map[string]struct{}

func (s StringSet) Contains(key string) bool {
    _, ok := s[key]
    return ok
}
```

### 5.3 Consistency Rule

```go
// ❌ Not recommended: mixing receiver types
func (u User) GetName() string   // value receiver
func (u *User) SetName(n string) // pointer receiver

// ✅ Recommended: if any method needs a pointer receiver, use pointers for all
func (u *User) GetName() string { return u.Name }
func (u *User) SetName(n string) { u.Name = n }
```

---

## 6. Performance Optimization

### 6.1 Preallocate Slices

```go
// ❌ Not recommended: dynamic growth
var result []int
for i := 0; i < 10000; i++ {
    result = append(result, i) // multiple allocations and copies
}

// ✅ Recommended: preallocate the known size
result := make([]int, 0, 10000)
for i := 0; i < 10000; i++ {
    result = append(result, i)
}

// ✅ Or initialize directly
result := make([]int, 10000)
for i := 0; i < 10000; i++ {
    result[i] = i
}
```

### 6.2 Avoid Unnecessary Heap Allocations

```go
// ❌ May escape to the heap
func NewUser() *User {
    return &User{} // escapes to the heap
}

// ✅ Consider returning a value (if applicable)
func NewUser() User {
    return User{} // may be allocated on the stack
}

// Check escape analysis
// go build -gcflags '-m -m' ./...
```

### 6.3 Reuse Objects with sync.Pool

```go
// ✅ Recommended: use sync.Pool for frequently created/destroyed objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.Write(data)
    return buf.String()
}
```

### 6.4 String Concatenation Optimization

```go
// ❌ Not recommended: using + in a loop
var result string
for _, s := range strings {
    result += s // creates a new string every time
}

// ✅ Recommended: use strings.Builder
var builder strings.Builder
for _, s := range strings {
    builder.WriteString(s)
}
result := builder.String()

// ✅ Or use strings.Join
result := strings.Join(strings, "")
```

### 6.5 Avoid interface{} Conversion Overhead

```go
// ❌ Using interface{} on a hot path
func process(data interface{}) {
    switch v := data.(type) { // type assertion has overhead
    case int:
        // ...
    }
}

// ✅ Use generics or concrete types on a hot path
func process[T int | int64 | float64](data T) {
    // the type is resolved at compile time, no runtime overhead
}
```

---

## 7. Testing

### 7.1 Table-Driven Tests

```go
// ✅ Recommended: table-driven tests
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 1, 2, 3},
        {"with zero", 0, 5, 5},
        {"negative numbers", -1, -2, -3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### 7.2 Parallel Tests

```go
// ✅ Recommended: run independent test cases in parallel
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"test1", "input1"},
        {"test2", "input2"},
    }

    for _, tt := range tests {
        tt := tt // Go < 1.22 requires a copy
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // mark as parallelizable
            result := Process(tt.input)
            // assertions...
        })
    }
}
```

### 7.3 Mocking with Interfaces

```go
// ✅ Define an interface for testing
type EmailSender interface {
    Send(to, subject, body string) error
}

// Production implementation
type SMTPSender struct { /* ... */ }

// Test mock
type MockEmailSender struct {
    SendFunc func(to, subject, body string) error
}

func (m *MockEmailSender) Send(to, subject, body string) error {
    return m.SendFunc(to, subject, body)
}

func TestUserRegistration(t *testing.T) {
    mock := &MockEmailSender{
        SendFunc: func(to, subject, body string) error {
            if to != "test@example.com" {
                t.Errorf("unexpected recipient: %s", to)
            }
            return nil
        },
    }

    service := NewUserService(mock)
    // test...
}
```

### 7.4 Test Helper Functions

```go
// ✅ Mark helper functions with t.Helper()
func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper() // reports the caller's location on failure
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

// ✅ Clean up resources with t.Cleanup()
func TestWithTempFile(t *testing.T) {
    f, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() {
        os.Remove(f.Name())
    })
    // test...
}
```

---

## 8. Common Pitfalls

### 8.1 Nil Slice vs Empty Slice

```go
var nilSlice []int     // nil, len=0, cap=0
emptySlice := []int{}  // not nil, len=0, cap=0
made := make([]int, 0) // not nil, len=0, cap=0

// ✅ JSON encoding difference
json.Marshal(nilSlice)   // null
json.Marshal(emptySlice) // []

// ✅ Recommended: initialize explicitly when an empty JSON array is required
if slice == nil {
    slice = []int{}
}
```

### 8.2 Map Initialization

```go
// ❌ Wrong: uninitialized map
var m map[string]int
m["key"] = 1 // panic: assignment to entry in nil map

// ✅ Correct: initialize with make
m := make(map[string]int)
m["key"] = 1

// ✅ Or use a literal
m := map[string]int{}
```

### 8.3 Defer Inside Loops

```go
// ❌ Potential issue: defer runs only when the function returns
func processFiles(files []string) error {
    for _, file := range files {
        f, err := os.Open(file)
        if err != nil {
            return err
        }
        defer f.Close() // all files are closed only when the function returns!
        // process...
    }
    return nil
}

// ✅ Correct: use a closure or extract a function
func processFiles(files []string) error {
    for _, file := range files {
        if err := processFile(file); err != nil {
            return err
        }
    }
    return nil
}

func processFile(file string) error {
    f, err := os.Open(file)
    if err != nil {
        return err
    }
    defer f.Close()
    // process...
    return nil
}
```

### 8.4 Slices Share the Underlying Array

```go
// ❌ Potential issue: slices share the underlying array
original := []int{1, 2, 3, 4, 5}
slice := original[1:3] // [2, 3]
slice[0] = 100         // modifies original!
// original becomes [1, 100, 3, 4, 5]

// ✅ Correct: copy explicitly when an independent copy is required
slice := make([]int, 2)
copy(slice, original[1:3])
slice[0] = 100 // does not affect original
```

### 8.5 String Substring Memory Leak

```go
// ❌ Potential issue: the substring holds the entire underlying array
func getPrefix(s string) string {
    return s[:10] // still references the whole underlying array of s
}

// ✅ Correct: create an independent copy (Go 1.18+)
func getPrefix(s string) string {
    return strings.Clone(s[:10])
}

// ✅ Before Go 1.18
func getPrefix(s string) string {
    return string([]byte(s[:10]))
}
```

### 8.6 Interface Nil Pitfall

```go
// ❌ Pitfall: nil check on an interface
type MyError struct{}
func (e *MyError) Error() string { return "error" }

func returnsError() error {
    var e *MyError = nil
    return e // the returned error is not nil!
}

func main() {
    err := returnsError()
    if err != nil { // true! interface{type: *MyError, value: nil}
        fmt.Println("error:", err)
    }
}

// ✅ Correct: return nil explicitly
func returnsError() error {
    var e *MyError = nil
    if e == nil {
        return nil // return nil explicitly
    }
    return e
}
```

### 8.7 Comparing Times

```go
// ❌ Not recommended: comparing time.Time with ==
if t1 == t2 { // may fail because of monotonic clock differences
    // ...
}

// ✅ Recommended: use the Equal method
if t1.Equal(t2) {
    // ...
}

// ✅ Compare time ranges
if t1.Before(t2) || t1.After(t2) {
    // ...
}
```

---

## 9. Code Organization

### 9.1 Package Naming

```go
// ❌ Not recommended
package common   // too broad
package utils    // too broad
package helpers  // too broad
package models   // grouped by type

// ✅ Recommended: name by functionality
package user     // user-related functionality
package order    // order-related functionality
package postgres // PostgreSQL implementation
```

### 9.2 Avoid Import Cycles

```go
// ❌ Import cycle
// package a imports package b
// package b imports package a

// ✅ Solution 1: move shared types into a separate package
// package types (shared types)
// package a imports types
// package b imports types

// ✅ Solution 2: decouple with interfaces
// package a defines the interface
// package b implements the interface
```

### 9.3 Exported Identifier Rules

```go
// ✅ Export only the necessary identifiers
type UserService struct {
    db *sql.DB // private
}

func (s *UserService) GetUser(id int) (*User, error) // public
func (s *UserService) validate(u *User) error         // private

// ✅ Internal packages restrict access
// internal/database/... can only be imported by code in the same project
```

---

## 10. Tools and Checks

### 10.1 Required Tools

```bash
# Formatting (mandatory)
gofmt -w .
goimports -w .

# Static analysis
go vet ./...

# Race detection
go test -race ./...

# Escape analysis
go build -gcflags '-m -m' ./...
```

### 10.2 Recommended Linters

```bash
# golangci-lint (bundles multiple linters)
golangci-lint run

# Common checks
# - errcheck: checks for unhandled errors
# - gosec: security checks
# - ineffassign: ineffective assignments
# - staticcheck: static analysis
# - unused: unused code
```

### 10.3 Benchmark Tests

```go
// ✅ Performance benchmarks
func BenchmarkProcess(b *testing.B) {
    data := prepareData()
    b.ResetTimer() // reset the timer

    for i := 0; i < b.N; i++ {
        Process(data)
    }
}

// Run the benchmark
// go test -bench=. -benchmem ./...
```

---

## Reference Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Go Common Mistakes](https://go.dev/wiki/CommonMistakes)
- [100 Go Mistakes](https://100go.co/)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
