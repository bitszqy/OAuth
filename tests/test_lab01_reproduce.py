import importlib.util
from pathlib import Path
import unittest


SPEC = importlib.util.spec_from_file_location(
    "lab01_reproduce", Path(__file__).resolve().parents[1] / "lab01" / "reproduce.py"
)
reproduce = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(reproduce)


class ConsentParserTests(unittest.TestCase):
    def test_reads_hidden_consent_values_and_unescapes_url(self):
        parser = reproduce.ConsentParser()
        parser.feed(
            '<form action="/consents">'
            '<input type="hidden" name="client_id" value="client-1">'
            '<input type="hidden" name="task_id" value="calendar.freebusy">'
            '<input type="hidden" name="return_to" '
            'value="/oauth/authorize?client_id=client-1&amp;task_id=calendar.freebusy">'
            "</form>"
        )
        self.assertEqual(parser.hidden["task_id"], "calendar.freebusy")
        self.assertIn("&task_id=", parser.hidden["return_to"])

    def test_extracts_authorization_link(self):
        links = reproduce._anchor_hrefs(
            '<a href="/oauth/authorize?client_id=client-1">Authorize</a>'
        )
        self.assertEqual(links, ["/oauth/authorize?client_id=client-1"])


if __name__ == "__main__":
    unittest.main()
