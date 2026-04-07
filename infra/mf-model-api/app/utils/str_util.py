import hashlib
import secrets
import string


def generate_random_string(length=32):
    alphabet = string.ascii_letters + string.digits
    return ''.join(secrets.choice(alphabet) for _ in range(length))


def get_md5_key(s: str):
    m = hashlib.new("md5")
    m.update(bytes(s, encoding="utf8"))
    res = m.hexdigest()
    return res


def has_common_substring(a, b):
    len_a = len(a)
    len_b = len(b)
    for i in range(1, len_b + 1):
        if a[len_a - i: len_a] == b[0: i]:
            return True
    return False
