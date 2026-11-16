#!/usr/bin/env bash
# MySQL -> Postgres copy for password-exchange (#569). See docs/runbooks/mysql-to-postgres.md.
#
#   copy.sh preflight [SUFFIX]   read-only gates: MySQL tz is UTC, no other sessions, frozen checksums
#   copy.sh prepare   [SUFFIX]   create the target schema (app migrations + slack-sdk tables)
#   copy.sh load      [SUFFIX]   pgloader passes, fileid shim, sequences, schema-diff gate
#   copy.sh verify    [SUFFIX]   row-by-row diff + Slack store lookups
#   copy.sh reset     [SUFFIX]   empty the target tables after an interrupted load, then rerun `load`
#
# SUFFIX is "" for prod or e.g. "_rehearsal". Required env (source the 0600 creds file):
#   MY_APP_USER MY_APP_PASS MY_APP_DB MY_OAUTH_USER MY_OAUTH_PASS MY_OAUTH_DB PG_APP_PASS PG_SLACK_PASS
#   APP_BIN (app binary built from the tag being deployed), PY (python with requirements.txt),
#   SLACK_CLIENT_ID, WORK (0700 dir for load files and schema dumps)
# Assumes podman pod $POD publishes MySQL on 127.0.0.1:13306 and Postgres on 127.0.0.1:15432.
# Prints counts, versions and timings only: never secrets or row values.
set -euo pipefail

STAGE=${1:?stage}
SUFFIX=${2:-}
HERE=$(cd "$(dirname "$0")" && pwd)
REPO=$(cd "$HERE/../.." && pwd)
POD=${POD:-pgmig}
PGLOADER_IMAGE=${PGLOADER_IMAGE:-ghcr.io/dimitri/pgloader@sha256:bd81512667c987c516c87c397d1039ce8e1a15f128e51b4cfe6177137489f4f5}
MY_PORT=13306
PG_PORT=15432
: "${WORK:?WORK must be a 0700 directory}"

my_app()   { MYSQL_PWD="$MY_APP_PASS"   mysql --skip-ssl -h127.0.0.1 -P$MY_PORT -u"$MY_APP_USER"   -N -B "$MY_APP_DB"   "$@"; }
my_oauth() { MYSQL_PWD="$MY_OAUTH_PASS" mysql --skip-ssl -h127.0.0.1 -P$MY_PORT -u"$MY_OAUTH_USER" -N -B "$MY_OAUTH_DB" "$@"; }
pg_app()   { PGPASSWORD="$PG_APP_PASS"   psql "host=127.0.0.1 port=$PG_PORT user=passwordexchange dbname=passwordexchange$SUFFIX sslmode=disable" -v ON_ERROR_STOP=1 -qAt "$@"; }
pg_slack() { PGPASSWORD="$PG_SLACK_PASS" psql "host=127.0.0.1 port=$PG_PORT user=slackbot dbname=slack_oauth$SUFFIX sslmode=disable" -v ON_ERROR_STOP=1 -qAt "$@"; }
die()  { echo "NO-GO: $*" >&2; exit 1; }
step() { echo "== $* ($(date -u +%H:%M:%S))"; }

slack_env() {
  env SLACKBOT_DIR="$REPO/slackbot" PASSWORDEXCHANGE_DBDRIVER=postgres PASSWORDEXCHANGE_DBHOST=127.0.0.1 \
    PASSWORDEXCHANGE_DBPORT=$PG_PORT OAUTHDB_USER=slackbot OAUTHDB_PASSWORD="$PG_SLACK_PASS" \
    OAUTHDB_NAME="slack_oauth$SUFFIX" "$@"
}

schema_dump() {  # $1 = label
  for db in passwordexchange slack_oauth; do
    local pw=$PG_APP_PASS user=passwordexchange
    [[ $db == slack_oauth ]] && { pw=$PG_SLACK_PASS; user=slackbot; }
    PGPASSWORD="$pw" pg_dump -h 127.0.0.1 -p $PG_PORT -U $user -d "$db$SUFFIX" \
      --schema-only --no-owner --no-privileges | grep -v '^\\\(un\)\?restrict ' > "$WORK/schema-$db-$1.sql"
  done
}

checksums() {
  my_app   -e "CHECKSUM TABLE messages, email_reminders, file_upload_sessions" | awk '{print $1, $2}'
  my_oauth -e "CHECKSUM TABLE slack_installations, slack_bots" | awk '{print $1, $2}'
}

case $STAGE in
preflight)
  step "MySQL time zone"
  tz=$(my_app -e "SELECT CONCAT(@@global.time_zone, '/', @@system_time_zone)")
  [[ $tz == "SYSTEM/UTC" || $tz == "+00:00/"* ]] || die "MySQL time zone is $tz; loads assume UTC"
  echo "ok: $tz"
  step "other sessions on the MySQL users (must be 0 during the freeze)"
  # The instance is shared with other apps and these users can see every thread: filter by user.
  others="SELECT COUNT(*) FROM information_schema.processlist WHERE user = SUBSTRING_INDEX(CURRENT_USER(), '@', 1) AND id <> CONNECTION_ID()"
  a=$(my_app -e "$others"); o=$(my_oauth -e "$others")
  echo "app user: $a, oauth user: $o"
  if [[ -z $SUFFIX ]]; then [[ $a == 0 && $o == 0 ]] || die "writers still connected"; fi
  step "source checksums, 30s apart (must be identical during the freeze)"
  checksums > "$WORK/checksum-1.txt"; sleep 30; checksums > "$WORK/checksum-2.txt"
  grep -q NULL "$WORK/checksum-1.txt" && die "CHECKSUM TABLE returned NULL; the stability check would be meaningless"
  if diff -q "$WORK/checksum-1.txt" "$WORK/checksum-2.txt" >/dev/null; then echo "ok: stable"
  elif [[ -z $SUFFIX ]]; then die "source changed during the freeze"
  else echo "note: source changed (expected outside a freeze)"; fi
  ;;

prepare)
  step "target databases must be empty"
  n=$(pg_app -c "SELECT count(*) FROM pg_tables WHERE schemaname = 'public'")
  m=$(pg_slack -c "SELECT count(*) FROM pg_tables WHERE schemaname = 'public'")
  [[ $n == 0 && $m == 0 ]] || die "target not empty (app $n tables, slack $m tables)"
  step "app migrations"
  PASSWORDEXCHANGE_DBDRIVER=postgres PASSWORDEXCHANGE_DBHOST=127.0.0.1 PASSWORDEXCHANGE_DBPORT=$PG_PORT \
    PASSWORDEXCHANGE_DBUSER=passwordexchange PASSWORDEXCHANGE_DBPASS="$PG_APP_PASS" \
    PASSWORDEXCHANGE_DBNAME="passwordexchange$SUFFIX" "$APP_BIN" database migrate up 2>&1 | grep -iE 'version|error|dirty' || true
  src=$(my_app -e "SELECT CONCAT(version, '/', dirty) FROM schema_migrations")
  dst=$(pg_app -c "SELECT version || '/' || CASE WHEN dirty THEN 1 ELSE 0 END FROM schema_migrations")
  [[ $src == "$dst" ]] || die "migration version MySQL $src != Postgres $dst"
  echo "ok: both at $dst"
  step "slack-sdk tables"
  slack_env "$PY" -I "$HERE/slack_tables.py"
  ;;

load)
  [[ $(pg_app -c "SELECT count(*) FROM messages") == 0 ]] || die "messages already has rows; reload into a fresh database"
  schema_dump before
  step "temporary fileid column (MySQL-only, unused, must be all NULL)"
  pg_app -c "ALTER TABLE messages ADD COLUMN fileid varchar(500)"
  "$PY" -I "$HERE/loadfiles.py" "$WORK" --suffix "$SUFFIX"
  for f in 1-messages 2-app-rest 3-slack; do
    step "pgloader $f"
    log="$WORK/$f.log"
    podman run --rm --pod "$POD" --platform linux/amd64 -v "$WORK:/load:ro" "$PGLOADER_IMAGE" \
      pgloader "/load/$f.load" > "$log" 2>&1 || { sed -E 's#//[^@ ]*@#//***@#g' "$log" | tail -20; die "pgloader $f failed"; }
    sed -E 's#//[^@ ]*@#//***@#g' "$log" > "$log.tmp" && mv "$log.tmp" "$log"
    grep -E 'ERROR|FATAL|KABOOM' "$log" && die "pgloader $f reported errors"
    grep -E '^ *public\.|Total import time' "$log"
  done
  rm -f "$WORK"/*.load
  step "drop fileid"
  pg_app -c "DO \$\$ BEGIN IF EXISTS (SELECT 1 FROM messages WHERE fileid IS NOT NULL) THEN RAISE EXCEPTION 'fileid has data'; END IF; END \$\$" \
         -c "ALTER TABLE messages DROP COLUMN fileid"
  step "sequences"
  "$PY" -I "$HERE/sequences.py" --suffix "$SUFFIX"
  step "schema unchanged by the load"
  schema_dump after
  for db in passwordexchange slack_oauth; do
    diff -u "$WORK/schema-$db-before.sql" "$WORK/schema-$db-after.sql" || die "$db schema changed during the load"
  done
  echo "ok: schemas identical"
  ;;

verify)
  step "row-by-row verification"
  (cd "$HERE" && "$PY" -I verify.py --suffix "$SUFFIX")
  step "Slack store lookups"
  slack_env SLACK_CLIENT_ID="$SLACK_CLIENT_ID" "$PY" -I "$HERE/slack_lookup.py"
  ;;

reset)
  step "empty target tables (keeps the schema)"
  pg_app -c "ALTER TABLE messages DROP COLUMN IF EXISTS fileid" \
         -c "TRUNCATE messages, email_reminders, file_upload_sessions RESTART IDENTITY"
  pg_slack -c "TRUNCATE slack_installations, slack_bots RESTART IDENTITY"
  echo "ok: rerun the load stage"
  ;;

*) die "unknown stage $STAGE" ;;
esac
