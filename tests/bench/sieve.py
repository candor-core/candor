def count_primes(limit):
    sieve = [True] * (limit + 1)
    i = 2
    while i * i <= limit:
        if sieve[i]:
            j = i * i
            while j <= limit:
                sieve[j] = False
                j += i
        i += 1
    return sum(1 for i in range(2, limit + 1) if sieve[i])

print(count_primes(1000000))
