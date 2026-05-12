#!/usr/bin/env python3
"""
Simple Python Flask application for testing eBPF profiling with Alloy.
No SDK injection required - Alloy will auto-instrument via eBPF.
"""

import os
import time
import random
import math
import json
import hashlib
from flask import Flask, request, jsonify
from concurrent.futures import ThreadPoolExecutor
import threading

app = Flask(__name__)

# Get pod info from environment
POD_NAME = os.environ.get('POD_NAME', 'unknown-pod')
POD_NAMESPACE = os.environ.get('POD_NAMESPACE', 'default')

# Thread pool for parallel work
executor = ThreadPoolExecutor(max_workers=4)


# ============= CPU Intensive Functions =============

def fibonacci(n):
    """Recursive Fibonacci - CPU intensive"""
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)


def calculate_primes(limit):
    """Sieve of Eratosthenes - CPU intensive"""
    sieve = [True] * (limit + 1)
    sieve[0] = sieve[1] = False
    for i in range(2, int(limit**0.5) + 1):
        if sieve[i]:
            for j in range(i*i, limit + 1, i):
                sieve[j] = False
    return [i for i, is_prime in enumerate(sieve) if is_prime]


def matrix_multiply(size):
    """Matrix multiplication - CPU intensive"""
    # Create random matrices
    a = [[random.random() for _ in range(size)] for _ in range(size)]
    b = [[random.random() for _ in range(size)] for _ in range(size)]
    result = [[0 for _ in range(size)] for _ in range(size)]
    
    # Multiply matrices
    for i in range(size):
        for j in range(size):
            for k in range(size):
                result[i][j] += a[i][k] * b[k][j]
    return result


def hash_compute(data_size):
    """Compute SHA256 hashes repeatedly - CPU intensive"""
    data = os.urandom(data_size)
    for _ in range(1000):
        hashlib.sha256(data).hexdigest()
    return True


def json_parse_heavy():
    """Parse large JSON structure repeatedly"""
    large_dict = {
        "nested": [{"id": i, "name": f"item_{i}", "value": random.random()} 
                   for i in range(500)]
    }
    json_str = json.dumps(large_dict)
    for _ in range(100):
        json.loads(json_str)
    return True


# ============= Handlers =============

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({"status": "healthy", "pod": POD_NAME}), 200


@app.route('/ready', methods=['GET'])
def ready():
    """Readiness check endpoint"""
    return jsonify({"status": "ready", "pod": POD_NAME}), 200


@app.route('/hello', methods=['GET'])
def hello():
    """Simple hello endpoint with minimal CPU"""
    time.sleep(0.01)  # Simulate I/O
    return jsonify({
        "message": "Hello from Python demo!",
        "pod": POD_NAME,
        "namespace": POD_NAMESPACE
    })


@app.route('/fib', methods=['GET'])
def fib():
    """Fibonacci endpoint - good for CPU profiling"""
    n = request.args.get('n', default=35, type=int)
    
    # Limit n to prevent excessive CPU
    n = min(n, 38)
    
    start = time.time()
    result = fibonacci(n)
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "fibonacci",
        "input": n,
        "result": result,
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/primes', methods=['GET'])
def primes():
    """Prime number calculation - CPU intensive"""
    limit = request.args.get('limit', default=100000, type=int)
    limit = min(limit, 500000)  # Cap at 500k
    
    start = time.time()
    prime_list = calculate_primes(limit)
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "calculate_primes",
        "limit": limit,
        "prime_count": len(prime_list),
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/matrix', methods=['GET'])
def matrix():
    """Matrix multiplication - CPU intensive"""
    size = request.args.get('size', default=50, type=int)
    size = min(size, 100)  # Cap at 100x100
    
    start = time.time()
    result = matrix_multiply(size)
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "matrix_multiply",
        "size": size,
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/hash', methods=['GET'])
def hash_endpoint():
    """Hash computation - CPU intensive"""
    data_size = request.args.get('size', default=1024, type=int)
    data_size = min(data_size, 10240)  # Cap at 10KB
    
    start = time.time()
    hash_compute(data_size)
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "hash_compute",
        "data_size_bytes": data_size,
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/json-heavy', methods=['GET'])
def json_heavy():
    """JSON parsing/processing - CPU intensive"""
    iterations = request.args.get('iterations', default=5, type=int)
    iterations = min(iterations, 20)
    
    start = time.time()
    for _ in range(iterations):
        json_parse_heavy()
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "json_processing",
        "iterations": iterations,
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/mixed', methods=['GET'])
def mixed():
    """Mixed workload - combination of operations"""
    start = time.time()
    
    # Do a mix of operations
    result_fib = fibonacci(30)
    primes_under_10000 = calculate_primes(10000)
    matrix_multiply(30)
    
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "mixed_workload",
        "fib_30": result_fib,
        "prime_count_10000": len(primes_under_10000),
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/parallel', methods=['GET'])
def parallel():
    """Parallel execution - multiple CPU tasks"""
    tasks = request.args.get('tasks', default=4, type=int)
    tasks = min(tasks, 8)
    
    start = time.time()
    
    futures = []
    for i in range(tasks):
        # Submit different tasks to thread pool
        if i % 4 == 0:
            futures.append(executor.submit(fibonacci, 32))
        elif i % 4 == 1:
            futures.append(executor.submit(calculate_primes, 50000))
        elif i % 4 == 2:
            futures.append(executor.submit(matrix_multiply, 40))
        else:
            futures.append(executor.submit(hash_compute, 2048))
    
    results = [f.result() for f in futures]
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "parallel_workload",
        "tasks_executed": tasks,
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/slow', methods=['GET'])
def slow():
    """Slow endpoint - mix of CPU and sleep"""
    duration = request.args.get('duration', default=3, type=int)
    duration = min(duration, 10)
    
    start = time.time()
    
    # Do CPU work then sleep, repeatedly
    for i in range(duration):
        # CPU work
        fibonacci(30)
        calculate_primes(10000)
        # I/O sim (sleep)
        time.sleep(1)
    
    total_duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "slow_workload",
        "duration_seconds": duration,
        "total_duration_ms": round(total_duration, 2),
        "pod": POD_NAME
    })


@app.route('/error', methods=['GET'])
def error():
    """Error endpoint for testing error handling"""
    error_type = request.args.get('type', default='exception', type=str)
    
    if error_type == 'exception':
        raise ValueError("Intentional ValueError for testing")
    elif error_type == 'division':
        result = 1 / 0
    elif error_type == 'keyerror':
        d = {}
        value = d['nonexistent']
    elif error_type == 'typeerror':
        result = len(123)
    else:
        return jsonify({"error": f"Unknown error type: {error_type}"}), 400
    
    return jsonify({"message": "Should not reach here"}), 500


@app.route('/stress', methods=['GET'])
def stress():
    """Heavy CPU stress test - use with caution"""
    intensity = request.args.get('intensity', default=1, type=int)
    intensity = min(intensity, 3)
    
    start = time.time()
    
    for i in range(intensity * 2):
        fibonacci(35)
        calculate_primes(100000)
        matrix_multiply(60)
    
    duration = (time.time() - start) * 1000
    
    return jsonify({
        "function": "stress_test",
        "intensity": intensity,
        "duration_ms": round(duration, 2),
        "pod": POD_NAME
    })


@app.route('/info', methods=['GET'])
def info():
    """Show application info and metrics"""
    return jsonify({
        "application": "python-demo-app",
        "version": "1.0.0",
        "language": "python",
        "pod": POD_NAME,
        "namespace": POD_NAMESPACE,
        "endpoints": [
            "/health", "/ready", "/hello", "/fib", "/primes",
            "/matrix", "/hash", "/json-heavy", "/mixed",
            "/parallel", "/slow", "/error", "/stress"
        ]
    })


@app.errorhandler(Exception)
def handle_exception(e):
    """Global error handler"""
    return jsonify({
        "error": str(e),
        "pod": POD_NAME,
        "timestamp": time.time()
    }), 500


if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    app.run(host='0.0.0.0', port=port, threaded=True)