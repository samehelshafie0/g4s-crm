"""Extract tabular line items from a vendor PDF.

Runs only on a server-staged file in a private directory. Vendor quotations
position text freely rather than as real tables, so geometric row grouping in
the browser misses most of them; MuPDF's table finder uses ruling lines and
column alignment and recovers far more. Output is a matrix per table, which the
caller maps to fields with the same column detection used for spreadsheets.
"""
import json
import re
import sys
import warnings

warnings.filterwarnings("ignore")
import pymupdf  # noqa: E402

MAX_PAGES = 200
MAX_TABLES = 60
MAX_ROWS = 4000
NUMERIC = re.compile(r"^\(?[-+]?[\d,]+(\.\d+)?\)?$")
# Some PDFs space every glyph, so "812.00" arrives as "8 12.00" or "1 ,250.00".
SPACED_NUMBER = re.compile(r"^[\s\d,.\-+()$£€¥]+$")


def clean(cell):
    text = (cell or "").replace("\n", " ").replace(" ", " ")
    text = re.sub(r"\s+", " ", text).strip()
    if text and SPACED_NUMBER.match(text) and re.search(r"\d", text):
        # Only collapse spacing when the cell is purely a number; never inside prose.
        compact = re.sub(r"\s+", "", text)
        if NUMERIC.match(compact.replace("$", "").replace("£", "").replace("€", "").replace("¥", "")):
            return compact
    return text


def numeric(cell):
    text = (cell or "").strip().replace("$", "").replace("£", "").replace("€", "").replace("¥", "")
    text = text.replace(",", "").strip()
    if not text or not NUMERIC.match(text):
        return None
    try:
        return float(text.strip("()"))
    except ValueError:
        return None


def is_line_item(cells):
    """A product line carries at least one positive amount and one real label."""
    has_amount = any((numeric(c) or 0) > 0 for c in cells)
    has_label = any(c and numeric(c) is None and len(c) > 2 for c in cells)
    return has_amount and has_label


MONEY = re.compile(r"[$£€¥]|\d\.\d{2}\b")


def is_priced(cells):
    """A money-formatted cell — a symbol or two decimal places — separates a
    price table from a contact or banking block, whose numbers are phone
    numbers and account references."""
    return is_line_item(cells) and any(MONEY.search(c or "") for c in cells)


def table_rows(table):
    try:
        raw = table.extract()
    except Exception:
        return []
    rows = []
    for row in raw:
        cells = [clean(c) for c in row]
        if any(cells):
            rows.append(cells)
    return rows


def score(rows):
    """Prefer tables that look like a price list over address and banking blocks.

    Money-formatted rows are weighted far above merely numeric ones, so a real
    price table outranks a block of phone numbers even when it has fewer rows.
    """
    priced = sum(1 for r in rows if is_priced(r))
    numeric_only = sum(1 for r in rows if is_line_item(r))
    return priced * 10 + numeric_only


def extract(path):
    document = pymupdf.open(path)
    tables = []
    try:
        for index, page in enumerate(document):
            if index >= MAX_PAGES or len(tables) >= MAX_TABLES:
                break
            found = []
            for strategy in ("lines_strict", "text"):
                try:
                    found = list(page.find_tables(strategy=strategy))
                except Exception:
                    found = []
                if found:
                    break
            for table in found:
                rows = table_rows(table)
                if len(rows) < 2:
                    continue
                if score(rows) == 0:
                    continue
                tables.append({
                    "page": index + 1,
                    "rank": score(rows),
                    "lineItemRows": sum(1 for r in rows if is_line_item(r)),
                    "rows": rows[:MAX_ROWS],
                })
    finally:
        document.close()

    # A price list often spans pages as separate tables with the same shape;
    # returning them ordered by usefulness lets the caller take the best one.
    tables.sort(key=lambda t: (-t["rank"], t["page"]))
    return {"tables": tables[:MAX_TABLES]}


if __name__ == "__main__":
    # The library writes advisory banners to stdout, so the result goes to a file
    # and stdout carries only the error message the caller shows the user.
    try:
        result = extract(sys.argv[1])
        if not result["tables"]:
            raise ValueError(
                "No priced table could be found in this PDF. It may be a scan, or laid out without a table."
            )
        with open(sys.argv[2], "w", encoding="utf-8") as out:
            json.dump(result, out)
    except Exception as exc:  # noqa: BLE001 - the message is shown to the user
        message = str(exc) if isinstance(exc, ValueError) else "The PDF could not be read."
        print(json.dumps({"error": message}))
        sys.exit(1)
