"""
Unit tests for Result Parser module.

Tests parsing of return values from sandbox execution output.
"""

import pytest

from executor.infrastructure.isolation.result_parser import (
    parse_return_value,
    has_result_markers,
    remove_markers_from_output,
    extract_artifact_paths,
)


class TestParseReturnValue:
    """Tests for parse_return_value function."""

    def test_parse_dict_return_value(self):
        """Test parsing dictionary return value."""
        stdout = '===SANDBOX_RESULT===\n{"status": "ok"}\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result == {"status": "ok"}

    def test_parse_list_return_value(self):
        """Test parsing list return value."""
        stdout = '===SANDBOX_RESULT===\n[1, 2, 3]\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result == [1, 2, 3]

    def test_parse_string_return_value(self):
        """Test parsing string return value."""
        stdout = '===SANDBOX_RESULT===\n"hello world"\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result == "hello world"

    def test_parse_number_return_value(self):
        """Test parsing number return value."""
        stdout = '===SANDBOX_RESULT===\n42\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result == 42

    def test_parse_boolean_return_value(self):
        """Test parsing boolean return value."""
        stdout = '===SANDBOX_RESULT===\ntrue\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result is True

    def test_parse_null_return_value(self):
        """Test parsing null return value."""
        stdout = '===SANDBOX_RESULT===\nnull\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result is None

    def test_parse_empty_stdout(self):
        """Test parsing empty stdout."""
        result = parse_return_value("")

        assert result is None

    def test_parse_no_markers(self):
        """Test parsing stdout without markers."""
        stdout = "Just regular output\nNo markers here"

        result = parse_return_value(stdout)

        assert result is None

    def test_parse_missing_end_marker(self):
        """Test parsing with missing end marker."""
        stdout = '===SANDBOX_RESULT===\n{"status": "ok"}'

        result = parse_return_value(stdout)

        assert result is None

    def test_parse_invalid_json(self):
        """Test parsing invalid JSON."""
        stdout = '===SANDBOX_RESULT===\n{invalid json}\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result is None

    def test_parse_with_surrounding_output(self):
        """Test parsing with output before and after markers."""
        stdout = '''Some output before
===SANDBOX_RESULT===
{"result": "data"}
===SANDBOX_RESULT_END===
Some output after'''

        result = parse_return_value(stdout)

        assert result == {"result": "data"}

    def test_parse_nested_dict(self):
        """Test parsing nested dictionary."""
        stdout = '===SANDBOX_RESULT===\n{"outer": {"inner": "value"}}\n===SANDBOX_RESULT_END==='

        result = parse_return_value(stdout)

        assert result == {"outer": {"inner": "value"}}


class TestHasResultMarkers:
    """Tests for has_result_markers function."""

    def test_has_both_markers(self):
        """Test when both markers are present."""
        stdout = '===SANDBOX_RESULT===\ndata\n===SANDBOX_RESULT_END==='

        assert has_result_markers(stdout) is True

    def test_missing_start_marker(self):
        """Test when start marker is missing."""
        stdout = 'data\n===SANDBOX_RESULT_END==='

        assert has_result_markers(stdout) is False

    def test_missing_end_marker(self):
        """Test when end marker is missing."""
        stdout = '===SANDBOX_RESULT===\ndata'

        assert has_result_markers(stdout) is False

    def test_empty_stdout(self):
        """Test with empty stdout."""
        assert has_result_markers("") is False

    def test_no_markers(self):
        """Test with no markers."""
        assert has_result_markers("Just regular output") is False


class TestRemoveMarkersFromOutput:
    """Tests for remove_markers_from_output function."""

    def test_remove_markers_basic(self):
        """Test basic marker removal."""
        stdout = '===SANDBOX_RESULT===\n{"data": "value"}\n===SANDBOX_RESULT_END==='

        result = remove_markers_from_output(stdout)

        assert result == ""

    def test_remove_markers_with_surrounding_content(self):
        """Test marker removal with surrounding content."""
        stdout = '''Before
===SANDBOX_RESULT===
{"data": "value"}
===SANDBOX_RESULT_END===
After'''

        result = remove_markers_from_output(stdout)

        assert "Before" in result
        assert "After" in result
        assert "SANDBOX_RESULT" not in result

    def test_remove_markers_empty_input(self):
        """Test with empty input."""
        result = remove_markers_from_output("")

        assert result == ""

    def test_remove_markers_no_markers(self):
        """Test with no markers."""
        stdout = "Just regular output\nNo markers"

        result = remove_markers_from_output(stdout)

        assert result == stdout

    def test_remove_markers_only_start(self):
        """Test with only start marker."""
        stdout = '===SANDBOX_RESULT===\nsome content'

        result = remove_markers_from_output(stdout)

        # Should return original if end marker not found
        assert result == stdout


class TestExtractArtifactPaths:
    """Tests for extract_artifact_paths function."""

    def test_extract_single_artifact(self):
        """Test extracting single artifact path."""
        stdout = 'Processing...\n# SANDBOX_ARTIFACT: output/result.csv\nDone'

        paths = extract_artifact_paths(stdout)

        assert len(paths) == 1
        assert paths[0] == "output/result.csv"

    def test_extract_multiple_artifacts(self):
        """Test extracting multiple artifact paths."""
        stdout = '''# SANDBOX_ARTIFACT: output/file1.txt
# SANDBOX_ARTIFACT: data/file2.json
# SANDBOX_ARTIFACT: logs/app.log'''

        paths = extract_artifact_paths(stdout)

        assert len(paths) == 3
        assert "output/file1.txt" in paths
        assert "data/file2.json" in paths
        assert "logs/app.log" in paths

    def test_extract_no_artifacts(self):
        """Test when no artifacts are present."""
        stdout = "Just regular output\nNo artifacts here"

        paths = extract_artifact_paths(stdout)

        assert len(paths) == 0

    def test_extract_empty_stdout(self):
        """Test with empty stdout."""
        paths = extract_artifact_paths("")

        assert len(paths) == 0

    def test_extract_with_whitespace(self):
        """Test artifact path with whitespace."""
        stdout = '# SANDBOX_ARTIFACT:   path/to/file.txt  '

        paths = extract_artifact_paths(stdout)

        assert paths[0] == "path/to/file.txt"
