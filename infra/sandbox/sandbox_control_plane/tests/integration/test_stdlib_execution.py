"""
Integration Tests for Python Standard Library Execution

Tests for code execution using various Python standard libraries.
Each test case covers a different standard library module.

These tests require:
- Running docker-compose stack (control-plane, executor)
- Test template available
"""
import pytest
import asyncio
from httpx import AsyncClient


@pytest.mark.asyncio
class TestStdlibDataProcessing:
    """Test cases for data processing standard libraries."""

    async def test_stdlib_json_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the json module for JSON processing."""
        code = '''
import json

def handler(event):
    # Create a complex data structure
    data = {
        "name": "Test",
        "values": [1, 2, 3],
        "nested": {
            "a": "alpha",
            "b": "bravo"
        }
    }

    # Serialize to JSON
    json_str = json.dumps(data, indent=2)

    # Parse back
    parsed = json.loads(json_str)

    # Test with different separators
    compact = json.dumps(data, separators=(",", ":"))

    # Test with sort_keys
    sorted_json = json.dumps({"z": 1, "a": 2}, sort_keys=True)

    return {
        "json_string": json_str,
        "parsed_name": parsed["name"],
        "compact_length": len(compact),
        "sorted_starts_with_a": sorted_json.startswith('{"a"')
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["parsed_name"] == "Test"

    async def test_stdlib_csv_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the csv module for CSV processing."""
        code = '''
import csv
import io

def handler(event):
    # Create CSV data
    output = io.StringIO()
    writer = csv.writer(output)

    writer.writerow(["Name", "Age", "City"])
    writer.writerow(["Alice", 30, "NYC"])
    writer.writerow(["Bob", 25, "LA"])

    csv_data = output.getvalue()

    # Parse CSV data
    input_data = io.StringIO(csv_data)
    reader = csv.reader(input_data)
    rows = list(reader)

    # Use DictReader
    dict_input = io.StringIO(csv_data)
    dict_reader = csv.DictReader(dict_input)
    dict_rows = list(dict_reader)

    return {
        "row_count": len(rows),
        "header": rows[0],
        "first_data_row": rows[1],
        "dict_first_name": dict_rows[0]["Name"]
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["dict_first_name"] == "Alice"

    async def test_stdlib_xml_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the xml.etree module for XML processing."""
        code = '''
import xml.etree.ElementTree as ET

def handler(event):
    # Create XML
    root = ET.Element("root")
    child = ET.SubElement(root, "child")
    child.set("attr", "value")
    child.text = "Text content"

    # Convert to string
    xml_str = ET.tostring(root, encoding="unicode")

    # Parse XML
    parsed_root = ET.fromstring(xml_str)

    # Find elements
    child_elem = parsed_root.find("child")

    return {
        "xml_string": xml_str,
        "child_tag": child_elem.tag,
        "child_attr": child_elem.get("attr"),
        "child_text": child_elem.text
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["child_text"] == "Text content"

    async def test_stdlib_pickle_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the pickle module for object serialization."""
        code = '''
import pickle

def handler(event):
    # Create a complex object
    data = {
        "list": [1, 2, 3],
        "dict": {"key": "value"},
        "tuple": (4, 5, 6),
        "set": {7, 8, 9}
    }

    # Serialize
    pickled = pickle.dumps(data)

    # Deserialize
    unpickled = pickle.loads(pickled)

    # Verify
    return {
        "original_list": data["list"],
        "unpickled_list": unpickled["list"],
        "pickle_size": len(pickled),
        "equal": data == unpickled
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["equal"] is True

    async def test_stdlib_base64_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the base64 module for base64 encoding/decoding."""
        code = '''
import base64

def handler(event):
    # String to encode
    original = "Hello, Base64!"

    # Encode
    encoded = base64.b64encode(original.encode()).decode()

    # Decode
    decoded = base64.b64decode(encoded).decode()

    # URL-safe encoding
    url_data = "data=data:value"
    url_encoded = base64.urlsafe_b64encode(url_data.encode()).decode()
    url_decoded = base64.urlsafe_b64decode(url_encoded).decode()

    return {
        "original": original,
        "encoded": encoded,
        "decoded": decoded,
        "url_safe_decoded": url_decoded
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["decoded"] == "Hello, Base64!"


@pytest.mark.asyncio
class TestStdlibDatetime:
    """Test cases for datetime standard libraries."""

    async def test_stdlib_datetime_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the datetime module for date and time operations."""
        code = '''
from datetime import datetime, date, time, timedelta

def handler(event):
    # Current datetime
    now = datetime.now()
    today = date.today()

    # Create specific datetime
    dt = datetime(2024, 1, 15, 10, 30, 45)

    # Date arithmetic
    tomorrow = today + timedelta(days=1)
    next_week = today + timedelta(weeks=1)

    # Format datetime
    formatted = now.strftime("%Y-%m-%d %H:%M:%S")

    # Parse datetime
    parsed = datetime.strptime("2024-01-15 10:30:00", "%Y-%m-%d %H:%M:%S")

    # Time difference
    diff = parsed - dt

    return {
        "formatted_now": formatted,
        "tomorrow_day": tomorrow.day,
        "dt_year": dt.year,
        "parsed_hour": parsed.hour,
        "diff_days": diff.days
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_time_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the time module for time-related functions."""
        code = '''
import time

def handler(event):
    # Get current time
    current = time.time()

    # Get struct_time
    local = time.localtime(current)

    # Format time
    formatted = time.strftime("%Y-%m-%d %H:%M:%S", local)

    # Parse time
    parsed = time.strptime("2024-01-15 10:30:00", "%Y-%m-%d %H:%M:%S")

    # Sleep briefly (0.1 seconds)
    start = time.time()
    time.sleep(0.1)
    elapsed = time.time() - start

    return {
        "current_time": current,
        "formatted": formatted,
        "parsed_hour": parsed.tm_hour,
        "slept_at_least": elapsed >= 0.1
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_calendar_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the calendar module for calendar operations."""
        code = '''
import calendar

def handler(event):
    # Get weekday name
    weekday_name = calendar.day_name[0]

    # Get month name
    month_name = calendar.month_name[1]

    # Check if leap year
    is_leap_2024 = calendar.isleap(2024)
    is_leap_2023 = calendar.isleap(2023)

    # Get week day
    weekday = calendar.weekday(2024, 1, 15)

    # Month range
    month_range = calendar.monthrange(2024, 2)

    return {
        "monday_name": weekday_name,
        "january_name": month_name,
        "2024_is_leap": is_leap_2024,
        "2023_is_leap": is_leap_2023,
        "weekday_jan_15_2024": weekday,
        "feb_2024_first_weekday": month_range[0],
        "feb_2024_days": month_range[1]
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["2024_is_leap"] is True


@pytest.mark.asyncio
class TestStdlibMath:
    """Test cases for math and statistics standard libraries."""

    async def test_stdlib_math_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the math module for mathematical operations."""
        code = '''
import math

def handler(event):
    # Basic operations
    sqrt_result = math.sqrt(16)
    power_result = math.pow(2, 10)

    # Trigonometry
    sin_val = math.sin(math.pi / 2)
    cos_val = math.cos(0)

    # Constants
    pi = math.pi
    e = math.e

    # Rounding
    ceil_val = math.ceil(3.2)
    floor_val = math.floor(3.8)

    # Logarithms
    log10_val = math.log10(100)

    # Factorial
    fact = math.factorial(5)

    return {
        "sqrt": sqrt_result,
        "power": power_result,
        "sin_pi_over_2": sin_val,
        "cos_0": cos_val,
        "pi_approx": round(pi, 5),
        "e_approx": round(e, 5),
        "ceil": ceil_val,
        "floor": floor_val,
        "log10_100": log10_val,
        "factorial_5": fact
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_random_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the random module for random number generation."""
        code = '''
import random

def handler(event):
    # Set seed for reproducibility
    random.seed(42)

    # Random float
    rand_float = random.random()

    # Random integer
    rand_int = random.randint(1, 100)

    # Random choice
    items = ["apple", "banana", "cherry"]
    choice = random.choice(items)

    # Random sample
    sample = random.sample([1, 2, 3, 4, 5], 3)

    # Shuffle
    deck = [1, 2, 3, 4, 5]
    random.shuffle(deck)

    # Random uniform
    uniform_val = random.uniform(1.0, 10.0)

    # Reset seed and get same value
    random.seed(42)
    same_float = random.random()

    return {
        "float_between_0_1": 0 <= rand_float <= 1,
        "int_between_1_100": 1 <= rand_int <= 100,
        "choice_in_options": choice in items,
        "sample_size": len(sample),
        "shuffled_deck": deck,
        "uniform_between_1_10": 1.0 <= uniform_val <= 10.0,
        "same_with_seed": rand_float == same_float
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_statistics_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the statistics module for statistical operations."""
        code = '''
import statistics

def handler(event):
    data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

    # Measures of central tendency
    mean_val = statistics.mean(data)
    median_val = statistics.median(data)
    mode_val = statistics.mode([1, 2, 2, 3, 4])

    # Measures of spread
    stdev_val = statistics.stdev(data)
    variance_val = statistics.variance(data)

    # Other statistics
    min_val = min(data)
    max_val = max(data)

    return {
        "mean": mean_val,
        "median": median_val,
        "mode": mode_val,
        "stdev": round(stdev_val, 4),
        "variance": round(variance_val, 4),
        "min": min_val,
        "max": max_val
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["mean"] == 5.5

    async def test_stdlib_decimal_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the decimal module for precise decimal arithmetic."""
        code = '''
from decimal import Decimal, getcontext

def handler(event):
    # Set precision
    getcontext().prec = 4

    # Create decimals
    a = Decimal('1.1')
    b = Decimal('2.2')

    # Arithmetic
    add = a + b
    mul = a * b

    # Higher precision for comparison
    getcontext().prec = 28
    precise = Decimal('0.1') + Decimal('0.2')

    # Comparison with float (shows difference)
    float_sum = 0.1 + 0.2
    decimal_sum = Decimal('0.1') + Decimal('0.2')

    # Quantize
    quantized = Decimal('10.123').quantize(Decimal('0.00'))

    return {
        "addition": float(add),
        "multiplication": float(mul),
        "precise_sum": str(precise),
        "float_sum_imprecise": abs(float_sum - 0.3) > 0,
        "quantized": str(quantized)
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"


@pytest.mark.asyncio
class TestStdlibString:
    """Test cases for string processing standard libraries."""

    async def test_stdlib_re_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the re module for regular expressions."""
        code = '''
import re

def handler(event):
    text = "Hello, my email is test@example.com and phone is 123-456-7890."

    # Find email
    email_pattern = r'\\b[\\w.-]+@[\\w.-]+\\.\\w+\\b'
    email_match = re.search(email_pattern, text)
    email = email_match.group() if email_match else None

    # Find phone
    phone_pattern = r'\\d{3}-\\d{3}-\\d{4}'
    phone_match = re.search(phone_pattern, text)
    phone = phone_match.group() if phone_match else None

    # Find all numbers
    numbers = re.findall(r'\\d+', text)

    # Replace
    replaced = re.sub(r'\\d+', 'X', phone)

    # Split
    words = re.split(r'\\s+', text)[:5]

    # Compile pattern
    pattern = re.compile(r'\\b\\w{4}\\b')
    four_letter_words = pattern.findall(text)

    return {
        "email": email,
        "phone": phone,
        "numbers": numbers,
        "replaced_phone": replaced,
        "first_five_words": words,
        "four_letter_words": four_letter_words
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["email"] == "test@example.com"

    async def test_stdlib_string_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the string module for string constants and operations."""
        code = '''
import string

def handler(event):
    # Get constants
    ascii_letters = string.ascii_letters
    ascii_lowercase = string.ascii_lowercase
    digits = string.digits
    punctuation = string.punctuation

    # Create a template
    template = string.Template("Hello, $name!")
    result = template.substitute(name="World")

    # Capwords
    text = "hello world from python"
    capped = string.capwords(text)

    # Formatter
    formatter = string.Formatter()
    formatted = formatter.format("{0} {1}", "Hello", "Python")

    return {
        "ascii_letters_length": len(ascii_letters),
        "ascii_lowercase": ascii_lowercase,
        "digits": digits,
        "punctuation_sample": punctuation[:5],
        "template_result": result,
        "capwords_result": capped,
        "formatter_result": formatted
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["template_result"] == "Hello, World!"


@pytest.mark.asyncio
class TestStdlibDataStructures:
    """Test cases for data structure standard libraries."""

    async def test_stdlib_collections_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the collections module for specialized containers."""
        code = '''
from collections import Counter, defaultdict, deque, OrderedDict, namedtuple

def handler(event):
    # Counter
    data = [1, 2, 2, 3, 3, 3, 4, 4, 4, 4]
    counter = Counter(data)

    # defaultdict
    dd = defaultdict(int)
    dd["a"] += 1
    dd["b"] += 2

    # deque
    dq = deque([1, 2, 3])
    dq.append(4)
    dq.appendleft(0)

    # OrderedDict
    od = OrderedDict()
    od["first"] = 1
    od["second"] = 2

    # namedtuple
    Point = namedtuple("Point", ["x", "y"])
    p = Point(10, 20)

    return {
        "most_common": counter.most_common(2),
        "defaultdict_a": dd["a"],
        "deque_length": len(dq),
        "deque_first": dq[0],
        "ordereddict_keys": list(od.keys()),
        "point_x": p.x,
        "point_y": p.y
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_itertools_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the itertools module for iterator functions."""
        code = '''
import itertools

def handler(event):
    # chain
    chained = list(itertools.chain([1, 2], [3, 4], [5, 6]))

    # cycle (take first 5)
    cycled = list(itertools.islice(itertools.cycle([1, 2, 3]), 5))

    # count
    counted = list(itertools.islice(itertools.count(10, 2), 5))

    # repeat
    repeated = list(itertools.repeat("hello", 3))

    # combinations
    combos = list(itertools.combinations([1, 2, 3], 2))

    # permutations
    perms = list(itertools.permutations([1, 2], 2))

    # product
    products = list(itertools.product([1, 2], ["a", "b"]))

    return {
        "chained": chained,
        "cycled": cycled,
        "counted": counted,
        "repeated": repeated,
        "combinations_count": len(combos),
        "permutations_count": len(perms),
        "products_count": len(products)
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_functools_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the functools module for higher-order functions."""
        code = '''
import functools

def handler(event):
    # partial
    def multiply(x, y):
        return x * y

    double = functools.partial(multiply, 2)
    double_result = double(5)

    # reduce
    numbers = [1, 2, 3, 4, 5]
    sum_result = functools.reduce(lambda x, y: x + y, numbers)

    # lru_cache
    @functools.lru_cache(maxsize=128)
    def fibonacci(n):
        if n < 2:
            return n
        return fibonacci(n-1) + fibonacci(n-2)

    fib_result = fibonacci(10)

    # wraps
    def my_decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)
        return wrapper

    @my_decorator
    def example():
        return "decorated"

    return {
        "double_of_5": double_result,
        "sum_of_list": sum_result,
        "fibonacci_10": fib_result,
        "decorated_result": example(),
        "wrapper_preserves_name": example.__name__
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"


@pytest.mark.asyncio
class TestStdlibCompression:
    """Test cases for data compression standard libraries."""

    async def test_stdlib_zlib_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the zlib module for compression."""
        code = '''
import zlib

def handler(event):
    # Original data
    data = b"This is a test string that will be compressed using zlib." * 10

    # Compress
    compressed = zlib.compress(data)

    # Decompress
    decompressed = zlib.decompress(compressed)

    # Compression ratio
    ratio = len(compressed) / len(data)

    # CRC32 checksum
    crc = zlib.crc32(data)

    return {
        "original_length": len(data),
        "compressed_length": len(compressed),
        "decompressed_matches": decompressed == data,
        "compression_ratio": round(ratio, 3),
        "crc32": crc
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_gzip_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the gzip module for gzip compression."""
        code = '''
import gzip
import io

def handler(event):
    # Original data
    data = b"Hello, World! " * 100

    # Compress with gzip
    compressed = gzip.compress(data)

    # Decompress
    decompressed = gzip.decompress(compressed)

    # Use file-like interface
    buf = io.BytesIO()
    with gzip.GzipFile(fileobj=buf, mode='wb') as f:
        f.write(data)
    buf.seek(0)

    with gzip.GzipFile(fileobj=buf, mode='rb') as f:
        file_decompressed = f.read()

    return {
        "original_length": len(data),
        "compressed_length": len(compressed),
        "decompressed_matches": decompressed == data,
        "file_decompressed_matches": file_decompressed == data
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_zipfile_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the zipfile module for ZIP file operations."""
        code = '''
import zipfile
import io

def handler(event):
    # Create a ZIP file in memory
    buf = io.BytesIO()

    with zipfile.ZipFile(buf, 'w', zipfile.ZIP_DEFLATED) as zf:
        zf.writestr("test1.txt", "Content of file 1")
        zf.writestr("test2.txt", "Content of file 2")
        zf.writestr("subdir/test3.txt", "Content of file 3")

    buf.seek(0)

    # Read the ZIP file
    with zipfile.ZipFile(buf, 'r') as zf:
        namelist = zf.namelist()
        file1_content = zf.read("test1.txt").decode()
        file2_content = zf.read("test2.txt").decode()

    # Create a new ZIP and extract
    buf2 = io.BytesIO()
    with zipfile.ZipFile(buf2, 'w') as zf:
        zf.writestr("extract_test.txt", "Extract me")

    buf2.seek(0)
    with zipfile.ZipFile(buf2, 'r') as zf:
        extract_content = zf.read("extract_test.txt").decode()

    return {
        "file_count": len(namelist),
        "namelist": namelist,
        "file1_content": file1_content,
        "file2_content": file2_content,
        "extract_content": extract_content
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert result["return_value"]["file_count"] == 3


@pytest.mark.asyncio
class TestStdlibCrypto:
    """Test cases for cryptographic standard libraries."""

    async def test_stdlib_hashlib_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the hashlib module for hash functions."""
        code = '''
import hashlib

def handler(event):
    data = b"Hello, World!"

    # MD5
    md5_hash = hashlib.md5(data).hexdigest()

    # SHA-1
    sha1_hash = hashlib.sha1(data).hexdigest()

    # SHA-256
    sha256_hash = hashlib.sha256(data).hexdigest()

    # SHA-512
    sha512_hash = hashlib.sha512(data).hexdigest()

    # BLAKE2b
    blake2b_hash = hashlib.blake2b(data).hexdigest()

    return {
        "md5": md5_hash,
        "sha1": sha1_hash,
        "sha256": sha256_hash,
        "sha512_length": len(sha512_hash),
        "blake2b_length": len(blake2b_hash)
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
        assert len(result["return_value"]["md5"]) == 32

    async def test_stdlib_hmac_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the hmac module for HMAC authentication."""
        code = '''
import hmac
import hashlib

def handler(event):
    message = b"Hello, World!"
    key = b"secret-key"

    # Create HMAC with SHA256
    h = hmac.new(key, message, hashlib.sha256)
    digest = h.hexdigest()

    # Verify HMAC
    h2 = hmac.new(key, message, hashlib.sha256)
    is_valid = hmac.compare_digest(digest, h2.hexdigest())

    # HMAC with different message
    h3 = hmac.new(key, b"Different message", hashlib.sha256)
    different_digest = h3.hexdigest()

    # Different key
    h4 = hmac.new(b"different-key", message, hashlib.sha256)
    is_different = not hmac.compare_digest(digest, h4.hexdigest())

    return {
        "hmac_digest": digest,
        "is_valid": is_valid,
        "different_digest": different_digest,
        "is_different": is_different
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_secrets_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the secrets module for secure random generation."""
        code = '''
import secrets

def handler(event):
    # Generate random token
    token = secrets.token_hex(16)

    # Generate random bytes
    bytes_token = secrets.token_bytes(16)

    # Generate URL-safe token
    url_token = secrets.token_urlsafe(16)

    # Compare digest
    a = secrets.token_bytes(32)
    b = secrets.token_bytes(32)
    is_equal = secrets.compare_digest(a, a)
    is_not_equal = not secrets.compare_digest(a, b)

    # Choice from sequence
    items = ["apple", "banana", "cherry", "date"]
    choice = secrets.choice(items)

    # Random number in range
    rand_below = secrets.randbelow(100)

    return {
        "token_length": len(token),
        "url_token_length": len(url_token),
        "token_is_hex": all(c in "0123456789abcdef" for c in token),
        "is_equal": is_equal,
        "is_not_equal": is_not_equal,
        "choice_in_items": choice in items,
        "rand_below_100": 0 <= rand_below < 100
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"


@pytest.mark.asyncio
class TestStdlibSystemAndProcess:
    """Test cases for system and process standard libraries."""

    async def test_stdlib_sys_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the sys module for system-specific parameters."""
        code = '''
import sys

def handler(event):
    # Get Python version
    version = sys.version

    # Get platform
    platform = sys.platform

    # Get path
    path = sys.path

    # Get argv
    argv = sys.argv

    # Get byte order
    byteorder = sys.byteorder

    # Get maxsize
    maxsize = sys.maxsize

    # Get module info
    modules_count = len(sys.modules)

    return {
        "python_version": version.split()[0],
        "platform": platform,
        "path_not_empty": len(path) > 0,
        "argv_length": len(argv),
        "byteorder": byteorder,
        "maxsize": maxsize,
        "modules_count": modules_count
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_subprocess_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the subprocess module for process spawning."""
        code = '''
import subprocess

def handler(event):
    # Run simple command
    result = subprocess.run(
        ["echo", "hello from subprocess"],
        capture_output=True,
        text=True
    )

    stdout = result.stdout.strip()
    returncode = result.returncode

    # Run command with pipes
    result2 = subprocess.run(
        ["cat", "/etc/osrelease"],
        capture_output=True,
        text=True
    )

    # Check if /etc/osrelease exists
    osrelease_exists = result2.returncode == 0

    return {
        "echo_output": stdout,
        "echo_returncode": returncode,
        "osrelease_exists": osrelease_exists,
        "osrelease_length": len(result2.stdout) if osrelease_exists else 0
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"

    async def test_stdlib_logging_module(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test using the logging module for logging."""
        code = '''
import logging
import io

def handler(event):
    # Create a string handler
    log_capture = io.StringIO()
    handler = logging.StreamHandler(log_capture)
    handler.setLevel(logging.INFO)

    # Configure logger
    logger = logging.getLogger("test_logger")
    logger.setLevel(logging.INFO)
    logger.addHandler(handler)

    # Log messages
    logger.info("Info message")
    logger.warning("Warning message")
    logger.error("Error message")

    # Get logged output
    log_output = log_capture.getvalue()

    # Clean up
    logger.removeHandler(handler)

    return {
        "log_output": log_output,
        "has_info": "Info message" in log_output,
        "has_warning": "Warning message" in log_output,
        "has_error": "Error message" in log_output
    }
'''
        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed: {result.get('stderr', '')}"
