#!/usr/bin/env python3

import requests
import time
import statistics
import json
from datetime import datetime, timedelta
import concurrent.futures

BASE_URL = "http://localhost:8085/api/v1"

# Login and get token
def get_auth_token():
    login_data = {
        "username": "admin",
        "password": "admin"
    }
    resp = requests.post(f"{BASE_URL}/auth/login", json=login_data)
    if resp.status_code == 200:
        return resp.json().get("token")
    return None

# Benchmark function
def benchmark_query(query_fn, name, iterations=100):
    times = []
    errors = 0
    
    for i in range(iterations):
        start = time.time()
        try:
            result = query_fn()
            elapsed = (time.time() - start) * 1000  # Convert to milliseconds
            times.append(elapsed)
        except Exception as e:
            errors += 1
    
    if times:
        return {
            'name': name,
            'avg': statistics.mean(times),
            'min': min(times),
            'max': max(times),
            'median': statistics.median(times),
            'p95': statistics.quantiles(times, n=20)[18] if len(times) > 20 else max(times),
            'p99': statistics.quantiles(times, n=100)[98] if len(times) > 100 else max(times),
            'samples': len(times),
            'errors': errors
        }
    return None

# Parallel benchmark
def parallel_benchmark(query_fn, name, threads=10, iterations=100):
    with concurrent.futures.ThreadPoolExecutor(max_workers=threads) as executor:
        # Submit all tasks
        futures = []
        start_time = time.time()
        
        for _ in range(iterations):
            futures.append(executor.submit(query_fn))
        
        # Wait for completion
        concurrent.futures.wait(futures)
        total_time = time.time() - start_time
        
        # Calculate metrics
        successful = sum(1 for f in futures if not f.exception())
        qps = successful / total_time
        
        return {
            'name': name,
            'threads': threads,
            'total_queries': iterations,
            'successful': successful,
            'failed': iterations - successful,
            'total_time': total_time,
            'qps': qps
        }

token = get_auth_token()
headers = {"Authorization": f"Bearer {token}"}

# Define test queries
def query_by_id():
    resp = requests.get(f"{BASE_URL}/entities/get?id=admin", headers=headers)
    return resp.json()

def query_by_tag():
    resp = requests.get(f"{BASE_URL}/entities/list?tag=type:user", headers=headers)
    return resp.json()

def query_namespace():
    resp = requests.get(f"{BASE_URL}/entities/list?namespace=type", headers=headers)
    return resp.json()

def complex_query():
    resp = requests.get(
        f"{BASE_URL}/entities/query?filter=type:user&sort=id&limit=10", 
        headers=headers)
    return resp.json()

# Run benchmarks
print("=== EntityDB Turbo Benchmark ===")
print("Running performance tests...")

tests = [
    ("Get by ID", query_by_id),
    ("Query by Tag", query_by_tag),
    ("Namespace Query", query_namespace),
    ("Complex Query", complex_query)
]

# Sequential benchmarks
print("\n--- Sequential Performance ---")
sequential_results = []

for name, fn in tests:
    print(f"Benchmarking: {name}")
    result = benchmark_query(fn, name, iterations=100)
    if result:
        sequential_results.append(result)
        print(f"  Avg: {result['avg']:.2f}ms, P95: {result['p95']:.2f}ms, P99: {result['p99']:.2f}ms")

# Parallel benchmarks
print("\n--- Parallel Performance (QPS) ---")
parallel_results = []

for name, fn in tests:
    print(f"Benchmarking: {name}")
    result = parallel_benchmark(fn, name, threads=20, iterations=1000)
    if result:
        parallel_results.append(result)
        print(f"  QPS: {result['qps']:.2f}, Success Rate: {result['successful']/result['total_queries']*100:.1f}%")

# Summary
print("\n=== Performance Summary ===")
print("\nSequential Latency (ms):")
print(f"{'Query':<20} {'Avg':<8} {'P95':<8} {'P99':<8}")
print("-" * 50)
for r in sequential_results:
    print(f"{r['name']:<20} {r['avg']:<8.2f} {r['p95']:<8.2f} {r['p99']:<8.2f}")

print("\nParallel Throughput (QPS):")
print(f"{'Query':<20} {'QPS':<10} {'Success %':<10}")
print("-" * 40)
for r in parallel_results:
    print(f"{r['name']:<20} {r['qps']:<10.2f} {r['successful']/r['total_queries']*100:<10.1f}")

# Calculate improvement factor
baseline_avg = 189  # Previous average: 189ms
current_avg = statistics.mean([r['avg'] for r in sequential_results])
improvement = baseline_avg / current_avg

print(f"\nOverall average latency: {current_avg:.2f}ms")
print(f"Performance improvement: {improvement:.1f}x faster")

if improvement >= 100:
    print("\n🚀 ACHIEVED 100X PERFORMANCE IMPROVEMENT! 🚀")
else:
    print(f"\nNeed {100/improvement:.1f}x more improvement to reach 100x target")