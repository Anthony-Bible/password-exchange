"""Dependency, import-hygiene and CI wiring requirements for the slackbot."""
import ast
import re

import yaml

from conftest import REPO_ROOT, SLACKBOT_DIR


def requirement_lines(path):
    return [
        l.strip() for l in path.read_text().splitlines()
        if l.strip() and not l.strip().startswith("#")
    ]


def test_runtime_requirements_move_to_sqlalchemy2_and_psycopg3():
    reqs = requirement_lines(SLACKBOT_DIR / "requirements.txt")
    assert "SQLAlchemy==2.0.*" in reqs or any(re.fullmatch(r"SQLAlchemy==2\.0\.\d+", r) for r in reqs)
    assert any(re.fullmatch(r"psycopg\[binary\]==3\.2\.(\*|\d+)", r) for r in reqs)


def test_mysqlclient_kept_and_unused_mariadb_removed():
    reqs = requirement_lines(SLACKBOT_DIR / "requirements.txt")
    assert "mysqlclient==2.1.1" in reqs
    assert not any(r.lower().startswith("mariadb") for r in reqs)


def test_dev_requirements_exist_with_test_tooling():
    reqs = " ".join(requirement_lines(SLACKBOT_DIR / "requirements-dev.txt")).lower()
    for pkg in ("pytest", "hypothesis", "testcontainers", "pyyaml"):
        assert pkg in reqs


def test_dev_requirements_are_exactly_pinned():
    """Renovate controls upgrades; unpinned test deps let an upstream release
    break CI (and gate the slackbot image build) with no repo change."""
    reqs = requirement_lines(SLACKBOT_DIR / "requirements-dev.txt")
    assert reqs
    for r in reqs:
        assert re.fullmatch(r"[A-Za-z0-9_.-]+(\[[A-Za-z0-9_,.-]+\])?==\d+(\.\d+)*", r), (
            f"requirements-dev.txt entry {r!r} must be pinned with an exact ==version"
        )


def test_program_does_not_import_mysqldb_directly():
    tree = ast.parse((SLACKBOT_DIR / "program.py").read_text())
    imported = set()
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            imported |= {a.name.split(".")[0] for a in node.names}
        elif isinstance(node, ast.ImportFrom) and node.module:
            imported.add(node.module.split(".")[0])
    assert "MySQLdb" not in imported


def load_ci():
    return yaml.safe_load((REPO_ROOT / ".github/workflows/ci.yml").read_text())


def test_ci_has_python311_slackbot_test_job():
    job = load_ci()["jobs"]["test_slackbot"]
    steps = job["steps"]
    blob = yaml.dump(steps)
    assert "3.11" in blob
    assert "slackbot/requirements.txt" in blob and "slackbot/requirements-dev.txt" in blob
    assert "pytest slackbot/tests" in blob


def test_slackbot_image_build_is_gated_by_slackbot_tests():
    needs = load_ci()["jobs"]["build_slackbot_image"]["needs"]
    needs = [needs] if isinstance(needs, str) else needs
    assert "test_slackbot" in needs
