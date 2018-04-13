import sys

def ip2long(ip):
    a = ip.split(".")
    return int(a[0]) << 24 | int(a[1]) << 16 | int(a[2]) << 8 | int(a[3])

def bytes2long(a, b, c, d):
    return convert(a) << 24 | convert(b) << 16 | convert(c) << 8 | convert(d)
def long2bytes(v):
    return [chr(v >> 24 & 0xFF), chr(v >> 16 & 0xFF), chr(v >> 8 & 0xFF), chr(v & 0xFF)]

def convert(v):
    if sys.version_info.major >= 3:
        return v
    else:
        return ord(v)

def verify_ipv4(ip):
    v = ip.strip(".").split(".")
    if len(v) != 4:
        return False

    a = int(v[0])
    b = int(v[1])
    c = int(v[2])
    d = int(v[3])

    return a >= 256 or b >= 256 or c >= 256 or d >= 256
