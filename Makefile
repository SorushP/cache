PACKAGES := $(shell go list ./pkg/...)

.PHONY: all test coverage benchmark benchmark-cpu benchmark-mem pprof-cpu pprof-mem clean

all: test coverage benchmark

test:
	go test $(PACKAGES) -v -count=1

coverage:
	go test $(PACKAGES) -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html

benchmark: benchmark-cpu benchmark-mem

benchmark-cpu:
	@for pkg in $(PACKAGES); do \
		echo "Running benchmarks for $$pkg with CPU profiling..."; \
		go test $$pkg -count=1 -bench=. -benchmem -cpuprofile=cpu.prof; \
	done

benchmark-mem:
	@for pkg in $(PACKAGES); do \
		echo "Running benchmarks for $$pkg with memory profiling..."; \
		go test $$pkg -count=1 -bench=. -benchmem -memprofile=mem.prof; \
	done

# Serve cpu.prof file with pprof
pprof-cpu:
	@if [ ! -z "$(shell find . -name "cpu.prof")" ]; then \
		go tool pprof -top cpu.prof; \
	else \
		echo "No profile found."; \
	fi

# Serve mem.prof file with pprof
pprof-mem:
	@if [ ! -z "$(shell find . -name "mem.prof")" ]; then \
		go tool pprof -top mem.prof; \
	else \
		echo "No profile found."; \
	fi

# Clean up any temporary files and profiles
clean:
	go clean ./...
	rm -f cpu.prof mem.prof coverage.out coverage.html
