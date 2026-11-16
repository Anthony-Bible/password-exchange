"""Contract tests for the MySQL -> Postgres verification logic (#569).

The verifier must report a mismatch for any real difference, must not report
one for representation-only differences (JSON whitespace/key order, bool vs
tinyint, hex case), and must never put row values in its report.
"""
import json
import os
import sys

from hypothesis import given, strategies as st

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import verify  # noqa: E402

KINDS = ["str", "int", "bool", "ts", "hex", "json"]

json_values = st.recursive(
    st.none() | st.booleans() | st.integers() | st.text(max_size=5),
    lambda inner: st.lists(inner, max_size=3) | st.dictionaries(st.text(max_size=3), inner, max_size=3),
    max_leaves=8,
)

rows = st.dictionaries(
    keys=st.integers(min_value=1, max_value=10_000),
    values=st.tuples(st.text(max_size=8), st.integers(-5, 5)),
    max_size=20,
)


@given(rows)
def test_identical_tables_match(table):
    report = verify.compare(table, dict(table), columns=["a", "b"])
    assert report.ok
    assert report.missing == [] and report.extra == [] and report.differing == {}


@given(rows, st.data())
def test_a_changed_cell_is_reported_by_key_and_column(table, data):
    if not table:
        return
    key = data.draw(st.sampled_from(sorted(table)))
    a, b = table[key]
    changed = dict(table)
    changed[key] = (a, b + 1)
    report = verify.compare(table, changed, columns=["a", "b"])
    assert not report.ok
    assert report.differing == {key: ["b"]}


@given(rows, st.data())
def test_missing_and_extra_rows_are_reported(table, data):
    if not table:
        return
    key = data.draw(st.sampled_from(sorted(table)))
    dst = dict(table)
    del dst[key]
    dst[-1] = ("x", 0)
    report = verify.compare(table, dst, columns=["a", "b"])
    assert report.missing == [key]
    assert report.extra == [-1]
    assert not report.ok


@given(json_values)
def test_json_representation_differences_are_not_mismatches(value):
    mysql_text = json.dumps(value, separators=(", ", ": "))
    pg_text = json.dumps(value, sort_keys=True)
    assert verify.normalize("json", mysql_text) == verify.normalize("json", pg_text)


@given(st.binary(max_size=16))
def test_hex_case_is_ignored(raw):
    h = raw.hex()
    assert verify.normalize("hex", h.upper()) == verify.normalize("hex", h.lower())


@given(st.sampled_from([0, 1]))
def test_tinyint_and_bool_compare_equal(n):
    assert verify.normalize("bool", n) == verify.normalize("bool", bool(n))


@given(st.sampled_from(KINDS))
def test_null_stays_distinct_from_empty_and_zero(kind):
    empty = {"str": "", "int": 0, "bool": 0, "ts": 0, "hex": "", "json": "null"}[kind]
    assert verify.normalize(kind, None) != verify.normalize(kind, empty) or kind == "json"


@given(rows)
def test_digest_ignores_row_order_but_not_content(table):
    shuffled = dict(reversed(list(table.items())))
    assert verify.digest(table) == verify.digest(shuffled)
    if table:
        k = next(iter(table))
        changed = dict(table)
        changed[k] = (table[k][0] + "!", table[k][1])
        assert verify.digest(changed) != verify.digest(table)


tagged_rows = st.dictionaries(
    keys=st.integers(min_value=1, max_value=10_000),
    values=st.tuples(st.text(max_size=8).map(lambda t: "VAL\x1f" + t), st.integers(-5, 5)),
    min_size=1,
    max_size=20,
)


@given(tagged_rows, tagged_rows)
def test_report_never_contains_row_values_or_keys(src, dst):
    text = verify.compare(src, dst, columns=["a", "b"]).render("messages")
    assert "VAL\x1f" not in text
    for key in set(src) | set(dst):
        assert f"={key}" not in text and f" {key}," not in text


@given(
    st.integers(min_value=1, max_value=10_000),
    st.booleans(),
    st.one_of(st.none(), st.integers(min_value=0, max_value=10_000)),
)
def test_next_id_must_exceed_source_max(last_value, is_called, src_max):
    next_id = last_value + 1 if is_called else last_value
    assert verify.next_id_ok(last_value, is_called, src_max) == (next_id > (src_max or 0))
