# test_05_expert.py
from typing import Generator, List, Dict, Any

class ReportGenerator:
    """
    A complex class that acts as a context manager for generating reports.
    It contains a nested class for formatting and a generator for processing.
    """
    class _Formatter:
        """
        A nested private class for formatting report data.
        Its existence and methods should be captured within the parent scope.
        """
        def to_json(self, data: Dict) -> str:
            import json
            return json.dumps(data, indent=2)

        def to_csv_line(self, data_row: List) -> str:
            return ",".join(map(str, data_row))

    def __init__(self, report_name: str):
        self.report_name = report_name
        self._formatter = self._Formatter()
        self.is_open = False
        print(f"Initialized report: {self.report_name}")

    def __enter__(self):
        """Enters the context, marking the report as open."""
        print("Entering report context...")
        self.is_open = True
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Exits the context, finalizing the report."""
        print("Exiting report context. Finalizing report.")
        self.is_open = False
        # Returning False re-raises any exceptions.
        return False

    def data_chunker(self, data: List[Dict], chunk_size: int) -> Generator[List[Dict], None, None]:
        """
        A generator function that yields data in chunks.
        The distiller should identify this as a generator due to 'yield'.
        """
        if not self.is_open:
            raise RuntimeError("Report context is not open.")
        for i in range(0, len(data), chunk_size):
            yield data[i:i + chunk_size]

    def get_formatter(self) -> _Formatter:
        """Returns an instance of the nested formatter class."""
        return self._formatter

if __name__ == "__main__":
    raw_data = [{"id": i, "value": i*10} for i in range(10)]

    with ReportGenerator("Monthly_Sales") as report:
        formatter = report.get_formatter()
        print(f"Report is open: {report.is_open}")

        # Using the generator
        for chunk in report.data_chunker(raw_data, chunk_size=3):
            print(f"Processing chunk of size {len(chunk)}")
            # Using the nested class's method
            print(formatter.to_json(chunk[0]))
    
    # After with-block, __exit__ is called
    # This would fail: report.data_chunker(raw_data, 2)