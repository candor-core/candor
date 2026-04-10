import time

def map_test():
    m = {}
    max_iter = 100000
    for i in range(max_iter):
        k = str(i)
        m[k] = i * 2
        
    s = 0
    for j in range(max_iter):
        k = str(j)
        val = m.get(k)
        if val is not None:
            s += val
    return s

start = time.time()
s = map_test()
end = time.time()
print(f"map_test sum: {s}")
print(f"Time: {int((end - start)*1000)}ms")
