import time

class Point:
    __slots__ = ['x', 'y', 'z']
    def __init__(self, x, y, z):
        self.x = x
        self.y = y
        self.z = z

def update_point(p):
    return Point(p.x + 1, p.y + 2, p.z + 3)

def struct_test():
    max_iter = 5000000
    v = [Point(i, i, i) for i in range(max_iter)]
    for j in range(max_iter):
        v[j] = update_point(v[j])
        v[j] = update_point(v[j])
    return v[0].x

start = time.time()
s = struct_test()
end = time.time()
print(f"struct_test val: {s}")
print(f"Time: {int((end - start)*1000)}ms")
